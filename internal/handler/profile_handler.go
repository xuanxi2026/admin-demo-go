package handler

import (
	"admin-demo-go/internal/pkg/ecode"
	"admin-demo-go/internal/pkg/event"
	"admin-demo-go/internal/pkg/response"
	"admin-demo-go/internal/service"

	"github.com/gin-gonic/gin"
)

type ProfileHandler struct {
	profileSvc *service.ProfileService
	rbacSvc    *service.RBACService
	pub        *event.Publisher
}

func NewProfileHandler(profileSvc *service.ProfileService, rbacSvc *service.RBACService, pub *event.Publisher) *ProfileHandler {
	return &ProfileHandler{profileSvc: profileSvc, rbacSvc: rbacSvc, pub: pub}
}

func (h *ProfileHandler) UserInfo(c *gin.Context) {
	userID := c.GetUint("userID")
	username := c.GetString("username")

	user, err := h.profileSvc.GetByID(userID)
	if err != nil {
		response.Fail(c, ecode.InternalError, "用户不存在")
		return
	}

	permissionCodes, err := h.rbacSvc.PermissionsByUserID(userID)
	if err != nil {
		response.Fail(c, ecode.InternalError, "查询权限失败")
		return
	}
	roleCodes, err := h.rbacSvc.RoleCodesByUserID(userID)
	if err != nil {
		response.Fail(c, ecode.InternalError, "查询角色失败")
		return
	}
	if len(roleCodes) == 0 {
		roleCodes = []string{"editor"}
	}

	response.OK(c, "success", gin.H{
		"username":        username,
		"avatar":          user.Avatar,
		"permissions":     roleCodes,
		"permissionCodes": permissionCodes,
		"role":            roleCodes[0],
		"googleBound":     user.GoogleSecret != "",
	})
}

func (h *ProfileHandler) GetProfile(c *gin.Context) {
	userID := c.GetUint("userID")
	user, err := h.profileSvc.GetByID(userID)
	if err != nil {
		response.Fail(c, ecode.InternalError, "用户不存在")
		return
	}
	response.OK(c, "success", gin.H{
		"username":    user.Username,
		"nickname":    user.Nickname,
		"phone":       user.Phone,
		"email":       user.Email,
		"avatar":      user.Avatar,
		"bio":         user.Bio,
		"googleBound": user.GoogleSecret != "",
	})
}

func (h *ProfileHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetUint("userID")
	var in service.ProfileUpdateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, ecode.InvalidParams, "参数错误")
		return
	}
	if err := h.profileSvc.Update(userID, in); err != nil {
		response.Fail(c, ecode.InternalError, "更新失败")
		return
	}
	h.pub.Publish("profile_update", map[string]any{
		"request_id": c.GetString("request_id"),
		"user_id":    userID,
	})
	response.OK(c, "更新成功", nil)
}
