package middleware

import (
	"admin-demo-go/internal/pkg/ecode"
	"admin-demo-go/internal/pkg/response"
	"admin-demo-go/internal/service"

	"github.com/gin-gonic/gin"
)

func RequirePermission(rbacSvc *service.RBACService, code string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetUint("userID")
		ok, err := rbacSvc.HasPermission(userID, code)
		if err != nil {
			response.Fail(c, ecode.InternalError, "权限校验失败")
			c.Abort()
			return
		}
		if !ok {
			response.Fail(c, ecode.Forbidden, "无权限访问")
			c.Abort()
			return
		}
		c.Next()
	}
}
