package handler

import (
	"github.com/gin-gonic/gin"

	"ticket/internal/config"
	"ticket/internal/pkg/response"
	"ticket/internal/service"
)

// PublicTemplates 编辑器使用的系统模板列表。
func PublicTemplates(c *gin.Context) {
	list, err := service.PublicTemplates(config.DB)
	if err != nil {
		response.ServerError(c, "获取模板失败")
		return
	}
	response.OK(c, gin.H{"list": list, "total": len(list)})
}
