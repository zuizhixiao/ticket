package repository

import (
	"ticket/internal/model"

	"gorm.io/gorm"
)

// ---------- template ----------

// ListPublicTemplates 编辑器用:系统模板且上架。
func ListPublicTemplates(db *gorm.DB) ([]model.Template, error) {
	var list []model.Template
	err := db.Where("userId = 0 AND status = ?", model.StatusNormal).
		Order("id asc").Find(&list).Error
	return list, err
}

// AdminListTemplates 管理后台:status<=0 表示全部。
func AdminListTemplates(db *gorm.DB, status int) ([]model.Template, error) {
	var list []model.Template
	q := db.Model(&model.Template{})
	if status > 0 {
		q = q.Where("status = ?", status)
	}
	err := q.Order("id desc").Find(&list).Error
	return list, err
}

func GetTemplateByID(db *gorm.DB, id int) (*model.Template, error) {
	var t model.Template
	err := db.Where("id = ?", id).First(&t).Error
	return &t, err
}

func CreateTemplate(db *gorm.DB, t *model.Template) error {
	return db.Create(t).Error
}

func UpdateTemplateFields(db *gorm.DB, id int, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	return db.Model(&model.Template{}).Where("id = ?", id).Updates(fields).Error
}
