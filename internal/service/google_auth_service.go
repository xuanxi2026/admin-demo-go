package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"image/png"
	"sync"
	"time"

	"admin-demo-go/internal/model"
	"admin-demo-go/internal/repository"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	"github.com/redis/go-redis/v9"
)

type GoogleAuthService struct {
	userRepo       *repository.UserRepository
	redis          *redis.Client
	appName        string
	pendingSecrets sync.Map
}

func NewGoogleAuthService(userRepo *repository.UserRepository, redisClient *redis.Client, appName string) *GoogleAuthService {
	return &GoogleAuthService{
		userRepo: userRepo,
		redis:    redisClient,
		appName:  appName,
	}
}

type GoogleBindSetup struct {
	Secret     string `json:"secret"`
	OtpAuthURL string `json:"otpAuthUrl"`
	QRCodeB64  string `json:"qrCodeBase64"`
}

func (s *GoogleAuthService) Setup(user *model.User) (*GoogleBindSetup, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.appName,
		AccountName: user.Username,
		Algorithm:   otp.AlgorithmSHA1,
		Digits:      otp.DigitsSix,
		Period:      30,
	})
	if err != nil {
		return nil, err
	}
	if err = s.savePendingSecret(user.ID, key.Secret()); err != nil {
		return nil, err
	}

	img, err := key.Image(256, 256)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err = png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return &GoogleBindSetup{
		Secret:     key.Secret(),
		OtpAuthURL: key.URL(),
		QRCodeB64:  "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes()),
	}, nil
}

func (s *GoogleAuthService) ConfirmBind(userID uint, code string) error {
	secret, err := s.getPendingSecret(userID)
	if err != nil {
		return err
	}
	if !totp.Validate(code, secret) {
		return errors.New("验证码错误")
	}
	if err = s.userRepo.UpdateGoogleSecret(userID, secret); err != nil {
		return err
	}
	_ = s.clearPendingSecret(userID)
	return nil
}

func (s *GoogleAuthService) Unbind(userID uint, code string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return err
	}
	if user.GoogleSecret == "" {
		return nil
	}
	if !totp.Validate(code, user.GoogleSecret) {
		return errors.New("验证码错误")
	}
	return s.userRepo.UpdateGoogleSecret(userID, "")
}

func (s *GoogleAuthService) savePendingSecret(userID uint, secret string) error {
	if s.redis == nil {
		s.pendingSecrets.Store(userID, secret)
		return nil
	}
	key := s.pendingSecretKey(userID)
	return s.redis.Set(context.Background(), key, secret, 10*time.Minute).Err()
}

func (s *GoogleAuthService) getPendingSecret(userID uint) (string, error) {
	if s.redis == nil {
		if value, ok := s.pendingSecrets.Load(userID); ok {
			if secret, ok2 := value.(string); ok2 && secret != "" {
				return secret, nil
			}
		}
		return "", errors.New("请先生成二维码")
	}
	key := s.pendingSecretKey(userID)
	secret, err := s.redis.Get(context.Background(), key).Result()
	if err != nil {
		return "", errors.New("二维码已过期，请重新生成")
	}
	return secret, nil
}

func (s *GoogleAuthService) clearPendingSecret(userID uint) error {
	if s.redis == nil {
		s.pendingSecrets.Delete(userID)
		return nil
	}
	return s.redis.Del(context.Background(), s.pendingSecretKey(userID)).Err()
}

func (s *GoogleAuthService) pendingSecretKey(userID uint) string {
	return fmt.Sprintf("ga:pending:%d", userID)
}
