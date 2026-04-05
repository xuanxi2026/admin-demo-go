package admin

import (
	"admin-demo-go/internal/pkg/ecode"
	"admin-demo-go/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func (m *Module) OperationLogGetList(c *gin.Context) {
	var in struct {
		PageNo   int    `json:"pageNo"`
		PageSize int    `json:"pageSize"`
		Module   string `json:"module"`
		Keyword  string `json:"keyword"`
	}
	_ = c.ShouldBindJSON(&in)
	in.PageNo, in.PageSize = normalizePage(in.PageNo, in.PageSize)

	list, total, err := m.systemRepo.ListOperationLogs(in.PageNo, in.PageSize, in.Module, in.Keyword)
	if err != nil {
		response.Fail(c, ecode.InternalError, "查询操作日志失败")
		return
	}

	rows := make([]gin.H, 0, len(list))
	for _, item := range list {
		rows = append(rows, gin.H{
			"id":        item.ID,
			"module":    item.Module,
			"action":    item.Action,
			"operator":  item.Operator,
			"target":    item.Target,
			"requestId": item.RequestID,
			"ip":        item.IP,
			"detail":    item.Detail,
			"datatime":  item.CreatedAt.Format("2006-01-02 15:04:05"),
		})
	}
	response.List(c, "success", rows, total)
}
