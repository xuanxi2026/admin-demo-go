package admin

import (
	"strings"

	"admin-demo-go/internal/model"
	"admin-demo-go/internal/pkg/ecode"
	"admin-demo-go/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func (m *Module) NoticeGetList(c *gin.Context) {
	var in struct {
		PageNo   int    `json:"pageNo"`
		PageSize int    `json:"pageSize"`
		Level    string `json:"level"`
		Status   string `json:"status"`
		Keyword  string `json:"keyword"`
	}
	_ = c.ShouldBindJSON(&in)
	in.PageNo, in.PageSize = normalizePage(in.PageNo, in.PageSize)
	list, total, err := m.systemRepo.ListNotices(in.PageNo, in.PageSize, in.Level, in.Status, in.Keyword)
	if err != nil {
		response.Fail(c, ecode.InternalError, "查询通知公告失败")
		return
	}
	rows := make([]gin.H, 0, len(list))
	for _, item := range list {
		rows = append(rows, gin.H{
			"id":        item.ID,
			"title":     item.Title,
			"content":   item.Content,
			"level":     item.Level,
			"status":    item.Status,
			"publisher": item.Publisher,
			"sort":      item.Sort,
			"remark":    item.Remark,
			"datatime":  item.UpdatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	response.List(c, "success", rows, total)
}

func (m *Module) NoticeDoEdit(c *gin.Context) {
	var in struct {
		ID        uint   `json:"id"`
		Title     string `json:"title"`
		Content   string `json:"content"`
		Level     string `json:"level"`
		Status    string `json:"status"`
		Publisher string `json:"publisher"`
		Sort      int    `json:"sort"`
		Remark    string `json:"remark"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, ecode.InvalidParams, "参数错误")
		return
	}
	in.Title = strings.TrimSpace(in.Title)
	in.Content = strings.TrimSpace(in.Content)
	in.Level = strings.TrimSpace(in.Level)
	in.Status = strings.TrimSpace(in.Status)
	in.Publisher = strings.TrimSpace(in.Publisher)
	in.Remark = strings.TrimSpace(in.Remark)
	if in.Title == "" || in.Content == "" {
		response.Fail(c, ecode.InvalidParams, "标题和内容不能为空")
		return
	}
	if in.Level == "" {
		in.Level = "normal"
	}
	if in.Status == "" {
		in.Status = "draft"
	}
	updates := map[string]any{
		"title":     in.Title,
		"content":   in.Content,
		"level":     in.Level,
		"status":    in.Status,
		"publisher": in.Publisher,
		"sort":      in.Sort,
		"remark":    in.Remark,
	}
	if in.ID > 0 {
		if err := m.systemRepo.UpdateNoticeByID(in.ID, updates); err != nil {
			response.Fail(c, ecode.InternalError, "更新通知公告失败")
			return
		}
	} else {
		if err := m.systemRepo.CreateNotice(&model.Notice{
			Title:     in.Title,
			Content:   in.Content,
			Level:     in.Level,
			Status:    in.Status,
			Publisher: in.Publisher,
			Sort:      in.Sort,
			Remark:    in.Remark,
		}); err != nil {
			response.Fail(c, ecode.InternalError, "新增通知公告失败")
			return
		}
	}
	m.pub.Publish("notice_mgmt_edit", map[string]any{
		"request_id": c.GetString("request_id"),
		"operator":   c.GetString("username"),
		"title":      in.Title,
		"level":      in.Level,
	})
	m.recordOperation(operationContext{
		Module:    "notice",
		Action:    "save",
		Operator:  c.GetString("username"),
		Target:    in.Title,
		RequestID: c.GetString("request_id"),
		IP:        c.ClientIP(),
		Detail:    buildDetail("level="+in.Level, "status="+in.Status),
	})
	response.OK(c, "保存成功", nil)
}

func (m *Module) NoticeDoDelete(c *gin.Context) {
	var in struct {
		IDs string `json:"ids"`
	}
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, ecode.InvalidParams, "参数错误")
		return
	}
	if err := m.systemRepo.DeleteNoticesByIDs(parseIDs(in.IDs)); err != nil {
		response.Fail(c, ecode.InternalError, "删除通知公告失败")
		return
	}
	m.pub.Publish("notice_mgmt_delete", map[string]any{
		"request_id": c.GetString("request_id"),
		"operator":   c.GetString("username"),
		"ids":        in.IDs,
	})
	m.recordOperation(operationContext{
		Module:    "notice",
		Action:    "delete",
		Operator:  c.GetString("username"),
		Target:    joinIDs(in.IDs),
		RequestID: c.GetString("request_id"),
		IP:        c.ClientIP(),
		Detail:    "删除通知公告",
	})
	response.OK(c, "删除成功", nil)
}
