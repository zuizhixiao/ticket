package model

import (
	"crypto/md5"
	"encoding/hex"

	"gorm.io/gorm"
)

// User 用户表
type User struct {
	Id            int    `gorm:"column:id;type:int(11);primaryKey;autoIncrement" json:"id"`
	Nickname      string `gorm:"column:nickname;type:varchar(255)" json:"nickname"`
	Avatar        string `gorm:"column:avatar;type:varchar(500)" json:"avatar"`
	Password      string `gorm:"column:password;type:char(32);not null" json:"-"`
	Status        int    `gorm:"column:status;type:int(11);not null;default:1" json:"status"`
	LastLoginTime *int64 `gorm:"column:lastLoginTime;type:bigint(20)" json:"lastLoginTime"`
	CreateTime    int64  `gorm:"column:createTime;type:bigint(20);not null" json:"createTime"`
	UpdateTime    *int64 `gorm:"column:updateTime;type:bigint(20)" json:"updateTime"`
}

func (m *User) TableName() string {
	return "user"
}

func (m *User) Create(db *gorm.DB) error {
	return db.Model(m).Create(m).Error
}

func (m *User) Update(db *gorm.DB, fields ...string) error {
	sql := db.Model(m)
	if len(fields) > 0 {
		sql = sql.Select(fields)
	}
	return sql.Where("id = ?", m.Id).Updates(m).Error
}

func (m *User) GetByNickname(db *gorm.DB) error {
	return db.Model(m).Where("nickname = ?", m.Nickname).First(m).Error
}

// SetPassword MD5加密并设置密码
func (m *User) SetPassword(password string) error {
	hash := md5.Sum([]byte(password))
	m.Password = hex.EncodeToString(hash[:])
	return nil
}

// CheckPassword 验证密码
func (m *User) CheckPassword(password string) bool {
	hash := md5.Sum([]byte(password))
	return m.Password == hex.EncodeToString(hash[:])
}

// GetUserByNickname 根据昵称获取用户
func GetUserByNickname(db *gorm.DB, nickname string) (*User, error) {
	var user User
	err := db.Where("nickname = ?", nickname).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserById 根据ID获取用户
func GetUserById(db *gorm.DB, id int) (*User, error) {
	var user User
	err := db.Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
