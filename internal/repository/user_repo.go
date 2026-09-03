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

// IsNotFound 判断是否为记录不存在。
func IsNotFound(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}
