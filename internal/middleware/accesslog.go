package middleware

import (
	"log"
	"time"

	"admin-demo-go/internal/pkg/event"

	"github.com/gin-gonic/gin"
)

func AccessLog(pub *event.Publisher) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		latency := time.Since(start).Milliseconds()
		status := c.Writer.Status()
		reqID, _ := c.Get("request_id")
		userID, _ := c.Get("userID")

		log.Printf("request_id=%v method=%s path=%s status=%d latency_ms=%d ip=%s user_id=%v",
			reqID, c.Request.Method, c.Request.URL.Path, status, latency, c.ClientIP(), userID)

		pub.Publish("http_access", map[string]any{
			"request_id": reqID,
			"method":     c.Request.Method,
			"path":       c.Request.URL.Path,
			"status":     status,
			"latency_ms": latency,
			"ip":         c.ClientIP(),
			"user_id":    userID,
		})
	}
}
