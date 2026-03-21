package handler

import (
	"log"

	"admin-demo-go/internal/pkg/apperr"
	"admin-demo-go/internal/pkg/ecode"
	"admin-demo-go/internal/pkg/event"
	"admin-demo-go/internal/pkg/response"
	"admin-demo-go/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authSvc *service.AuthService
	pub     *event.Publisher
}

func NewAuthHandler(authSvc *service.AuthService, pub *event.Publisher) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, pub: pub}
}

func (h *AuthHandler) Register(c *gin.Context) {
	var in service.RegisterInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Fail(c, ecode.InvalidParams, "参数错误")
		return
	}
	if err := h.authSvc.Register(in); err != nil {
		code, msg := apperr.Parse(err, ecode.InternalError, "注册失败")
		response.Fail(c, code, msg)
		return
	}
	h.pub.Publish("user_register", map[string]any{
		"request_id": c.GetString("request_id"),
		"username":   in.Username,
	})
	response.OK(c, "注册成功", nil)
}

func (h *AuthHandler) Login(c *gin.Context) {
	reqID := c.GetString("request_id")
	var in service.LoginInput
	if err := c.ShouldBindJSON(&in); err != nil {
		log.Printf("request_id=%s stage=login_bind_json status=failed err=%v", reqID, err)
		response.Fail(c, ecode.InvalidParams, "参数错误")
		return
	}
	log.Printf("request_id=%s stage=login_start username=%s", reqID, in.Username)
	token, user, googleBound, err := h.authSvc.Login(reqID, in)
	if err != nil {
		code, msg := apperr.Parse(err, ecode.InternalError, "登录失败")
		log.Printf("request_id=%s stage=login_end status=failed code=%d msg=%s username=%s", reqID, code, msg, in.Username)
		response.Fail(c, code, msg)
		h.pub.Publish("user_login_failed", map[string]any{
			"request_id": reqID,
			"username":   in.Username,
			"reason":     msg,
		})
		return
	}
	log.Printf("request_id=%s stage=login_end status=success user_id=%d username=%s google_bound=%v", reqID, user.ID, user.Username, googleBound)
	h.pub.Publish("user_login", map[string]any{
		"request_id":   reqID,
		"user_id":      user.ID,
		"username":     user.Username,
		"google_bound": googleBound,
	})
	response.OK(c, "success", gin.H{
		"accessToken": token,
		"googleBound": googleBound,
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	h.pub.Publish("user_logout", map[string]any{
		"request_id": c.GetString("request_id"),
		"user_id":    c.GetUint("userID"),
	})
	response.OK(c, "success", nil)
}
