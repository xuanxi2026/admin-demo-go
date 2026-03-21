package service

import (
	"log"
	"strings"

	"admin-demo-go/internal/model"
	"admin-demo-go/internal/pkg/apperr"
	"admin-demo-go/internal/pkg/ecode"
	"admin-demo-go/internal/pkg/jwt"
	"admin-demo-go/internal/pkg/password"
	"admin-demo-go/internal/repository"
)

type AuthService struct {
	repo       *repository.UserRepository
	rbacRepo   *repository.RBACRepository
	jwtSecret  string
	jwtExpires int
}

func NewAuthService(repo *repository.UserRepository, rbacRepo *repository.RBACRepository, jwtSecret string, jwtExpires int) *AuthService {
	return &AuthService{
		repo:       repo,
		rbacRepo:   rbacRepo,
		jwtSecret:  jwtSecret,
		jwtExpires: jwtExpires,
	}
}

type RegisterInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Nickname string `json:"nickname"`
}

func (s *AuthService) Register(in RegisterInput) error {
	if in.Username == "" || in.Password == "" {
		return apperr.New(ecode.InvalidParams, "用户名和密码不能为空")
	}
	if _, err := s.repo.FindByUsername(in.Username); err == nil {
		return apperr.New(ecode.UserExists, "用户名已存在")
	}

	hash, err := password.Hash(in.Password)
	if err != nil {
		return err
	}
	user := &model.User{
		Username:     in.Username,
		PasswordHash: hash,
		Phone:        in.Phone,
		Email:        in.Email,
		Nickname:     in.Nickname,
		Role:         "editor",
		Avatar:       "https://gcore.jsdelivr.net/gh/zxwk1998/image/avatar/avatar_1.png",
	}
	if err = s.repo.Create(user); err != nil {
		return apperr.New(ecode.InternalError, "创建用户失败")
	}
	role, err := s.rbacRepo.FindRoleByCode("editor")
	if err != nil {
		return apperr.New(ecode.InternalError, "默认角色不存在")
	}
	if err = s.rbacRepo.BindUserRole(user.ID, role.ID); err != nil {
		return apperr.New(ecode.InternalError, "绑定默认角色失败")
	}
	return nil
}

type LoginInput struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	GoogleCode string `json:"googleCode"`
}

func (s *AuthService) Login(reqID string, in LoginInput) (string, *model.User, bool, error) {
	log.Printf("request_id=%s stage=login_query_user username=%s", reqID, in.Username)
	user, err := s.repo.FindByUsername(in.Username)
	if err != nil {
		log.Printf("request_id=%s stage=login_query_user status=failed username=%s err=%v", reqID, in.Username, err)
		return "", nil, false, apperr.New(ecode.LoginFailed, "账号或密码错误")
	}
	log.Printf("request_id=%s stage=login_verify_password user_id=%d", reqID, user.ID)
	passwordHash := strings.TrimSpace(user.PasswordHash)
	if !password.Verify(passwordHash, in.Password) {
		log.Printf("request_id=%s stage=login_verify_password status=failed user_id=%d hash_len=%d", reqID, user.ID, len(passwordHash))
		return "", nil, false, apperr.New(ecode.LoginFailed, "账号或密码错误")
	}
	log.Printf("request_id=%s stage=login_verify_password status=success user_id=%d", reqID, user.ID)
	googleBound := user.GoogleSecret != ""
	if googleBound && in.GoogleCode == "" {
		log.Printf("request_id=%s stage=login_verify_google status=failed reason=missing_code user_id=%d", reqID, user.ID)
		return "", nil, true, apperr.New(ecode.GoogleCodeFailed, "请输入谷歌验证码")
	}
	if googleBound {
		// 仅已绑定用户需要校验
		if ok := ValidateTOTP(user.GoogleSecret, in.GoogleCode); !ok {
			log.Printf("request_id=%s stage=login_verify_google status=failed reason=invalid_code user_id=%d", reqID, user.ID)
			return "", nil, true, apperr.New(ecode.GoogleCodeFailed, "谷歌验证码错误")
		}
		log.Printf("request_id=%s stage=login_verify_google status=success user_id=%d", reqID, user.ID)
	}

	log.Printf("request_id=%s stage=login_query_roles user_id=%d", reqID, user.ID)
	roleCodes, err := s.rbacRepo.GetRoleCodesByUserID(user.ID)
	if err != nil {
		log.Printf("request_id=%s stage=login_query_roles status=failed user_id=%d err=%v", reqID, user.ID, err)
		return "", nil, googleBound, apperr.New(ecode.InternalError, "查询角色失败")
	}
	role := "editor"
	if len(roleCodes) > 0 {
		role = roleCodes[0]
	}
	log.Printf("request_id=%s stage=login_sign_jwt user_id=%d role=%s", reqID, user.ID, role)

	token, err := jwt.Sign(s.jwtSecret, s.jwtExpires, user.ID, user.Username, role)
	if err != nil {
		log.Printf("request_id=%s stage=login_sign_jwt status=failed user_id=%d err=%v", reqID, user.ID, err)
		return "", nil, googleBound, apperr.New(ecode.InternalError, "签发令牌失败")
	}
	log.Printf("request_id=%s stage=login_sign_jwt status=success user_id=%d", reqID, user.ID)
	return token, user, googleBound, nil
}
