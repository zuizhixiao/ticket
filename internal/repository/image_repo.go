package repository

import (
	"ticket/internal/model"

	"gorm.io/gorm"
)

// ---------- image ----------

func CreateImage(db *gorm.DB, img *model.Image) error {
	return db.Create(img).Error
}

// PageImageByUserAndType 分页查询某用户某类型图片(新→旧)。
func PageImageByUserAndType(db *gorm.DB, userId int, imgType string, page, size int) ([]model.Image, error) {
	var list []model.Image
	q := db.Model(&model.Image{}).Where("userId = ? AND type = ?", userId, imgType)
	if page > 0 && size > 0 {
		q = q.Order("id desc").Limit(size).Offset((page - 1) * size)
	} else {
		q = q.Order("id desc")
	}
	err := q.Find(&list).Error
	return list, err
}

func CountImageByUserAndType(db *gorm.DB, userId int, imgType string) (int64, error) {
	var n int64
	err := db.Model(&model.Image{}).Where("userId = ? AND type = ?", userId, imgType).Count(&n).Error
	return n, err
}

// GetOwnedImage 获取属于该用户的图片(用于删除权限校验)。
func GetOwnedImage(db *gorm.DB, userId, id int) (*model.Image, error) {
	var img model.Image
	err := db.Where("id = ? AND userId = ?", id, userId).First(&img).Error
	return &img, err
}

// AdminImageRow 管理后台列表行:图片 + 上传者昵称。
type AdminImageRow struct {
	model.Image
	Nickname string `json:"nickname"`
}

// AdminPageImagesByType 管理后台:按类型分页查全部图片(左连用户取昵称,游客显示"游客")。
func AdminPageImagesByType(db *gorm.DB, imgType string, page, size int) ([]AdminImageRow, error) {
	var rows []AdminImageRow
	q := db.Table("image").
		Select("image.*, IFNULL(u.nickname, '游客') AS nickname").
		Joins("LEFT JOIN user AS u ON u.id = image.userId").
		Where("image.type = ?", imgType).
		Order("image.id DESC")
	if page > 0 && size > 0 {
		q = q.Limit(size).Offset((page - 1) * size)
	}
	err := q.Scan(&rows).Error
	return rows, err
}

// CountImageByType 按类型统计全部图片。
func CountImageByType(db *gorm.DB, imgType string) (int64, error) {
	var n int64
	err := db.Model(&model.Image{}).Where("type = ?", imgType).Count(&n).Error
	return n, err
}

// GetImageByID 按 ID 取任意图片(管理删除用)。
func GetImageByID(db *gorm.DB, id int) (*model.Image, error) {
	var img model.Image
	err := db.Where("id = ?", id).First(&img).Error
	return &img, err
}

func DeleteImageByID(db *gorm.DB, id int) error {
	return db.Delete(&model.Image{}, id).Error
}
