package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"ticket/internal/config"
	"ticket/internal/middleware"
	"ticket/internal/model"
	"ticket/internal/pkg/response"
	"ticket/internal/service"
)

// Upload 上传图片(multipart: file + type)。
// 支持游客上传 poster/product(记录 userId=0);
// 带有效登录态时记录归属;avatar/template 仅登录且管理员可用。
func Upload(c *gin.Context) {
	claims, _ := middleware.ClaimsOf(c) // 可选鉴权:可能为 nil(游客)

	imgType := c.PostForm("type")
	if imgType == "" {
		imgType = model.ImageTypeProduct
	}

	// 游客仅允许海报与成品;其余类型需登录管理员
	if claims == nil {
		if imgType != model.ImageTypePoster && imgType != model.ImageTypeProduct {
			response.Forbidden(c, "请先登录后上传")
			return
		}
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.ParamError(c, "请选择要上传的图片")
		return
	}
	defer file.Close()

	userId, isAdmin := 0, false
	if claims != nil {
		userId = claims.UserId
		isAdmin = claims.Role == model.RoleAdmin
	}

	img, err := service.UploadImage(config.DB, service.UploadParam{
		UserId:      userId,
		ImgType:     imgType,
		IsAdmin:     isAdmin,
		Ip:          c.ClientIP(),
		Filename:    header.Filename,
		ContentType: header.Header.Get("Content-Type"),
		Size:        header.Size,
		Reader:      file,
	})
	if err != nil {
		replyServiceError(c, err)
		return
	}
	response.OK(c, gin.H{
		"id":         img.Id,
		"url":        img.Url,
		"filename":   img.Filename,
		"type":       img.Type,
		"createTime": img.CreateTime,
	})
}

// ListProducts 我的成品分页。
func ListProducts(c *gin.Context) {
	claims, _ := middleware.ClaimsOf(c)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "12"))
	list, total, err := service.UserProducts(config.DB, claims.UserId, page, size)
	if err != nil {
		response.ServerError(c, "查询失败")
		return
	}
	response.OK(c, gin.H{"list": list, "total": total, "page": page, "size": size})
}

// DeleteProduct 删除本人成品。
func DeleteProduct(c *gin.Context) {
	claims, _ := middleware.ClaimsOf(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.ParamError(c, "参数错误")
		return
	}
	if err := service.DeleteUserProduct(config.DB, claims.UserId, id); err != nil {
		replyServiceError(c, err)
		return
	}
	response.OKMessage(c, "已删除", nil)
}
