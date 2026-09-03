package handler

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"ticket/internal/config"
	"ticket/internal/model"
	"ticket/internal/pkg/response"
	"ticket/internal/repository"
	"ticket/internal/service"
)

// AdminTemplates 模板列表(?status=1|2,默认全部)。
func AdminTemplates(c *gin.Context) {
	status, _ := strconv.Atoi(c.DefaultQuery("status", "0"))
	list, err := repository.AdminListTemplates(config.DB, status)
	if err != nil {
		response.ServerError(c, "查询模板失败")
		return
	}
	response.OK(c, gin.H{"list": list, "total": len(list)})
}

// AdminCreateTemplate 新增系统模板。
func AdminCreateTemplate(c *gin.Context) {
	var in service.TemplateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.ParamError(c, "参数格式错误")
		return
	}
	t, err := service.AdminCreateTemplate(config.DB, in)
	if err != nil {
		replyServiceError(c, err)
		return
	}
	response.OKMessage(c, "创建成功", t)
}

// AdminUpdateTemplate 编辑模板/上下架。
func AdminUpdateTemplate(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.ParamError(c, "参数错误")
		return
	}
	var in service.TemplateInput
	if err := c.ShouldBindJSON(&in); err != nil {
		response.ParamError(c, "参数格式错误")
		return
	}
	t, err := service.AdminUpdateTemplate(config.DB, id, in)
	if err != nil {
		replyServiceError(c, err)
		return
	}
	response.OKMessage(c, "已保存", t)
}

// AdminImages 查看全部成品/海报(type=product|poster,分页)。
func AdminImages(c *gin.Context) {
	imgType := c.Query("type")
	if imgType != model.ImageTypeProduct && imgType != model.ImageTypePoster {
		response.ParamError(c, "type 仅支持 product 或 poster")
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "24"))
	list, total, err := service.AdminListImages(config.DB, imgType, page, size)
	if err != nil {
		response.ServerError(c, "查询失败")
		return
	}
	response.OK(c, gin.H{"list": list, "total": total, "page": page, "size": size})
}

// AdminDeleteImage 管理员删除任意图片。
func AdminDeleteImage(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.ParamError(c, "参数错误")
		return
	}
	if err := service.AdminDeleteImage(config.DB, id); err != nil {
		replyServiceError(c, err)
		return
	}
	response.OKMessage(c, "已删除", nil)
}

// AdminDeleteTemplate 软删模板。
func AdminDeleteTemplate(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.ParamError(c, "参数错误")
		return
	}
	if err := service.AdminDeleteTemplate(config.DB, id); err != nil {
		if errors.Is(err, service.ErrTemplateNotFound) {
			response.NotFound(c, "模板不存在")
			return
		}
		response.ServerError(c, "删除失败")
		return
	}
	response.OKMessage(c, "已下架", nil)
}
