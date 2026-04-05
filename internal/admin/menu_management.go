package admin

import (
	"strings"

	"admin-demo-go/internal/pkg/ecode"
	"admin-demo-go/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func (m *Module) MenuGetTree(c *gin.Context) {
	roles, err := m.rbacRepo.ListAllRoles()
	if err != nil {
		response.Fail(c, ecode.InternalError, "查询角色树失败")
		return
	}
	children := make([]gin.H, 0, len(roles))
	for _, role := range roles {
		children = append(children, gin.H{
			"id":         role.ID,
			"permission": role.Code,
			"label":      role.Code + "角色",
		})
	}
	response.List(c, "success", []gin.H{
		{
			"id":       "root",
			"label":    "全部角色",
			"children": children,
		},
	}, int64(len(children)))
}

func (m *Module) MenuDoEdit(c *gin.Context) {
	var in struct {
		ID         uint   `json:"id"`
		ParentID   uint   `json:"parentId"`
		Name       string `json:"name"`
		Path       string `json:"path"`
		Component  string `json:"component"`
		Redirect   string `json:"redirect"`
		Hidden     bool   `json:"hidden"`
		AlwaysShow bool   `json:"alwaysShow"`
		Meta       struct {
			Title       string `json:"title"`
			Icon        string `json:"icon"`
			Badge       string `json:"badge"`
			Affix       bool   `json:"affix"`
			NoKeepAlive bool   `json:"noKeepAlive"`
		} `json:"meta"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, ecode.InvalidParams, "参数错误")
		return
	}
	name := strings.TrimSpace(in.Name)
	path := strings.TrimSpace(in.Path)
	if name == "" || path == "" {
		response.Fail(c, ecode.InvalidParams, "name与path不能为空")
		return
	}
	title := strings.TrimSpace(in.Meta.Title)
	if title == "" {
		title = name
	}
	updates := map[string]any{
		"parent_id":     in.ParentID,
		"name":          name,
		"path":          path,
		"component":     in.Component,
		"redirect":      in.Redirect,
		"title":         title,
		"icon":          in.Meta.Icon,
		"badge":         in.Meta.Badge,
		"affix":         in.Meta.Affix,
		"no_keep_alive": in.Meta.NoKeepAlive,
		"hidden":        in.Hidden,
		"always_show":   in.AlwaysShow,
	}
	if in.ID > 0 {
		if err := m.rbacRepo.UpdateMenuByID(in.ID, updates); err != nil {
			response.Fail(c, ecode.InternalError, "更新菜单失败")
			return
		}
	} else {
		menu := map[string]any{
			"parent_id":     in.ParentID,
			"name":          name,
			"path":          path,
			"component":     in.Component,
			"redirect":      in.Redirect,
			"title":         title,
			"icon":          in.Meta.Icon,
			"badge":         in.Meta.Badge,
			"affix":         in.Meta.Affix,
			"no_keep_alive": in.Meta.NoKeepAlive,
			"hidden":        in.Hidden,
			"always_show":   in.AlwaysShow,
			"sort":          999,
		}
		if err := m.rbacRepo.CreateMenuMap(menu); err != nil {
			response.Fail(c, ecode.InternalError, "新增菜单失败")
			return
		}
	}
	m.pub.Publish("menu_mgmt_edit", map[string]any{
		"request_id": c.GetString("request_id"),
		"operator":   c.GetString("username"),
		"name":       name,
		"path":       path,
	})
	m.recordOperation(operationContext{
		Module:    "menu",
		Action:    "save",
		Operator:  c.GetString("username"),
		Target:    name,
		RequestID: c.GetString("request_id"),
		IP:        c.ClientIP(),
		Detail:    buildDetail("path="+path, "title="+title),
	})
	response.OK(c, "保存成功", nil)
}

func (m *Module) MenuDoDelete(c *gin.Context) {
	var in struct {
		IDs string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, ecode.InvalidParams, "参数错误")
		return
	}
	ids := parseIDs(in.IDs)
	if err := m.rbacRepo.DeleteMenusByIDs(ids); err != nil {
		response.Fail(c, ecode.InternalError, "删除菜单失败")
		return
	}
	m.pub.Publish("menu_mgmt_delete", map[string]any{
		"request_id": c.GetString("request_id"),
		"operator":   c.GetString("username"),
		"ids":        in.IDs,
	})
	m.recordOperation(operationContext{
		Module:    "menu",
		Action:    "delete",
		Operator:  c.GetString("username"),
		Target:    joinIDs(in.IDs),
		RequestID: c.GetString("request_id"),
		IP:        c.ClientIP(),
		Detail:    "删除菜单",
	})
	response.OK(c, "删除成功", nil)
}
