package admin

import (
	"fmt"
	"strconv"
	"strings"

	"admin-demo-go/internal/model"
)

func normalizePage(pageNo, pageSize int) (int, int) {
	if pageNo <= 0 {
		pageNo = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return pageNo, pageSize
}

func parseIDs(csv string) []uint {
	parts := strings.Split(strings.TrimSpace(csv), ",")
	ids := make([]uint, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		id64, err := strconv.ParseUint(p, 10, 64)
		if err != nil {
			continue
		}
		ids = append(ids, uint(id64))
	}
	return ids
}

func (m *Module) recordOperation(c operationContext) {
	if m == nil || m.systemRepo == nil {
		return
	}
	_ = m.systemRepo.CreateOperationLog(&model.OperationLog{
		Module:    c.Module,
		Action:    c.Action,
		Operator:  strings.TrimSpace(c.Operator),
		Target:    strings.TrimSpace(c.Target),
		RequestID: strings.TrimSpace(c.RequestID),
		IP:        strings.TrimSpace(c.IP),
		Detail:    strings.TrimSpace(c.Detail),
	})
}

type operationContext struct {
	Module    string
	Action    string
	Operator  string
	Target    string
	RequestID string
	IP        string
	Detail    string
}

func buildDetail(parts ...string) string {
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			items = append(items, part)
		}
	}
	return strings.Join(items, "; ")
}

func joinIDs(ids string) string {
	normalized := parseIDs(ids)
	if len(normalized) == 0 {
		return ""
	}
	values := make([]string, 0, len(normalized))
	for _, id := range normalized {
		values = append(values, fmt.Sprintf("%d", id))
	}
	return strings.Join(values, ",")
}
