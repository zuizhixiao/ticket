package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"

	"ticket/internal/model"
	"ticket/internal/repository"
	"ticket/internal/storage"
)

var (
	ErrImageType      = errors.New("不支持的图片类型")
	ErrImageExt       = errors.New("仅支持 JPG/PNG/GIF/BMP/WebP 图片")
	ErrImageTooLarge  = errors.New("图片大小超出限制(10MB)")
	ErrImageNotFound  = errors.New("图片不存在")
	ErrNotYourImage   = errors.New("无权操作该图片")
	ErrStorageUpload  = errors.New("图片上传存储失败")
	ErrStorageDelete  = errors.New("图片删除存储失败")
	ErrTemplateDenied = errors.New("仅管理员可上传模板图片")
)

const maxUploadBytes = 10 << 20 // 10MB

var allowedUploadTypes = map[string]bool{
	model.ImageTypePoster:   true,
	model.ImageTypeProduct:  true,
	model.ImageTypeAvatar:   true,
	model.ImageTypeTemplate: true,
}

var allowedExts = map[string]bool{
	".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true, ".webp": true,
}

// UploadParam 上传入参(文件流由调用方持有)。
type UploadParam struct {
	UserId      int
	ImgType     string // poster/product/avatar/template
	IsAdmin     bool
	Ip          string
	Filename    string
	ContentType string
	Size        int64
	Reader      io.Reader
}

// UploadImage 上传对象存储并记录到 image 表。
func UploadImage(db *gorm.DB, p UploadParam) (*model.Image, error) {
	if !allowedUploadTypes[p.ImgType] {
		return nil, ErrImageType
	}
	if p.ImgType == model.ImageTypeTemplate && !p.IsAdmin {
		return nil, ErrTemplateDenied
	}
	ext := strings.ToLower(filepath.Ext(p.Filename))
	if !allowedExts[ext] {
		return nil, ErrImageExt
	}
	contentType := p.ContentType
	if contentType == "" || !strings.HasPrefix(contentType, "image/") {
		return nil, ErrImageType
	}

	// 读入内存(≤10MB),两种 SDK 均需可测长度/字节流。
	buf, err := io.ReadAll(io.LimitReader(p.Reader, maxUploadBytes+1))
	if err != nil {
		return nil, ErrStorageUpload
	}
	if len(buf) > maxUploadBytes {
		return nil, ErrImageTooLarge
	}

	date := time.Now().Format("20060102")
	object := fmt.Sprintf("%s/%s/%d%03d%s", p.ImgType, date, time.Now().UnixMilli(), rand.Intn(1000), ext)

	if storage.Default == nil {
		return nil, ErrStorageUpload
	}
	url, err := storage.Default.Put(context.Background(), object, bytes.NewReader(buf), int64(len(buf)), contentType)
	if err != nil {
		return nil, ErrStorageUpload
	}

	img := &model.Image{
		UserId:   p.UserId,
		Type:     p.ImgType,
		Filename: p.Filename,
		Url:      url,
		Object:   object,
		Ip:       p.Ip,
	}
	if err := repository.CreateImage(db, img); err != nil {
		// 记录失败不回滚存储(尽力而为),返回成功 URL 至少可用
		return img, nil
	}
	return img, nil
}

// UserProducts 当前用户成品分页。
func UserProducts(db *gorm.DB, userId, page, size int) (list []model.Image, total int64, err error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 12
	}
	list, err = repository.PageImageByUserAndType(db, userId, model.ImageTypeProduct, page, size)
	if err != nil {
		return nil, 0, err
	}
	total, err = repository.CountImageByUserAndType(db, userId, model.ImageTypeProduct)
	return list, total, err
}

// DeleteUserProduct 删除本人成品(先删存储,再删记录;存储失败不阻断记录删除)。
func DeleteUserProduct(db *gorm.DB, userId, imageId int) error {
	img, err := repository.GetOwnedImage(db, userId, imageId)
	if err != nil {
		if repository.IsNotFound(err) {
			return ErrImageNotFound
		}
		return err
	}
	if img.Type != model.ImageTypeProduct {
		return ErrNotYourImage
	}
	if storage.Default != nil && img.Object != "" {
		_ = storage.Default.Delete(context.Background(), img.Object)
	}
	return repository.DeleteImageByID(db, imageId)
}

// AdminListImages 管理后台按类型分页查看全部图片。
func AdminListImages(db *gorm.DB, imgType string, page, size int) ([]repository.AdminImageRow, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 24
	}
	list, err := repository.AdminPageImagesByType(db, imgType, page, size)
	if err != nil {
		return nil, 0, err
	}
	total, err := repository.CountImageByType(db, imgType)
	if err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

// AdminDeleteImage 管理员删除任意图片(先删存储对象,再删记录)。
func AdminDeleteImage(db *gorm.DB, imageId int) error {
	img, err := repository.GetImageByID(db, imageId)
	if err != nil {
		if repository.IsNotFound(err) {
			return ErrImageNotFound
		}
		return err
	}
	if storage.Default != nil && img.Object != "" {
		_ = storage.Default.Delete(context.Background(), img.Object)
	}
	return repository.DeleteImageByID(db, imageId)
}
