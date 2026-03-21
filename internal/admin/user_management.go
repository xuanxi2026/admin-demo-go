package admin

import (
	"time"

	"admin-demo-go/internal/model"
	"admin-demo-go/internal/pkg/ecode"
	"admin-demo-go/internal/pkg/password"

	"github.com/gin-gonic/gin"
)

func (m *Module) UserGetList(c *gin.Context) {
	var in struct {
		PageNo   int    `json:"pageNo"`
		PageSize int    `json:"pageSize"`
		Username string `json:"username"`
	}
	_ = c.ShouldBindJSON(&in)
	if in.PageNo <= 0 {
		in.PageNo = 1
	}
	if in.PageSize <= 0 {
		in.PageSize = 10
	}
	list, total, err := m.userRepo.List(in.PageNo, in.PageSize, in.Username)
	if err != nil {
		c.JSON(200, gin.H{"code": ecode.InternalError, "msg": "查询用户失败", "request_id": c.GetString("request_id")})
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
	c.JSON(200, gin.H{
		"code":       200,
		"msg":        "success",
		"data":       rows,
		"totalCount": total,
		"request_id": c.GetString("request_id"),
	})
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
		c.JSON(200, gin.H{"code": ecode.InvalidParams, "msg": "参数错误", "request_id": c.GetString("request_id")})
		return
	}
	if in.Username == "" {
		c.JSON(200, gin.H{"code": ecode.InvalidParams, "msg": "用户名不能为空", "request_id": c.GetString("request_id")})
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
				c.JSON(200, gin.H{"code": ecode.InternalError, "msg": "密码加密失败", "request_id": c.GetString("request_id")})
				return
			}
			updates["password_hash"] = hash
		}
		if err := m.userRepo.UpdateByID(in.ID, updates); err != nil {
			c.JSON(200, gin.H{"code": ecode.InternalError, "msg": "更新失败", "request_id": c.GetString("request_id")})
			return
		}
		_ = m.rbacRepo.ReplaceUserRoles(in.ID, in.Permissions)
	} else {
		if in.Password == "" {
			c.JSON(200, gin.H{"code": ecode.InvalidParams, "msg": "密码不能为空", "request_id": c.GetString("request_id")})
			return
		}
		hash, err := password.Hash(in.Password)
		if err != nil {
			c.JSON(200, gin.H{"code": ecode.InternalError, "msg": "密码加密失败", "request_id": c.GetString("request_id")})
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
			c.JSON(200, gin.H{"code": ecode.InternalError, "msg": "创建失败", "request_id": c.GetString("request_id")})
			return
		}
		_ = m.rbacRepo.ReplaceUserRoles(user.ID, in.Permissions)
	}
	m.pub.Publish("user_mgmt_edit", map[string]any{
		"request_id": c.GetString("request_id"),
		"operator":   c.GetString("username"),
		"target":     in.Username,
	})
	c.JSON(200, gin.H{"code": 200, "msg": "保存成功", "request_id": c.GetString("request_id")})
}

func (m *Module) UserDoDelete(c *gin.Context) {
	var in struct {
		IDs string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(200, gin.H{"code": ecode.InvalidParams, "msg": "参数错误", "request_id": c.GetString("request_id")})
		return
	}
	ids := parseIDs(in.IDs)
	if err := m.userRepo.DeleteByIDs(ids); err != nil {
		c.JSON(200, gin.H{"code": ecode.InternalError, "msg": "删除失败", "request_id": c.GetString("request_id")})
		return
	}
	m.pub.Publish("user_mgmt_delete", map[string]any{
		"request_id": c.GetString("request_id"),
		"operator":   c.GetString("username"),
		"ids":        in.IDs,
	})
	c.JSON(200, gin.H{"code": 200, "msg": "删除成功", "request_id": c.GetString("request_id")})
}
