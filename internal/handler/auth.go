// Package handler HTTP 处理层。
package handler

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"

	"ticket/internal/config"
	"ticket/internal/middleware"
	"ticket/internal/pkg/captcha"
	"ticket/internal/pkg/response"
	"ticket/internal/repository"
	"ticket/internal/service"
)

// ---------- 图形验证码 ----------

func GetCaptcha(c *gin.Context) {
	id, b64, err := captcha.Generate()
	if err != nil {
		response.ServerError(c, "生成验证码失败")
		return
	}
	response.OK(c, gin.H{"captchaId": id, "captchaImg": b64})
}

// ---------- 注册 ----------

type registerReq struct {
	Nickname  string `json:"nickname"`
	Password  string `json:"password"`
	CaptchaId string `json:"captchaId"`
	Captcha   string `json:"captcha"`
}

func Register(c *gin.Context) {
	var req registerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数格式错误")
		return
	}
	if !captcha.Verify(req.CaptchaId, req.Captcha) {
		response.BizError(c, "图形验证码错误")
		return
	}
	u, err := service.Register(config.DB, req.Nickname, req.Password)
	if err != nil {
		replyServiceError(c, err)
		return
	}
	response.OKMessage(c, "注册成功", u.Profile())
}

// ---------- 登录 ----------

type loginReq struct {
	Nickname string `json:"nickname"`
	Password string `json:"password"`
}

func Login(c *gin.Context) {
	var req loginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数格式错误")
		return
	}
	u, token, err := service.Login(config.DB, req.Nickname, req.Password)
	if err != nil {
		replyServiceError(c, err)
		return
	}
	response.OK(c, gin.H{"token": token, "user": u.Profile()})
}

// ---------- 找回密码 ----------

type resetPwdReq struct {
	Nickname  string `json:"nickname"`
	Password  string `json:"password"`
	CaptchaId string `json:"captchaId"`
	Captcha   string `json:"captcha"`
}

func ResetPassword(c *gin.Context) {
	var req resetPwdReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数格式错误")
		return
	}
	if !captcha.Verify(req.CaptchaId, req.Captcha) {
		response.BizError(c, "图形验证码错误")
		return
	}
	if err := service.ResetPassword(config.DB, req.Nickname, req.Password); err != nil {
		replyServiceError(c, err)
		return
	}
	response.OKMessage(c, "密码已重置,请重新登录", nil)
}

// ---------- 当前用户 ----------

func Me(c *gin.Context) {
	claims, ok := middleware.ClaimsOf(c)
	if !ok {
		response.Unauthorized(c, "请先登录")
		return
	}
	u, err := repository.GetUserByID(config.DB, claims.UserId)
	if err != nil {
		response.Unauthorized(c, "用户不存在")
		return
	}
	if u.Status == 0 {
		response.Forbidden(c, "账号已被禁用")
		return
	}
	response.OK(c, u.Profile())
}

// ---------- 资料更新 ----------

type profileReq struct {
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
}

func UpdateProfile(c *gin.Context) {
	claims, _ := middleware.ClaimsOf(c)
	var req profileReq
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数格式错误")
		return
	}
	if err := service.UpdateProfile(config.DB, claims.UserId, req.Nickname, req.Avatar); err != nil {
		replyServiceError(c, err)
		return
	}
	u, err := repository.GetUserByID(config.DB, claims.UserId)
	if err != nil {
		response.ServerError(c, "读取用户失败")
		return
	}
	response.OKMessage(c, "已保存", u.Profile())
}

// replyServiceError 领域错误 → HTTP 响应。
func replyServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrNicknameTaken), errors.Is(err, service.ErrNicknameLen), errors.Is(err, service.ErrPasswordLen):
		response.BizError(c, err.Error())
	case errors.Is(err, service.ErrAdminProtected), errors.Is(err, service.ErrSelfOperate):
		response.Forbidden(c, err.Error())
	case errors.Is(err, service.ErrUserNotFound), errors.Is(err, service.ErrTemplateNotFound),
		errors.Is(err, service.ErrImageNotFound), errors.Is(err, service.ErrNotYourImage):
		response.NotFound(c, err.Error())
	case errors.Is(err, service.ErrWrongCredential), errors.Is(err, service.ErrBanned):
		response.BizError(c, err.Error())
	case errors.Is(err, service.ErrWechatService), errors.Is(err, service.ErrNotSubscribed),
		errors.Is(err, service.ErrImageType), errors.Is(err, service.ErrImageExt),
		errors.Is(err, service.ErrImageTooLarge), errors.Is(err, service.ErrTemplateDenied),
		errors.Is(err, service.ErrStorageUpload), errors.Is(err, service.ErrStorageDelete):
		response.BizError(c, err.Error())
	case err != nil && strings.Contains(err.Error(), "验证码"):
		response.BizError(c, err.Error())
	case err != nil:
		response.ServerError(c, "操作失败,请稍后重试")
	}
}
