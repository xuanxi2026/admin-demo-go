package admin

import (
	"strings"

	"admin-demo-go/internal/model"
	"admin-demo-go/internal/pkg/ecode"
	"admin-demo-go/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func (m *Module) ConfigGetList(c *gin.Context) {
	var in struct {
		PageNo   int    `json:"pageNo"`
		PageSize int    `json:"pageSize"`
		Group    string `json:"group"`
		Keyword  string `json:"keyword"`
	}
	_ = c.ShouldBindJSON(&in)
	in.PageNo, in.PageSize = normalizePage(in.PageNo, in.PageSize)

	list, total, err := m.systemRepo.ListConfigs(in.PageNo, in.PageSize, in.Group, in.Keyword)
	if err != nil {
		response.Fail(c, ecode.InternalError, "查询系统配置失败")
		return
	}

	rows := make([]gin.H, 0, len(list))
	for _, item := range list {
		rows = append(rows, gin.H{
			"id":          item.ID,
			"configKey":   item.ConfigKey,
			"configValue": item.ConfigValue,
			"name":        item.Name,
			"group":       item.Group,
			"valueType":   item.ValueType,
			"remark":      item.Remark,
			"datatime":    item.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	response.List(c, "success", rows, total)
}

func (m *Module) ConfigDoEdit(c *gin.Context) {
	var in struct {
		ID          uint   `json:"id"`
		ConfigKey   string `json:"configKey"`
		ConfigValue string `json:"configValue"`
		Name        string `json:"name"`
		Group       string `json:"group"`
		ValueType   string `json:"valueType"`
		Remark      string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, ecode.InvalidParams, "参数错误")
		return
	}

	in.ConfigKey = strings.TrimSpace(in.ConfigKey)
	in.ConfigValue = strings.TrimSpace(in.ConfigValue)
	in.Name = strings.TrimSpace(in.Name)
	in.Group = strings.TrimSpace(in.Group)
	in.ValueType = strings.TrimSpace(in.ValueType)
	in.Remark = strings.TrimSpace(in.Remark)

	if in.ConfigKey == "" || in.Name == "" || in.Group == "" {
		response.Fail(c, ecode.InvalidParams, "配置键、名称和分组不能为空")
		return
	}
	if in.ValueType == "" {
		in.ValueType = "string"
	}

	updates := map[string]any{
		"config_key":   in.ConfigKey,
		"config_value": in.ConfigValue,
		"name":         in.Name,
		"group":        in.Group,
		"value_type":   in.ValueType,
		"remark":       in.Remark,
	}
	if in.ID > 0 {
		if err := m.systemRepo.UpdateConfigByID(in.ID, updates); err != nil {
			response.Fail(c, ecode.InternalError, "更新系统配置失败")
			return
		}
	} else {
		if err := m.systemRepo.CreateConfig(&model.SystemConfig{
			ConfigKey:   in.ConfigKey,
			ConfigValue: in.ConfigValue,
			Name:        in.Name,
			Group:       in.Group,
			ValueType:   in.ValueType,
			Remark:      in.Remark,
		}); err != nil {
			response.Fail(c, ecode.InternalError, "新增系统配置失败")
			return
		}
	}

	m.pub.Publish("config_mgmt_edit", map[string]any{
		"request_id": c.GetString("request_id"),
		"operator":   c.GetString("username"),
		"config_key": in.ConfigKey,
		"group":      in.Group,
	})
	m.recordOperation(operationContext{
		Module:    "config",
		Action:    "save",
		Operator:  c.GetString("username"),
		Target:    in.ConfigKey,
		RequestID: c.GetString("request_id"),
		IP:        c.ClientIP(),
		Detail:    buildDetail("group="+in.Group, "type="+in.ValueType),
	})
	response.OK(c, "保存成功", nil)
}

func (m *Module) ConfigDoDelete(c *gin.Context) {
	var in struct {
		IDs string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, ecode.InvalidParams, "参数错误")
		return
	}
	if err := m.systemRepo.DeleteConfigsByIDs(parseIDs(in.IDs)); err != nil {
		response.Fail(c, ecode.InternalError, "删除系统配置失败")
		return
	}
	m.pub.Publish("config_mgmt_delete", map[string]any{
		"request_id": c.GetString("request_id"),
		"operator":   c.GetString("username"),
		"ids":        in.IDs,
	})
	m.recordOperation(operationContext{
		Module:    "config",
		Action:    "delete",
		Operator:  c.GetString("username"),
		Target:    joinIDs(in.IDs),
		RequestID: c.GetString("request_id"),
		IP:        c.ClientIP(),
		Detail:    "删除系统配置",
	})
	response.OK(c, "删除成功", nil)
}
