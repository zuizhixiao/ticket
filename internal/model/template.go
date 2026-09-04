package model

import (
	"time"

	"gorm.io/gorm"
)

// Template 票根背景模板;UserId=0 表示系统模板
type Template struct {
	Id         int    `gorm:"column:id;type:int(11);primaryKey;autoIncrement" json:"id"`
	UserId     int    `gorm:"column:userId;type:int(11);default:0;not null" json:"userId"`
	Name       string `gorm:"column:name;type:varchar(100);default:''" json:"name"`
	Url        string `gorm:"column:url;type:varchar(500);not null" json:"url"`
	TitleColor string `gorm:"column:titleColor;type:varchar(20);default:'#ffffff'" json:"titleColor"`
	TextColor  string `gorm:"column:textColor;type:varchar(20);default:'#ffffff'" json:"textColor"`
	Status     int    `gorm:"column:status;type:tinyint(1);default:1;not null" json:"status"`
	Sort       int    `gorm:"column:sort;type:int(11);default:0;not null" json:"sort"` // 展示排序,越小越靠前
	CreateTime int    `gorm:"column:createTime;type:int(11);not null" json:"createTime"`
}

func (Template) TableName() string { return "ticket_template" }

func (t *Template) BeforeCreate(*gorm.DB) error {
	if t.CreateTime == 0 {
		t.CreateTime = int(time.Now().Unix())
	}
	return nil
}
