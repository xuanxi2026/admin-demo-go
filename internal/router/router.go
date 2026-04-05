package router

import (
	"admin-demo-go/internal/admin"
	"admin-demo-go/internal/bootstrap"
	"admin-demo-go/internal/handler"
	"admin-demo-go/internal/middleware"
	"admin-demo-go/internal/repository"
	"admin-demo-go/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func New(app *bootstrap.App) *gin.Engine {
	gin.SetMode(app.Cfg.App.Mode)
	r := gin.New()
	r.Use(gin.Recovery(), middleware.RequestID(), middleware.AccessLog(app.EventPub), cors.New(middleware.CORS()))

	userRepo := repository.NewUserRepository(app.DB)
	rbacRepo := repository.NewRBACRepository(app.DB)
	systemRepo := repository.NewSystemRepository(app.DB)
	authSvc := service.NewAuthService(userRepo, rbacRepo, app.Cfg.App.JWTSecret, app.Cfg.App.JWTExpireHours)
	rbacSvc := service.NewRBACService(rbacRepo)
	googleAuthSvc := service.NewGoogleAuthService(userRepo, app.Redis, app.Cfg.App.Name)
	profileSvc := service.NewProfileService(userRepo)
	adminModule := admin.NewModule(userRepo, rbacRepo, systemRepo, app.EventPub, app.Storage, app.Cfg.Storage.Public.BaseURL)
	authHandler := handler.NewAuthHandler(authSvc, app.EventPub)
	profileHandler := handler.NewProfileHandler(profileSvc, rbacSvc, app.EventPub)
	menuHandler := handler.NewMenuHandler(rbacSvc)
	googleAuthHandler := handler.NewGoogleAuthHandler(userRepo, googleAuthSvc, app.EventPub)
	rbacHandler := handler.NewRBACHandler(rbacSvc)

	api := r.Group("/api")
	{
		api.POST("/login", authHandler.Login)
		api.POST("/register", authHandler.Register)
	}

	protected := api.Group("")
	protected.Use(middleware.Auth(app.Cfg.App.JWTSecret))
	{
		protected.POST("/logout", authHandler.Logout)
		protected.POST("/userInfo", profileHandler.UserInfo)
		protected.GET("/profile", profileHandler.GetProfile)
		protected.PUT("/profile", middleware.RequirePermission(rbacSvc, "profile:update"), profileHandler.UpdateProfile)
		protected.POST("/menu/navigate", menuHandler.Navigate)

		protected.GET("/auth/google/setup", middleware.RequirePermission(rbacSvc, "google:bind"), googleAuthHandler.Setup)
		protected.POST("/auth/google/bind", middleware.RequirePermission(rbacSvc, "google:bind"), googleAuthHandler.Bind)
		protected.POST("/auth/google/unbind", middleware.RequirePermission(rbacSvc, "google:bind"), googleAuthHandler.Unbind)

		protected.GET("/rbac/permissions", middleware.RequirePermission(rbacSvc, "rbac:view"), rbacHandler.MyPermissions)
		protected.GET("/rbac/menus", middleware.RequirePermission(rbacSvc, "rbac:view"), rbacHandler.MyMenus)

		protected.POST("/userManagement/getList", middleware.RequirePermission(rbacSvc, "rbac:view"), adminModule.UserGetList)
		protected.POST("/userManagement/doEdit", middleware.RequirePermission(rbacSvc, "rbac:view"), adminModule.UserDoEdit)
		protected.POST("/userManagement/doDelete", middleware.RequirePermission(rbacSvc, "rbac:view"), adminModule.UserDoDelete)

		protected.POST("/roleManagement/getList", middleware.RequirePermission(rbacSvc, "rbac:view"), adminModule.RoleGetList)
		protected.POST("/roleManagement/doEdit", middleware.RequirePermission(rbacSvc, "rbac:view"), adminModule.RoleDoEdit)
		protected.POST("/roleManagement/doDelete", middleware.RequirePermission(rbacSvc, "rbac:view"), adminModule.RoleDoDelete)

		protected.POST("/menuManagement/getTree", middleware.RequirePermission(rbacSvc, "rbac:view"), adminModule.MenuGetTree)
		protected.POST("/menuManagement/doEdit", middleware.RequirePermission(rbacSvc, "rbac:view"), adminModule.MenuDoEdit)
		protected.POST("/menuManagement/doDelete", middleware.RequirePermission(rbacSvc, "rbac:view"), adminModule.MenuDoDelete)
		protected.POST("/dictManagement/getList", middleware.RequirePermission(rbacSvc, "rbac:view"), adminModule.DictGetList)
		protected.POST("/dictManagement/doEdit", middleware.RequirePermission(rbacSvc, "rbac:view"), adminModule.DictDoEdit)
		protected.POST("/dictManagement/doDelete", middleware.RequirePermission(rbacSvc, "rbac:view"), adminModule.DictDoDelete)
		protected.POST("/configManagement/getList", middleware.RequirePermission(rbacSvc, "rbac:view"), adminModule.ConfigGetList)
		protected.POST("/configManagement/doEdit", middleware.RequirePermission(rbacSvc, "rbac:view"), adminModule.ConfigDoEdit)
		protected.POST("/configManagement/doDelete", middleware.RequirePermission(rbacSvc, "rbac:view"), adminModule.ConfigDoDelete)
		protected.POST("/operationLog/getList", middleware.RequirePermission(rbacSvc, "rbac:view"), adminModule.OperationLogGetList)

		protected.POST("/fileManagement/public/list", adminModule.FilePublicList)
		protected.POST("/fileManagement/private/list", adminModule.FilePrivateList)
		protected.POST("/fileManagement/public/upload", adminModule.FilePublicUpload)
		protected.POST("/fileManagement/private/upload", adminModule.FilePrivateUpload)
		protected.POST("/fileManagement/public/delete", adminModule.FilePublicDelete)
		protected.POST("/fileManagement/private/delete", adminModule.FilePrivateDelete)
		protected.GET("/fileManagement/private/download/:name", adminModule.FilePrivateDownload)
	}

	r.GET("/files/public/:name", adminModule.FilePublicDownload)

	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	return r
}
