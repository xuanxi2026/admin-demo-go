package response

import "github.com/gin-gonic/gin"

func OK(c *gin.Context, msg string, data any) {
	c.JSON(200, gin.H{
		"code":       200,
		"msg":        msg,
		"data":       data,
		"request_id": requestID(c),
	})
}

func List(c *gin.Context, msg string, data any, totalCount int64) {
	c.JSON(200, gin.H{
		"code":       200,
		"msg":        msg,
		"data":       data,
		"totalCount": totalCount,
		"request_id": requestID(c),
	})
}

func Fail(c *gin.Context, code int, msg string) {
	c.JSON(200, gin.H{
		"code":       code,
		"msg":        msg,
		"request_id": requestID(c),
	})
}

func requestID(c *gin.Context) string {
	v, ok := c.Get("request_id")
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}
