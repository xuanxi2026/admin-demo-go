package admin

import (
	"strings"
	"time"

	"admin-demo-go/internal/model"
	"admin-demo-go/internal/pkg/ecode"
	"admin-demo-go/internal/pkg/password"
	"admin-demo-go/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func (m *Module) UserGetList(c *gin.Context) {
	var in struct {
		PageNo   int    `json:"pageNo"`
		PageSize int    `json:"pageSize"`
		Username string `json:"username"`
	}
	_ = c.ShouldBindJSON(&in)
	in.PageNo, in.PageSize = normalizePage(in.PageNo, in.PageSize)
	list, total, err := m.userRepo.List(in.PageNo, in.PageSize, in.Username)
	if err != nil {
		response.Fail(c, ecode.InternalError, "查询用户失败")
		return
	}
	rows := make([]gin.H, 0, len(list))
	for _, u := range list {
		roles, _ := m.rbacRepo.GetRoleCodesByUserID(u.ID)
		rows = append(rows, gin.H{
			"id":          u.ID,
			"username":    u.Username,
			"email":       u.Email,
			"permissions": roles,
			"datatime":    u.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	response.List(c, "success", rows, total)
}

func (m *Module) UserDoEdit(c *gin.Context) {
	var in struct {
		ID          uint     `json:"id"`
		Username    string   `json:"username"`
		Password    string   `json:"password"`
		Email       string   `json:"email"`
		Permissions []string `json:"permissions"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, ecode.InvalidParams, "参数错误")
		return
	}
	if in.Username == "" {
		response.Fail(c, ecode.InvalidParams, "用户名不能为空")
		return
	}
	if len(in.Permissions) == 0 {
		in.Permissions = []string{"editor"}
	}
	mainRole := in.Permissions[0]
	if mainRole != "admin" && mainRole != "editor" && mainRole != "test" {
		mainRole = "editor"
	}

	if in.ID > 0 {
		updates := map[string]any{
			"username": in.Username,
			"email":    in.Email,
			"role":     mainRole,
		}
		if in.Password != "" {
			hash, err := password.Hash(in.Password)
			if err != nil {
				response.Fail(c, ecode.InternalError, "密码加密失败")
				return
			}
			updates["password_hash"] = hash
		}
		if err := m.userRepo.UpdateByID(in.ID, updates); err != nil {
			response.Fail(c, ecode.InternalError, "更新失败")
			return
		}
		_ = m.rbacRepo.ReplaceUserRoles(in.ID, in.Permissions)
	} else {
		if in.Password == "" {
			response.Fail(c, ecode.InvalidParams, "密码不能为空")
			return
		}
		hash, err := password.Hash(in.Password)
		if err != nil {
			response.Fail(c, ecode.InternalError, "密码加密失败")
			return
		}
		user := &model.User{
			Username:     in.Username,
			PasswordHash: hash,
			Email:        in.Email,
			Role:         mainRole,
			Nickname:     in.Username,
			Avatar:       "https://gcore.jsdelivr.net/gh/zxwk1998/image/avatar/avatar_1.png",
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}
		if err = m.userRepo.Create(user); err != nil {
			response.Fail(c, ecode.InternalError, "创建失败")
			return
		}
		_ = m.rbacRepo.ReplaceUserRoles(user.ID, in.Permissions)
	}
	m.pub.Publish("user_mgmt_edit", map[string]any{
		"request_id": c.GetString("request_id"),
		"operator":   c.GetString("username"),
		"target":     in.Username,
	})
	m.recordOperation(operationContext{
		Module:    "user",
		Action:    "save",
		Operator:  c.GetString("username"),
		Target:    in.Username,
		RequestID: c.GetString("request_id"),
		IP:        c.ClientIP(),
		Detail:    buildDetail("email="+in.Email, "roles="+strings.Join(in.Permissions, ",")),
	})
	response.OK(c, "保存成功", nil)
}

func (m *Module) UserDoDelete(c *gin.Context) {
	var in struct {
		IDs string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, ecode.InvalidParams, "参数错误")
		return
	}
	ids := parseIDs(in.IDs)
	if err := m.userRepo.DeleteByIDs(ids); err != nil {
		response.Fail(c, ecode.InternalError, "删除失败")
		return
	}
	m.pub.Publish("user_mgmt_delete", map[string]any{
		"request_id": c.GetString("request_id"),
		"operator":   c.GetString("username"),
		"ids":        in.IDs,
	})
	m.recordOperation(operationContext{
		Module:    "user",
		Action:    "delete",
		Operator:  c.GetString("username"),
		Target:    joinIDs(in.IDs),
		RequestID: c.GetString("request_id"),
		IP:        c.ClientIP(),
		Detail:    "删除用户",
	})
	response.OK(c, "删除成功", nil)
}
