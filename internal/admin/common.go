package admin

import (
	"strconv"
	"strings"
)

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
