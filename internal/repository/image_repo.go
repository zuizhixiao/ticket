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

func DeleteImageByID(db *gorm.DB, id int) error {
	return db.Delete(&model.Image{}, id).Error
}
