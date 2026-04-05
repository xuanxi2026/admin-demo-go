package admin

import (
	"strings"

	"admin-demo-go/internal/model"
	"admin-demo-go/internal/pkg/ecode"
	"admin-demo-go/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func (m *Module) DepartmentGetList(c *gin.Context) {
	list, err := m.systemRepo.ListDepartments()
	if err != nil {
		response.Fail(c, ecode.InternalError, "查询部门列表失败")
		return
	}
	response.OK(c, "success", buildDepartmentRows(list))
}

func (m *Module) DepartmentGetTree(c *gin.Context) {
	list, err := m.systemRepo.ListDepartments()
	if err != nil {
		response.Fail(c, ecode.InternalError, "查询部门树失败")
		return
	}
	response.OK(c, "success", buildDepartmentTreeOptions(list))
}

func (m *Module) DepartmentDoEdit(c *gin.Context) {
	var in struct {
		ID       uint   `json:"id"`
		ParentID uint   `json:"parentId"`
		Name     string `json:"name"`
		Code     string `json:"code"`
		Leader   string `json:"leader"`
		Phone    string `json:"phone"`
		Status   string `json:"status"`
		Sort     int    `json:"sort"`
		Remark   string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, ecode.InvalidParams, "参数错误")
		return
	}
	in.Name = strings.TrimSpace(in.Name)
	in.Code = strings.TrimSpace(in.Code)
	in.Leader = strings.TrimSpace(in.Leader)
	in.Phone = strings.TrimSpace(in.Phone)
	in.Status = strings.TrimSpace(in.Status)
	in.Remark = strings.TrimSpace(in.Remark)
	if in.Name == "" || in.Code == "" {
		response.Fail(c, ecode.InvalidParams, "部门名称和编码不能为空")
		return
	}
	if in.Status == "" {
		in.Status = "enabled"
	}
	updates := map[string]any{
		"parent_id": in.ParentID,
		"name":      in.Name,
		"code":      in.Code,
		"leader":    in.Leader,
		"phone":     in.Phone,
		"status":    in.Status,
		"sort":      in.Sort,
		"remark":    in.Remark,
	}
	if in.ID > 0 {
		if err := m.systemRepo.UpdateDepartmentByID(in.ID, updates); err != nil {
			response.Fail(c, ecode.InternalError, "更新部门失败")
			return
		}
	} else {
		if err := m.systemRepo.CreateDepartment(&model.Department{
			ParentID: in.ParentID,
			Name:     in.Name,
			Code:     in.Code,
			Leader:   in.Leader,
			Phone:    in.Phone,
			Status:   in.Status,
			Sort:     in.Sort,
			Remark:   in.Remark,
		}); err != nil {
			response.Fail(c, ecode.InternalError, "新增部门失败")
			return
		}
	}
	m.pub.Publish("department_mgmt_edit", map[string]any{
		"request_id": c.GetString("request_id"),
		"operator":   c.GetString("username"),
		"name":       in.Name,
		"code":       in.Code,
	})
	m.recordOperation(operationContext{
		Module:    "department",
		Action:    "save",
		Operator:  c.GetString("username"),
		Target:    in.Name,
		RequestID: c.GetString("request_id"),
		IP:        c.ClientIP(),
		Detail:    buildDetail("code="+in.Code, "leader="+in.Leader),
	})
	response.OK(c, "保存成功", nil)
}

func (m *Module) DepartmentDoDelete(c *gin.Context) {
	var in struct {
		IDs string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, ecode.InvalidParams, "参数错误")
		return
	}
	if err := m.systemRepo.DeleteDepartmentsByIDs(parseIDs(in.IDs)); err != nil {
		response.Fail(c, ecode.InternalError, "删除部门失败")
		return
	}
	m.pub.Publish("department_mgmt_delete", map[string]any{
		"request_id": c.GetString("request_id"),
		"operator":   c.GetString("username"),
		"ids":        in.IDs,
	})
	m.recordOperation(operationContext{
		Module:    "department",
		Action:    "delete",
		Operator:  c.GetString("username"),
		Target:    joinIDs(in.IDs),
		RequestID: c.GetString("request_id"),
		IP:        c.ClientIP(),
		Detail:    "删除部门",
	})
	response.OK(c, "删除成功", nil)
}

func buildDepartmentRows(list []model.Department) []gin.H {
	childrenMap := map[uint][]model.Department{}
	for _, item := range list {
		childrenMap[item.ParentID] = append(childrenMap[item.ParentID], item)
	}
	var build func(parentID uint) []gin.H
	build = func(parentID uint) []gin.H {
		children := childrenMap[parentID]
		rows := make([]gin.H, 0, len(children))
		for _, item := range children {
			rows = append(rows, gin.H{
				"id":          item.ID,
				"parentId":    item.ParentID,
				"name":        item.Name,
				"code":        item.Code,
				"leader":      item.Leader,
				"phone":       item.Phone,
				"status":      item.Status,
				"sort":        item.Sort,
				"remark":      item.Remark,
				"datatime":    item.UpdatedAt.Format("2006-01-02 15:04:05"),
				"children":    build(item.ID),
				"hasChildren": len(childrenMap[item.ID]) > 0,
			})
		}
		return rows
	}
	return build(0)
}

func buildDepartmentTreeOptions(list []model.Department) []gin.H {
	childrenMap := map[uint][]model.Department{}
	for _, item := range list {
		childrenMap[item.ParentID] = append(childrenMap[item.ParentID], item)
	}
	var build func(parentID uint) []gin.H
	build = func(parentID uint) []gin.H {
		children := childrenMap[parentID]
		rows := make([]gin.H, 0, len(children))
		for _, item := range children {
			rows = append(rows, gin.H{
				"id":       item.ID,
				"label":    item.Name,
				"code":     item.Code,
				"children": build(item.ID),
			})
		}
		return rows
	}
	return []gin.H{
		{
			"id":       "root",
			"label":    "全部部门",
			"children": build(0),
		},
	}
}
