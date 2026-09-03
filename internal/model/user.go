// Package model 定义 GORM 数据模型(单数表名)。
package model

import (
	"time"

	"gorm.io/gorm"
)

// Role 常量
const (
	RoleUser  = 0
	RoleAdmin = 1
)

// Status 常量
const (
	StatusNormal = 1 // 正常 / 上架
	StatusBanned = 0 // 禁用
	StatusDelete = 2 // 软删除 / 下架
)

// User 用户表
type User struct {
	Id            int    `gorm:"column:id;type:int(11);primaryKey;autoIncrement" json:"id"`
	Nickname      string `gorm:"column:nickname;type:varchar(255);not null" json:"nickname"`
	Password      string `gorm:"column:password;type:varchar(100);not null" json:"-"`
	Avatar        string `gorm:"column:avatar;type:varchar(500)" json:"avatar"`
	Openid        string `gorm:"column:openid;type:varchar(64);default:'';index" json:"openid"`
	Role          int    `gorm:"column:role;type:int(11);default:0" json:"role"`
	Status        int    `gorm:"column:status;type:int(11);not null;default:1" json:"status"`
	LastLoginTime *int64 `gorm:"column:lastLoginTime;type:bigint(20)" json:"lastLoginTime"`
	CreateTime    int64  `gorm:"column:createTime;type:bigint(20);not null" json:"createTime"`
	UpdateTime    *int64 `gorm:"column:updateTime;type:bigint(20)" json:"updateTime"`
}

func (User) TableName() string { return "user" }

func (u *User) BeforeCreate(*gorm.DB) error {
	if u.CreateTime == 0 {
		u.CreateTime = time.Now().Unix()
	}
	return nil
}

// IsAdmin 是否管理员
func (u *User) IsAdmin() bool { return u.Role == RoleAdmin }

// Profile 对外可见的用户信息
func (u *User) Profile() map[string]any {
	return map[string]any{
		"id":         u.Id,
		"nickname":   u.Nickname,
		"avatar":     u.Avatar,
		"role":       u.Role,
		"openid":     u.Openid,
		"createTime": u.CreateTime,
	}
}
