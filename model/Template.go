package model

import "gorm.io/gorm"

// 模板库
type Template struct {
	Id         int    `gorm:"column:id;type:int(11);primary_key;AUTO_INCREMENT" json:"id"`
	UserId     int    `gorm:"column:userId;type:int(11);default:0;NOT NULL" json:"userId"`    // 用户id 为0代表系统
	Url        string `gorm:"column:url;type:varchar(250);NOT NULL" json:"url"`               // 模板图片
	TitleColor string `gorm:"column:titleColor;type:varchar(20);NOT NULL" json:"titleColor"`  // 电影标题颜色
	TextColor  string `gorm:"column:textColor;type:varchar(20);NOT NULL" json:"textColor"`    // 电影详情颜色
	Status     int    `gorm:"column:status;type:tinyint(1);default:1;NOT NULL" json:"status"` // 1正常 2删除
	CreateTime int    `gorm:"column:createTime;type:int(11);NOT NULL" json:"createTime"`      // 创建时间
}

func (m *Template) TableName() string {
	return "template"
}

func (m *Template) Create(Db *gorm.DB) error {
	err := Db.Model(&m).Create(&m).Error
	return err
}

func (m *Template) Update(Db *gorm.DB, field ...string) error {
	sql := Db.Model(&m)
	if len(field) > 0 {
		sql = sql.Select(field)
	}
	err := sql.Where("id", m.Id).Updates(m).Error
	return err
}

func (m *Template) GetInfo(Db *gorm.DB) error {
	sql := Db.Model(m).Where("id = ? ", m.Id)
	err := sql.First(&m).Error
	return err
}

func GetTemplateList(Db *gorm.DB, userId, page, num int) ([]Template, error) {
	list := []Template{}
	sql := Db.Model(Template{}).Where("userId = ?", userId)
	if page > 0 && num > 0 {
		sql = sql.Limit(num).Offset((page - 1) * num)
	}
	err := sql.Order("id desc").Find(&list).Error
	return list, err
}

func GetTemplateCount(Db *gorm.DB, userId int) (int64, error) {
	var count int64
	sql := Db.Model(Template{}).Where("userId = ?", userId)
	err := sql.Count(&count).Error
	return count, err
}
