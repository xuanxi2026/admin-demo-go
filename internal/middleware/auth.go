package middleware

import (
	"strings"

	"admin-demo-go/internal/pkg/ecode"
	"admin-demo-go/internal/pkg/jwt"
	"admin-demo-go/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("accessToken")
		if token == "" {
			authHeader := c.GetHeader("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}
		if token == "" {
			response.Fail(c, ecode.Unauthorized, "请先登录")
			c.Abort()
			return
		}

		claims, err := jwt.Parse(secret, token)
		if err != nil {
			response.Fail(c, ecode.TokenExpired, "登录已过期，请重新登录")
			c.Abort()
			return
		}
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}
