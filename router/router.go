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
}
