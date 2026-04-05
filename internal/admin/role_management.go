package admin

import (
	"strings"

	"admin-demo-go/internal/model"
	"admin-demo-go/internal/pkg/ecode"
	"admin-demo-go/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func (m *Module) RoleGetList(c *gin.Context) {
	var in struct {
		PageNo     int    `json:"pageNo"`
		PageSize   int    `json:"pageSize"`
		Permission string `json:"permission"`
	}
	_ = c.ShouldBindJSON(&in)
	in.PageNo, in.PageSize = normalizePage(in.PageNo, in.PageSize)
	list, total, err := m.rbacRepo.ListRoles(in.PageNo, in.PageSize, in.Permission)
	if err != nil {
		response.Fail(c, ecode.InternalError, "查询角色失败")
		return
	}
	rows := make([]gin.H, 0, len(list))
	for _, item := range list {
		rows = append(rows, gin.H{
			"id":         item.ID,
			"permission": item.Code,
		})
	}
	response.List(c, "success", rows, total)
}

func (m *Module) RoleDoEdit(c *gin.Context) {
	var in struct {
		ID         uint   `json:"id"`
		Permission string `json:"permission"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, ecode.InvalidParams, "参数错误")
		return
	}
	code := strings.TrimSpace(in.Permission)
	if code == "" {
		response.Fail(c, ecode.InvalidParams, "权限码不能为空")
		return
	}

	if in.ID > 0 {
		if err := m.rbacRepo.UpdateRoleByID(in.ID, map[string]any{"code": code, "name": code}); err != nil {
			response.Fail(c, ecode.InternalError, "更新失败")
			return
		}
	} else {
		if err := m.rbacRepo.CreateRole(&model.Role{Code: code, Name: code}); err != nil {
			response.Fail(c, ecode.InternalError, "创建失败")
			return
		}
	}
	m.pub.Publish("role_mgmt_edit", map[string]any{
		"request_id": c.GetString("request_id"),
		"operator":   c.GetString("username"),
		"permission": code,
	})
	m.recordOperation(operationContext{
		Module:    "role",
		Action:    "save",
		Operator:  c.GetString("username"),
		Target:    code,
		RequestID: c.GetString("request_id"),
		IP:        c.ClientIP(),
		Detail:    "保存角色",
	})
	response.OK(c, "保存成功", nil)
}

func (m *Module) RoleDoDelete(c *gin.Context) {
	var in struct {
		IDs string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, ecode.InvalidParams, "参数错误")
		return
	}
	ids := parseIDs(in.IDs)
	if err := m.rbacRepo.DeleteRolesByIDs(ids); err != nil {
		response.Fail(c, ecode.InternalError, "删除失败")
		return
	}
	m.pub.Publish("role_mgmt_delete", map[string]any{
		"request_id": c.GetString("request_id"),
		"operator":   c.GetString("username"),
		"ids":        in.IDs,
	})
	m.recordOperation(operationContext{
		Module:    "role",
		Action:    "delete",
		Operator:  c.GetString("username"),
		Target:    joinIDs(in.IDs),
		RequestID: c.GetString("request_id"),
		IP:        c.ClientIP(),
		Detail:    "删除角色",
	})
	response.OK(c, "删除成功", nil)
}
