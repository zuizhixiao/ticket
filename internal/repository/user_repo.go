// Package repository 数据访问层(收口 SQL)。
package repository

import (
	"errors"

	"ticket/internal/model"

	"gorm.io/gorm"
)

// ---------- user ----------

func CreateUser(db *gorm.DB, u *model.User) error {
	return db.Create(u).Error
}

func GetUserByNickname(db *gorm.DB, nickname string) (*model.User, error) {
	var u model.User
	err := db.Where("nickname = ?", nickname).First(&u).Error
	return &u, err
}

func GetUserByID(db *gorm.DB, id int) (*model.User, error) {
	var u model.User
	err := db.Where("id = ?", id).First(&u).Error
	return &u, err
}

func GetUserByOpenid(db *gorm.DB, openid string) (*model.User, error) {
	var u model.User
	err := db.Where("openid = ? AND openid <> ''", openid).First(&u).Error
	return &u, err
}

func UpdateUserFields(db *gorm.DB, id int, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	return db.Model(&model.User{}).Where("id = ?", id).Updates(fields).Error
}

// ListUsers 分页查询用户(仅普通用户 role!=1;昵称模糊搜索)。
func ListUsers(db *gorm.DB, keyword string, page, size int) ([]model.User, error) {
	var list []model.User
	q := db.Model(&model.User{}).Where("role <> ?", model.RoleAdmin)
	if keyword != "" {
		q = q.Where("nickname LIKE ?", "%"+keyword+"%")
	}
	if page > 0 && size > 0 {
		q = q.Order("id DESC").Limit(size).Offset((page - 1) * size)
	} else {
		q = q.Order("id DESC")
	}
	err := q.Find(&list).Error
	return list, err
}

func CountUsers(db *gorm.DB, keyword string) (int64, error) {
	var n int64
	q := db.Model(&model.User{}).Where("role <> ?", model.RoleAdmin)
	if keyword != "" {
		q = q.Where("nickname LIKE ?", "%"+keyword+"%")
	}
	err := q.Count(&n).Error
	return n, err
}

// IsNotFound 判断是否为记录不存在。
func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
