package service

import (
	"admin-demo-go/internal/model"
	"admin-demo-go/internal/repository"
)

type RBACService struct {
	repo *repository.RBACRepository
}

func NewRBACService(repo *repository.RBACRepository) *RBACService {
	return &RBACService{repo: repo}
}

func (s *RBACService) PermissionsByUserID(userID uint) ([]string, error) {
	return s.repo.GetPermissionsByUserID(userID)
}

func (s *RBACService) HasPermission(userID uint, code string) (bool, error) {
	return s.repo.UserHasPermission(userID, code)
}

func (s *RBACService) MenusByUserID(userID uint) ([]model.Menu, error) {
	return s.repo.GetMenusByUserID(userID)
}

func (s *RBACService) RoleCodesByUserID(userID uint) ([]string, error) {
	return s.repo.GetRoleCodesByUserID(userID)
}
