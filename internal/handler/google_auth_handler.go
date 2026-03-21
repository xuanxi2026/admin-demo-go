package handler

import (
	"admin-demo-go/internal/pkg/ecode"
	"admin-demo-go/internal/pkg/event"
	"admin-demo-go/internal/pkg/response"
	"admin-demo-go/internal/repository"
	"admin-demo-go/internal/service"

	"github.com/gin-gonic/gin"
)

type GoogleAuthHandler struct {
	userRepo   *repository.UserRepository
	googleAuth *service.GoogleAuthService
	pub        *event.Publisher
}

func NewGoogleAuthHandler(userRepo *repository.UserRepository, googleAuth *service.GoogleAuthService, pub *event.Publisher) *GoogleAuthHandler {
	return &GoogleAuthHandler{
		userRepo:   userRepo,
		googleAuth: googleAuth,
		pub:        pub,
	}
}

func (h *GoogleAuthHandler) Setup(c *gin.Context) {
	user, err := h.userRepo.FindByID(c.GetUint("userID"))
	if err != nil {
		response.Fail(c, ecode.InternalError, "用户不存在")
		return
	}
	data, err := h.googleAuth.Setup(user)
	if err != nil {
		response.Fail(c, ecode.InternalError, "生成二维码失败")
		return
	}
	h.pub.Publish("google_setup", map[string]any{
		"request_id": c.GetString("request_id"),
		"user_id":    user.ID,
	})
	response.OK(c, "success", data)
}

func (h *GoogleAuthHandler) Bind(c *gin.Context) {
	var in struct {
		Code string `json:"code"`
	}
	if err := c.ShouldBindJSON(&in); err != nil || in.Code == "" {
		response.Fail(c, ecode.InvalidParams, "验证码不能为空")
		return
	}
	userID := c.GetUint("userID")
	if err := h.googleAuth.ConfirmBind(userID, in.Code); err != nil {
		response.Fail(c, ecode.GoogleCodeFailed, err.Error())
		return
	}
	h.pub.Publish("google_bind", map[string]any{
		"request_id": c.GetString("request_id"),
		"user_id":    userID,
	})
	response.OK(c, "绑定成功", nil)
}

func (h *GoogleAuthHandler) Unbind(c *gin.Context) {
	var in struct {
		Code string `json:"code"`
	}
	if err := c.ShouldBindJSON(&in); err != nil || in.Code == "" {
		response.Fail(c, ecode.InvalidParams, "验证码不能为空")
		return
	}
	userID := c.GetUint("userID")
	if err := h.googleAuth.Unbind(userID, in.Code); err != nil {
		response.Fail(c, ecode.GoogleCodeFailed, err.Error())
		return
	}
	h.pub.Publish("google_unbind", map[string]any{
		"request_id": c.GetString("request_id"),
		"user_id":    userID,
	})
	response.OK(c, "解绑成功", nil)
}
