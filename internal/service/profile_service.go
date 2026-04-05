package service

import (
	"admin-demo-go/internal/model"
	"admin-demo-go/internal/pkg/password"
	"admin-demo-go/internal/repository"
	"errors"
)

type ProfileService struct {
	repo *repository.UserRepository
}

func NewProfileService(repo *repository.UserRepository) *ProfileService {
	return &ProfileService{repo: repo}
}

func (s *ProfileService) GetByID(userID uint) (*model.User, error) {
	return s.repo.FindByID(userID)
}

type ProfileUpdateInput struct {
	Nickname string `json:"nickname"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
	Bio      string `json:"bio"`
}

func (s *ProfileService) Update(userID uint, in ProfileUpdateInput) error {
	return s.repo.UpdateProfile(userID, map[string]any{
		"nickname": in.Nickname,
		"phone":    in.Phone,
		"email":    in.Email,
		"avatar":   in.Avatar,
		"bio":      in.Bio,
	})
}

var ErrCurrentPasswordMismatch = errors.New("current password mismatch")

type ChangePasswordInput struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

func (s *ProfileService) ChangePassword(userID uint, in ChangePasswordInput) error {
	user, err := s.repo.FindByID(userID)
	if err != nil {
		return err
	}
	if !password.Verify(user.PasswordHash, in.CurrentPassword) {
		return ErrCurrentPasswordMismatch
	}
	hash, err := password.Hash(in.NewPassword)
	if err != nil {
		return err
	}
	return s.repo.UpdateByID(userID, map[string]any{
		"password_hash": hash,
	})
}
