package initialize

import (
	"fmt"
	"os"
	"ticket/config"
	"ticket/model"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func Mysql() {
	conf := config.GVA_CONFIG.Mysql
	link := conf.Username + ":" + conf.Password + "@(" + conf.Path + ")/" + conf.Dbname + "?" + conf.Config
	if db, err := gorm.Open(mysql.Open(link), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, // 使用单数表名，启用该选项后，`User` 表将是`user`
		},
		Logger: logger.Default.LogMode(logger.Info),
	}); err != nil {
		fmt.Println("mysql connect failed", err.Error())
		os.Exit(0)
	} else {
		config.GVA_DB = db
		// 自动迁移用户表
		db.AutoMigrate(&model.User{})
		sqlDb, _ := db.DB()
		sqlDb.SetMaxIdleConns(10)
		sqlDb.SetMaxOpenConns(100)
	}

}
