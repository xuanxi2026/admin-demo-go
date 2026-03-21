package handler

import (
	"admin-demo-go/internal/pkg/ecode"
	"admin-demo-go/internal/pkg/response"
	"admin-demo-go/internal/service"

	"github.com/gin-gonic/gin"
)

type RBACHandler struct {
	rbacSvc *service.RBACService
}

func NewRBACHandler(rbacSvc *service.RBACService) *RBACHandler {
	return &RBACHandler{rbacSvc: rbacSvc}
}

func (h *RBACHandler) MyPermissions(c *gin.Context) {
	userID := c.GetUint("userID")
	perms, err := h.rbacSvc.PermissionsByUserID(userID)
	if err != nil {
		response.Fail(c, ecode.InternalError, "获取权限失败")
		return
	}
	response.OK(c, "success", perms)
}

func (h *RBACHandler) MyMenus(c *gin.Context) {
	userID := c.GetUint("userID")
	menus, err := h.rbacSvc.MenusByUserID(userID)
	if err != nil {
		response.Fail(c, ecode.InternalError, "获取菜单失败")
		return
	}
	response.OK(c, "success", menus)
}
