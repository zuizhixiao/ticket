package model

import (
	"time"

	"gorm.io/gorm"
)

type Image struct {
	Id         int    `gorm:"column:id;type:int(11);primary_key;AUTO_INCREMENT" json:"id"`
	UserId     int    `gorm:"column:userId;type:int(11)" json:"userId"`          // 用户id
	Type       string `gorm:"column:type;type:varchar(20)" json:"type"`          // product 成品 poster 海报
	Filename   string `gorm:"column:filename;type:varchar(100)" json:"filename"` // 文件名
	Url        string `gorm:"column:url;type:varchar(255)" json:"url"`           // 文件地址
	Ip         string `gorm:"column:ip;type:varchar(20)" json:"ip"`              // 访问ip
	CreateTime int64  `gorm:"column:createTime;type:int(11)" json:"createTime"`  // 创建时间
}

func (m *Image) TableName() string {
	return "image"
}

func (m *Image) Create(Db *gorm.DB) error {
	m.CreateTime = time.Now().Unix()
	err := Db.Model(&m).Create(&m).Error
	return err
}

func (m *Image) Update(Db *gorm.DB, field ...string) error {
	sql := Db.Model(&m)
	if len(field) > 0 {
		sql = sql.Select(field)
	}
	err := sql.Where("id", m.Id).Updates(m).Error
	return err
}

func (m *Image) GetInfo(Db *gorm.DB) error {
	sql := Db.Model(m).Where("id = ? ", m.Id)
	err := sql.First(&m).Error
	return err
}

func GetImageList(Db *gorm.DB, page, num int) ([]Image, error) {
	var list []Image
	sql := Db.Model(Image{})
	if page > 0 && num > 0 {
		sql = sql.Limit(num).Offset((page - 1) * num)
	}
	err := sql.Order("id desc").Find(&list).Error
	return list, err
}

func GetImageCount(Db *gorm.DB) (int64, error) {
	var count int64
	sql := Db.Model(Image{})
	err := sql.Count(&count).Error
	return count, err
}

// GetImageListByUserAndType 根据用户ID和类型获取图片列表
func GetImageListByUserAndType(Db *gorm.DB, userId int, imgType string, page, num int) ([]Image, error) {
	var list []Image
	sql := Db.Model(Image{}).Where("userId = ? AND type = ?", userId, imgType)
	if page > 0 && num > 0 {
		sql = sql.Limit(num).Offset((page - 1) * num)
	}
	err := sql.Order("id desc").Find(&list).Error
	return list, err
}

// GetImageCountByUserAndType 根据用户ID和类型获取图片数量
func GetImageCountByUserAndType(Db *gorm.DB, userId int, imgType string) (int64, error) {
	var count int64
	err := Db.Model(Image{}).Where("userId = ? AND type = ?", userId, imgType).Count(&count).Error
	return count, err
}
