package router

import (
	"ticket/api"

	"github.com/gin-gonic/gin"
)

func InitRouter(Router *gin.RouterGroup) {
	// 图片上传接口
	Router.POST("/upload/image", api.SavePic)

	Router.GET("/getSystemTemplate", api.GetTemplate)
	Router.GET("/getUserTemplate", api.GetUserTemplate)
}
