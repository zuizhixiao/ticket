// Package response 统一 JSON 响应壳:{code, message, data}。
// code=0 成功;失败时 code 与 HTTP 状态一致(400/401/403/404/409/500...)。
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Body struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

// OK 成功响应。
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Body{Code: 0, Message: "ok", Data: data})
}

// OKMessage 携带自定义成功消息。
func OKMessage(c *gin.Context, message string, data any) {
	c.JSON(http.StatusOK, Body{Code: 0, Message: message, Data: data})
}

// Error 统一失败响应(HTTP 状态码 = code)。
func Error(c *gin.Context, code int, message string) {
	c.JSON(code, Body{Code: code, Message: message, Data: nil})
	c.Abort()
}

// ParamError 参数错误(400)。
func ParamError(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

// BizError 业务失败(422,如验证码/密码错)。
func BizError(c *gin.Context, message string) {
	Error(c, http.StatusUnprocessableEntity, message)
}

// Unauthorized 401。
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, message)
}

// Forbidden 403。
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, message)
}

// NotFound 404。
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}

// Conflict 409(如昵称已存在)。
func Conflict(c *gin.Context, message string) {
	Error(c, http.StatusConflict, message)
}

// ServerError 500。
func ServerError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message)
}
