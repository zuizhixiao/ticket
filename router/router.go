package router

import (
	"ticket/api"

	"github.com/gin-gonic/gin"
)

func InitRouter(Router *gin.RouterGroup) {
	// 图片上传接口
	Router.POST("/upload/image", api.SavePic)

	Router.GET("/getSystemTemplate", api.GetTemplate)
	Router.GET("/getTypeSettingTemplate", api.GetTypeSettingTemplate)

	// 微信公众号消息接收接口
	Router.Any("/wechat/message", api.WechatMessageHandler)

	// 验证验证码并获取用户信息接口
	Router.POST("/wechat/verify-code", api.VerifyCodeAndGetUserInfo)
	Router.GET("/wechat/verify-code", api.VerifyCodeAndGetUserInfo)

	// 认证相关接口
	Router.POST("/auth/login", api.Login)
	Router.POST("/auth/register", api.Register)
	Router.POST("/auth/reset-password", api.ResetPassword)
	Router.GET("/auth/captcha", api.GetCaptcha)
	Router.GET("/auth/user-info", api.GetUserInfo)

	// 用户成品
	Router.GET("/user/products", api.GetUserProductList)
}
