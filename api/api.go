package api

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"ticket/config"
	"ticket/model"

	"ticket/types/response"

	"github.com/gin-gonic/gin"
)

// SavePic 处理图片上传
func SavePic(c *gin.Context) {
	// 设置最大文件大小为10MB
	c.Request.ParseMultipartForm(10 << 20)

	// 获取上传的文件
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		response.Result(2, nil, "获取上传文件失败: "+err.Error(), c)
		return
	}
	defer file.Close()

	// 获取图片类型参数，默认为poster
	imageType := c.PostForm("type")
	if imageType == "" {
		imageType = "poster" // 默认为海报类型
	}

	// 验证文件类型
	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		response.Result(3, nil, "只支持图片文件上传", c)
		return
	}

	// 验证文件扩展名
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"}
	isValidExt := false
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			isValidExt = true
			break
		}
	}
	if !isValidExt {
		response.Result(3, nil, "不支持的图片格式，请上传JPG、PNG、GIF、BMP或WebP格式的图片", c)
		return
	}

	dateFormat := time.Unix(time.Now().Unix(), 0).Format("20060102")
	filename := fmt.Sprintf("images/ticket/%s/%s/%d%s", imageType, dateFormat, time.Now().Unix(), ext)
	bufferBytes, _ := io.ReadAll(file)

	imageUrl, err := config.CosClient.PutFileWithBody(filename, bufferBytes)
	if err != nil {
		response.Result(500, nil, err.Error(), c)
		return
	}

	image := model.Image{
		Type:     imageType,
		Filename: header.Filename,
		Url:      imageUrl,
		Ip:       c.ClientIP(),
	}
	image.Create(config.GVA_DB)

	// 返回成功响应
	response.Success(gin.H{
		"filename": filename,
		"size":     header.Size,
		"type":     contentType,
	}, c)
}

func GetTemplate(c *gin.Context) {
	list, err := model.GetTemplateList(config.GVA_DB, 0, 1, 10)
	if err != nil {
		response.Result(500, nil, err.Error(), c)
		return
	}

	//获取模板总数
	count, _ := model.GetTemplateCount(config.GVA_DB, 0)
	response.Success(gin.H{
		"list":  list,
		"count": count,
	}, c)
}

func GetTypeSettingTemplate(c *gin.Context) {
	userId := 1
	list, err := model.GetTemplateList(config.GVA_DB, userId, 1, 10)
	if err != nil {
		response.Result(500, nil, err.Error(), c)
		return
	}

	//获取模板总数
	count, _ := model.GetTemplateCount(config.GVA_DB, userId)
	response.Success(gin.H{
		"list":  list,
		"count": count,
	}, c)
}
