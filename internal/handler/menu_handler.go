package handler

import (
	"admin-demo-go/internal/model"
	"admin-demo-go/internal/pkg/ecode"
	"admin-demo-go/internal/pkg/response"
	"admin-demo-go/internal/service"

	"github.com/gin-gonic/gin"
)

type MenuHandler struct {
	rbacSvc *service.RBACService
}

func NewMenuHandler(rbacSvc *service.RBACService) *MenuHandler {
	return &MenuHandler{rbacSvc: rbacSvc}
}

func (h *MenuHandler) Navigate(c *gin.Context) {
	userID := c.GetUint("userID")
	menus, err := h.rbacSvc.MenusByUserID(userID)
	if err != nil {
		response.Fail(c, ecode.InternalError, "查询菜单失败")
		return
	}
	roleCodes, _ := h.rbacSvc.RoleCodesByUserID(userID)
	response.OK(c, "success", buildMenuTree(menus, roleCodes))
}

func buildMenuTree(menus []model.Menu, roleCodes []string) []gin.H {
	nodeMap := map[uint]gin.H{}
	for _, m := range menus {
		meta := gin.H{"title": m.Title}
		if m.Icon != "" {
			meta["icon"] = m.Icon
		}
		if m.Badge != "" {
			meta["badge"] = m.Badge
		}
		if m.Affix {
			meta["affix"] = true
		}
		if m.NoKeepAlive {
			meta["noKeepAlive"] = true
		}
		if m.PermissionCode != "" {
			meta["permissions"] = roleCodes
		}
		nodeMap[m.ID] = gin.H{
			"id":         m.ID,
			"path":       m.Path,
			"name":       m.Name,
			"component":  m.Component,
			"redirect":   m.Redirect,
			"alwaysShow": m.AlwaysShow,
			"hidden":     m.Hidden,
			"meta":       meta,
			"children":   []gin.H{},
		}
	}
	roots := make([]gin.H, 0)
	for _, m := range menus {
		current := nodeMap[m.ID]
		if m.ParentID == 0 {
			roots = append(roots, current)
			continue
		}
		if parent, ok := nodeMap[m.ParentID]; ok {
			parent["children"] = append(parent["children"].([]gin.H), current)
		}
	}
	return roots
}
