package admin

import (
	"strings"

	"admin-demo-go/internal/model"
	"admin-demo-go/internal/pkg/ecode"
	"admin-demo-go/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func (m *Module) DictGetList(c *gin.Context) {
	var in struct {
		PageNo   int    `json:"pageNo"`
		PageSize int    `json:"pageSize"`
		DictType string `json:"dictType"`
		Keyword  string `json:"keyword"`
	}
	_ = c.ShouldBindJSON(&in)
	in.PageNo, in.PageSize = normalizePage(in.PageNo, in.PageSize)

	list, total, err := m.systemRepo.ListDictItems(in.PageNo, in.PageSize, in.DictType, in.Keyword)
	if err != nil {
		response.Fail(c, ecode.InternalError, "查询字典项失败")
		return
	}

	rows := make([]gin.H, 0, len(list))
	for _, item := range list {
		rows = append(rows, gin.H{
			"id":       item.ID,
			"dictType": item.DictType,
			"label":    item.Label,
			"value":    item.Value,
			"status":   item.Status,
			"sort":     item.Sort,
			"remark":   item.Remark,
			"datatime": item.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	response.List(c, "success", rows, total)
}

func (m *Module) DictDoEdit(c *gin.Context) {
	var in struct {
		ID       uint   `json:"id"`
		DictType string `json:"dictType"`
		Label    string `json:"label"`
		Value    string `json:"value"`
		Status   string `json:"status"`
		Sort     int    `json:"sort"`
		Remark   string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, ecode.InvalidParams, "参数错误")
		return
	}

	in.DictType = strings.TrimSpace(in.DictType)
	in.Label = strings.TrimSpace(in.Label)
	in.Value = strings.TrimSpace(in.Value)
	in.Status = strings.TrimSpace(in.Status)
	in.Remark = strings.TrimSpace(in.Remark)

	if in.DictType == "" || in.Label == "" || in.Value == "" {
		response.Fail(c, ecode.InvalidParams, "字典类型、标签和值不能为空")
		return
	}
	if in.Status == "" {
		in.Status = "enabled"
	}

	updates := map[string]any{
		"dict_type": in.DictType,
		"label":     in.Label,
		"value":     in.Value,
		"status":    in.Status,
		"sort":      in.Sort,
		"remark":    in.Remark,
	}
	if in.ID > 0 {
		if err := m.systemRepo.UpdateDictItemByID(in.ID, updates); err != nil {
			response.Fail(c, ecode.InternalError, "更新字典项失败")
			return
		}
	} else {
		if err := m.systemRepo.CreateDictItem(&model.DictItem{
			DictType: in.DictType,
			Label:    in.Label,
			Value:    in.Value,
			Status:   in.Status,
			Sort:     in.Sort,
			Remark:   in.Remark,
		}); err != nil {
			response.Fail(c, ecode.InternalError, "新增字典项失败")
			return
		}
	}

	m.pub.Publish("dict_mgmt_edit", map[string]any{
		"request_id": c.GetString("request_id"),
		"operator":   c.GetString("username"),
		"dict_type":  in.DictType,
		"value":      in.Value,
	})
	m.recordOperation(operationContext{
		Module:    "dict",
		Action:    "save",
		Operator:  c.GetString("username"),
		Target:    in.DictType + ":" + in.Value,
		RequestID: c.GetString("request_id"),
		IP:        c.ClientIP(),
		Detail:    buildDetail("label="+in.Label, "status="+in.Status),
	})
	response.OK(c, "保存成功", nil)
}

func (m *Module) DictDoDelete(c *gin.Context) {
	var in struct {
		IDs string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, ecode.InvalidParams, "参数错误")
		return
	}
	if err := m.systemRepo.DeleteDictItemsByIDs(parseIDs(in.IDs)); err != nil {
		response.Fail(c, ecode.InternalError, "删除字典项失败")
		return
	}
	m.pub.Publish("dict_mgmt_delete", map[string]any{
		"request_id": c.GetString("request_id"),
		"operator":   c.GetString("username"),
		"ids":        in.IDs,
	})
	m.recordOperation(operationContext{
		Module:    "dict",
		Action:    "delete",
		Operator:  c.GetString("username"),
		Target:    joinIDs(in.IDs),
		RequestID: c.GetString("request_id"),
		IP:        c.ClientIP(),
		Detail:    "删除字典项",
	})
	response.OK(c, "删除成功", nil)
}
