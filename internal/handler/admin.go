package handler

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"ticket/internal/config"
	"ticket/internal/middleware"
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
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
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

// AdminUsers 用户列表(keyword 昵称模糊,分页)。
func AdminUsers(c *gin.Context) {
	keyword := strings.TrimSpace(c.Query("keyword"))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	list, total, err := service.AdminListUsers(config.DB, keyword, page, size)
	if err != nil {
		response.ServerError(c, "查询失败")
		return
	}
	response.OK(c, gin.H{"list": list, "total": total, "page": page, "size": size})
}

// AdminUserStatus 冻结(0)/解冻(1)。
func AdminUserStatus(c *gin.Context) {
	operator, _ := middleware.ClaimsOf(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.ParamError(c, "参数错误")
		return
	}
	var req struct {
		Status int `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数格式错误")
		return
	}
	if req.Status != 0 && req.Status != 1 {
		response.ParamError(c, "status 仅支持 0(冻结)/1(正常)")
		return
	}
	if err := service.AdminSetUserStatus(config.DB, operator.UserId, id, req.Status); err != nil {
		replyServiceError(c, err)
		return
	}
	if req.Status == 0 {
		response.OKMessage(c, "账号已冻结", nil)
	} else {
		response.OKMessage(c, "账号已解冻", nil)
	}
}

// AdminUserResetPassword 重置用户密码。
func AdminUserResetPassword(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.ParamError(c, "参数错误")
		return
	}
	var req struct {
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数格式错误")
		return
	}
	if err := service.AdminResetUserPassword(config.DB, id, req.Password); err != nil {
		replyServiceError(c, err)
		return
	}
	response.OKMessage(c, "密码已重置", nil)
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
