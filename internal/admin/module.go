package admin

import (
	"admin-demo-go/internal/pkg/event"
	"admin-demo-go/internal/repository"
	"admin-demo-go/internal/storage"
)

type Module struct {
	userRepo      *repository.UserRepository
	rbacRepo      *repository.RBACRepository
	systemRepo    *repository.SystemRepository
	pub           *event.Publisher
	storage       storage.Client
	publicBaseURL string
}

func NewModule(userRepo *repository.UserRepository, rbacRepo *repository.RBACRepository, systemRepo *repository.SystemRepository, pub *event.Publisher, st storage.Client, publicBaseURL string) *Module {
	return &Module{
		userRepo:      userRepo,
		rbacRepo:      rbacRepo,
		systemRepo:    systemRepo,
		pub:           pub,
		storage:       st,
		publicBaseURL: publicBaseURL,
	}
}
