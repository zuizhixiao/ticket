package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"math/rand"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/image/bmp"
	"golang.org/x/image/draw"
	"golang.org/x/image/webp"

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
// 成品/海报会额外生成一份压缩图(thumb),供后台列表展示;详情仍用原图。
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
		Ip:       p.Ip,
	}

	// 成品/海报:生成压缩图(尽力而为,失败不影响原图)
	if p.ImgType == model.ImageTypeProduct || p.ImgType == model.ImageTypePoster {
		thumb, tErr := makeThumb(buf, thumbMaxEdge)
		if tErr == nil {
			thumbObj := fmt.Sprintf("%s/%s/%d%03d_t.jpg", p.ImgType, date, time.Now().UnixMilli(), rand.Intn(1000))
			if thumbURL, pErr := storage.Default.Put(context.Background(), thumbObj, bytes.NewReader(thumb), int64(len(thumb)), "image/jpeg"); pErr == nil {
				img.ThumbUrl = thumbURL
			}
		}
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

// deleteRemoteFiles 按 URL 尽力删除原图与压缩图(失败不阻断记录删除)。
func deleteRemoteFiles(img *model.Image) {
	if storage.Default == nil {
		return
	}
	if img.Url != "" {
		_ = storage.Default.DeleteByURL(context.Background(), img.Url)
	}
	if img.ThumbUrl != "" {
		_ = storage.Default.DeleteByURL(context.Background(), img.ThumbUrl)
	}
}

// DeleteUserProduct 删除本人成品(先删存储,再删记录)。
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
	deleteRemoteFiles(img)
	return repository.DeleteImageByID(db, imageId)
}

// AdminListImages 管理后台按类型分页查看全部图片。
func AdminListImages(db *gorm.DB, imgType string, page, size int) ([]repository.AdminImageRow, int64, error) {
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
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

// AdminDeleteImage 管理员删除任意图片(按 URL 删除原图与压缩图,再删记录)。
func AdminDeleteImage(db *gorm.DB, imageId int) error {
	img, err := repository.GetImageByID(db, imageId)
	if err != nil {
		if repository.IsNotFound(err) {
			return ErrImageNotFound
		}
		return err
	}
	deleteRemoteFiles(img)
	return repository.DeleteImageByID(db, imageId)
}

// thumbMaxEdge 压缩图最长边(px)。
const thumbMaxEdge = 400

// decodeAny 依次尝试 jpeg/png/gif/webp/bmp 解码。
func decodeAny(b []byte) (image.Image, error) {
	readers := []func() (image.Image, error){
		func() (image.Image, error) { return jpeg.Decode(bytes.NewReader(b)) },
		func() (image.Image, error) { return png.Decode(bytes.NewReader(b)) },
		func() (image.Image, error) { return gif.Decode(bytes.NewReader(b)) },
		func() (image.Image, error) { return webp.Decode(bytes.NewReader(b)) },
		func() (image.Image, error) { return bmp.Decode(bytes.NewReader(b)) },
	}
	var lastErr error
	for _, fn := range readers {
		if img, err := fn(); err == nil {
			return img, nil
		} else {
			lastErr = err
		}
	}
	return nil, lastErr
}

// makeThumb 等比压缩为 JPEG(最长边 thumbMaxEdge;原图更小则仅重编码降体积)。
func makeThumb(b []byte, maxEdge int) ([]byte, error) {
	src, err := decodeAny(b)
	if err != nil {
		return nil, err
	}
	bounds := src.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	if w <= 0 || h <= 0 {
		return nil, fmt.Errorf("invalid image size")
	}

	dw, dh := w, h
	if w > maxEdge || h > maxEdge {
		scale := float64(maxEdge) / float64(max(w, h))
		dw = max(1, int(float64(w)*scale))
		dh = max(1, int(float64(h)*scale))
	}

	dst := image.NewRGBA(image.Rect(0, 0, dw, dh))
	draw.ApproxBiLinear.Scale(dst, dst.Bounds(), src, bounds, draw.Over, nil)

	var out bytes.Buffer
	if err := jpeg.Encode(&out, dst, &jpeg.Options{Quality: 80}); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

// MakeThumbBytes 对外:由原图字节生成压缩图(最长边 thumbMaxEdge)。
func MakeThumbBytes(b []byte) ([]byte, error) {
	return makeThumb(b, thumbMaxEdge)
}

// BackfillImageThumb 为旧图补压缩图:下载原图 → 生成 → 上传 → 回填 thumbUrl。
func BackfillImageThumb(db *gorm.DB, img *model.Image) error {
	if img.ThumbUrl != "" || img.Url == "" {
		return nil
	}
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(img.Url)
	if err != nil {
		return fmt.Errorf("下载原图失败: %w", err)
	}
	defer resp.Body.Close()
	buf, err := io.ReadAll(io.LimitReader(resp.Body, maxUploadBytes+1))
	if err != nil {
		return fmt.Errorf("读取原图失败: %w", err)
	}
	if len(buf) > maxUploadBytes {
		return fmt.Errorf("原图超过 10MB,跳过")
	}
	thumb, err := MakeThumbBytes(buf)
	if err != nil {
		return fmt.Errorf("压缩失败: %w", err)
	}
	if storage.Default == nil {
		return fmt.Errorf("对象存储未初始化")
	}
	thumbObj := fmt.Sprintf("%s/%s/%d%03d_t.jpg", img.Type,
		time.Now().Format("20060102"), time.Now().UnixMilli(), rand.Intn(1000))
	thumbURL, err := storage.Default.Put(context.Background(), thumbObj,
		bytes.NewReader(thumb), int64(len(thumb)), "image/jpeg")
	if err != nil {
		return fmt.Errorf("上传压缩图失败: %w", err)
	}
	return repository.UpdateImageFields(db, img.Id, map[string]any{"thumbUrl": thumbURL})
}
