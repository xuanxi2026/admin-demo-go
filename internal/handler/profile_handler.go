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
	roleCodes, err := h.rbacSvc.RoleCodesByUserID(userID)
	if err != nil {
		response.Fail(c, ecode.InternalError, "查询角色失败")
		return
	}
	permissionCodes, err := h.rbacSvc.PermissionsByUserID(userID)
	if err != nil {
		response.Fail(c, ecode.InternalError, "查询权限失败")
		return
	}
	response.OK(c, "success", gin.H{
		"username":        user.Username,
		"nickname":        user.Nickname,
		"phone":           user.Phone,
		"email":           user.Email,
		"avatar":          user.Avatar,
		"bio":             user.Bio,
		"googleBound":     user.GoogleSecret != "",
		"roles":           roleCodes,
		"permissionCodes": permissionCodes,
		"createdAt":       user.CreatedAt,
		"updatedAt":       user.UpdatedAt,
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

func (h *ProfileHandler) ChangePassword(c *gin.Context) {
	userID := c.GetUint("userID")
	var in service.ChangePasswordInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, ecode.InvalidParams, "参数错误")
		return
	}
	if len(in.CurrentPassword) < 6 || len(in.NewPassword) < 6 {
		response.Fail(c, ecode.InvalidParams, "密码长度至少为 6 位")
		return
	}
	if in.CurrentPassword == in.NewPassword {
		response.Fail(c, ecode.InvalidParams, "新密码不能与原密码相同")
		return
	}
	if err := h.profileSvc.ChangePassword(userID, in); err != nil {
		if err == service.ErrCurrentPasswordMismatch {
			response.Fail(c, ecode.InvalidParams, "原密码错误")
			return
		}
		response.Fail(c, ecode.InternalError, "修改密码失败")
		return
	}
	h.pub.Publish("profile_change_password", map[string]any{
		"request_id": c.GetString("request_id"),
		"user_id":    userID,
	})
	response.OK(c, "密码修改成功", nil)
}
