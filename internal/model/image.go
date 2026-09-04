package model

import (
	"time"

	"gorm.io/gorm"
)

// 图片类型常量
const (
	ImageTypePoster   = "poster"   // 海报(编辑器临时素材)
	ImageTypeProduct  = "product"  // 成品票根
	ImageTypeAvatar   = "avatar"   // 用户头像
	ImageTypeTemplate = "template" // 模板背景图(管理员)
)

// Image 图片记录(对象存储 URL + 归属)
type Image struct {
	Id         int    `gorm:"column:id;type:int(11);primaryKey;autoIncrement" json:"id"`
	UserId     int    `gorm:"column:userId;type:int(11);index" json:"userId"`
	Type       string `gorm:"column:type;type:varchar(20);index" json:"type"`
	Filename   string `gorm:"column:filename;type:varchar(200)" json:"filename"`
	Url        string `gorm:"column:url;type:varchar(500)" json:"url"`                  // 原图 URL
	ThumbUrl   string `gorm:"column:thumbUrl;type:varchar(500);default:''" json:"thumbUrl"` // 压缩图 URL(列表展示)
	Ip         string `gorm:"column:ip;type:varchar(64)" json:"ip"`
	CreateTime int64  `gorm:"column:createTime;type:bigint(20)" json:"createTime"`
}

func (Image) TableName() string { return "ticket_image" }

func (i *Image) BeforeCreate(*gorm.DB) error {
	if i.CreateTime == 0 {
		i.CreateTime = time.Now().Unix()
	}
	return nil
}
