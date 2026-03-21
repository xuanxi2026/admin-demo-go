package admin

import (
	"strings"

	"admin-demo-go/internal/model"
	"admin-demo-go/internal/pkg/ecode"

	"github.com/gin-gonic/gin"
)

func (m *Module) RoleGetList(c *gin.Context) {
	var in struct {
		PageNo     int    `json:"pageNo"`
		PageSize   int    `json:"pageSize"`
		Permission string `json:"permission"`
	}
	_ = c.ShouldBindJSON(&in)
	if in.PageNo <= 0 {
		in.PageNo = 1
	}
	if in.PageSize <= 0 {
		in.PageSize = 10
	}
	list, total, err := m.rbacRepo.ListRoles(in.PageNo, in.PageSize, in.Permission)
	if err != nil {
		c.JSON(200, gin.H{"code": ecode.InternalError, "msg": "查询角色失败", "request_id": c.GetString("request_id")})
		return
	}
	rows := make([]gin.H, 0, len(list))
	for _, item := range list {
		rows = append(rows, gin.H{
			"id":         item.ID,
			"permission": item.Code,
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

func (m *Module) RoleDoEdit(c *gin.Context) {
	var in struct {
		ID         uint   `json:"id"`
		Permission string `json:"permission"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(200, gin.H{"code": ecode.InvalidParams, "msg": "参数错误", "request_id": c.GetString("request_id")})
		return
	}
	code := strings.TrimSpace(in.Permission)
	if code == "" {
		c.JSON(200, gin.H{"code": ecode.InvalidParams, "msg": "权限码不能为空", "request_id": c.GetString("request_id")})
		return
	}

	if in.ID > 0 {
		if err := m.rbacRepo.UpdateRoleByID(in.ID, map[string]any{"code": code, "name": code}); err != nil {
			c.JSON(200, gin.H{"code": ecode.InternalError, "msg": "更新失败", "request_id": c.GetString("request_id")})
			return
		}
	} else {
		if err := m.rbacRepo.CreateRole(&model.Role{Code: code, Name: code}); err != nil {
			c.JSON(200, gin.H{"code": ecode.InternalError, "msg": "创建失败", "request_id": c.GetString("request_id")})
			return
		}
	}
	m.pub.Publish("role_mgmt_edit", map[string]any{
		"request_id": c.GetString("request_id"),
		"operator":   c.GetString("username"),
		"permission": code,
	})
	c.JSON(200, gin.H{"code": 200, "msg": "保存成功", "request_id": c.GetString("request_id")})
}

func (m *Module) RoleDoDelete(c *gin.Context) {
	var in struct {
		IDs string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(200, gin.H{"code": ecode.InvalidParams, "msg": "参数错误", "request_id": c.GetString("request_id")})
		return
	}
	ids := parseIDs(in.IDs)
	if err := m.rbacRepo.DeleteRolesByIDs(ids); err != nil {
		c.JSON(200, gin.H{"code": ecode.InternalError, "msg": "删除失败", "request_id": c.GetString("request_id")})
		return
	}
	m.pub.Publish("role_mgmt_delete", map[string]any{
		"request_id": c.GetString("request_id"),
		"operator":   c.GetString("username"),
		"ids":        in.IDs,
	})
	c.JSON(200, gin.H{"code": 200, "msg": "删除成功", "request_id": c.GetString("request_id")})
}
