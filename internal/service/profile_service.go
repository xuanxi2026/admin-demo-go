package service

import (
	"admin-demo-go/internal/model"
	"admin-demo-go/internal/repository"
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
