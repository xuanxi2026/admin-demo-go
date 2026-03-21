package admin

import (
	"strings"

	"admin-demo-go/internal/pkg/ecode"

	"github.com/gin-gonic/gin"
)

func (m *Module) MenuGetTree(c *gin.Context) {
	roles, err := m.rbacRepo.ListAllRoles()
	if err != nil {
		c.JSON(200, gin.H{"code": ecode.InternalError, "msg": "查询角色树失败", "request_id": c.GetString("request_id")})
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
	c.JSON(200, gin.H{
		"code":       200,
		"msg":        "success",
		"totalCount": len(children),
		"data": []gin.H{
			{
				"id":       "root",
				"label":    "全部角色",
				"children": children,
			},
		},
		"request_id": c.GetString("request_id"),
	})
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
		c.JSON(200, gin.H{"code": ecode.InvalidParams, "msg": "参数错误", "request_id": c.GetString("request_id")})
		return
	}
	name := strings.TrimSpace(in.Name)
	path := strings.TrimSpace(in.Path)
	if name == "" || path == "" {
		c.JSON(200, gin.H{"code": ecode.InvalidParams, "msg": "name与path不能为空", "request_id": c.GetString("request_id")})
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
			c.JSON(200, gin.H{"code": ecode.InternalError, "msg": "更新菜单失败", "request_id": c.GetString("request_id")})
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
			c.JSON(200, gin.H{"code": ecode.InternalError, "msg": "新增菜单失败", "request_id": c.GetString("request_id")})
			return
		}
	}
	m.pub.Publish("menu_mgmt_edit", map[string]any{
		"request_id": c.GetString("request_id"),
		"operator":   c.GetString("username"),
		"name":       name,
		"path":       path,
	})
	c.JSON(200, gin.H{"code": 200, "msg": "保存成功", "request_id": c.GetString("request_id")})
}

func (m *Module) MenuDoDelete(c *gin.Context) {
	var in struct {
		IDs string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(200, gin.H{"code": ecode.InvalidParams, "msg": "参数错误", "request_id": c.GetString("request_id")})
		return
	}
	ids := parseIDs(in.IDs)
	if err := m.rbacRepo.DeleteMenusByIDs(ids); err != nil {
		c.JSON(200, gin.H{"code": ecode.InternalError, "msg": "删除菜单失败", "request_id": c.GetString("request_id")})
		return
	}
	m.pub.Publish("menu_mgmt_delete", map[string]any{
		"request_id": c.GetString("request_id"),
		"operator":   c.GetString("username"),
		"ids":        in.IDs,
	})
	c.JSON(200, gin.H{"code": 200, "msg": "删除成功", "request_id": c.GetString("request_id")})
}
