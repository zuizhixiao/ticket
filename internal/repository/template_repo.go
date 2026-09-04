package repository

import (
	"database/sql"

	"ticket/internal/model"

	"gorm.io/gorm"
)

// ---------- template ----------

// ListPublicTemplates 编辑器用:系统模板且上架(按 sort 升序,同值按新→旧)。
func ListPublicTemplates(db *gorm.DB) ([]model.Template, error) {
	var list []model.Template
	err := db.Where("userId = 0 AND status = ?", model.StatusNormal).
		Order("sort asc, id desc").Find(&list).Error
	return list, err
}

// AdminListTemplates 管理后台:status<=0 表示全部,按 sort 升序。
func AdminListTemplates(db *gorm.DB, status int) ([]model.Template, error) {
	var list []model.Template
	q := db.Model(&model.Template{})
	if status > 0 {
		q = q.Where("status = ?", status)
	}
	err := q.Order("sort asc, id desc").Find(&list).Error
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

// DeleteTemplate 硬删除模板记录。
func DeleteTemplate(db *gorm.DB, id int) error {
	return db.Delete(&model.Template{}, id).Error
}

// MinTemplateSort 当前最小 sort(新建默认置顶用)。
func MinTemplateSort(db *gorm.DB) (int, error) {
	var m sql.NullInt64
	err := db.Model(&model.Template{}).Select("MIN(sort)").Scan(&m).Error
	if err != nil {
		return 0, err
	}
	if !m.Valid {
		return 0, nil
	}
	return int(m.Int64), nil
}
