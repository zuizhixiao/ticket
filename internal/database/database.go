// Package database 负责 MySQL 连接与表结构自动迁移。
package database

import (
	"ticket/internal/config"
	"ticket/internal/model"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

// Init 建立连接并迁移表结构(表名统一 ticket_ 前缀)。
func Init(m config.Mysql) (*gorm.DB, error) {
	var (
		db  *gorm.DB
		err error
	)
	opts := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{TablePrefix: "ticket_", SingularTable: true},
		Logger:         logger.Default.LogMode(logger.Warn),
	}
	if m.Driver == "sqlite" {
		// 本地开发/测试:纯 Go SQLite,无需 MySQL。
		db, err = gorm.Open(sqlite.Open(m.Path), opts)
	} else {
		db, err = gorm.Open(mysql.Open(m.DSN()), opts)
	}
	if err != nil {
		return nil, err
	}
	if m.Driver == "sqlite" {
		// glebarez/sqlite 对显式 int(11) 列不会生成自增主键,这里用原生 DDL 建表,
		// 保证 id 为 INTEGER PRIMARY KEY AUTOINCREMENT(gorm 写入时省略零值 id 由库自增)。
		if err := ensureSQLiteSchema(db); err != nil {
			return nil, err
		}
	}
	if err := db.AutoMigrate(&model.User{}, &model.Image{}, &model.Template{}); err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	return db, nil
}

const sqliteDDL = `
CREATE TABLE IF NOT EXISTS "ticket_user" (
  "id" integer PRIMARY KEY AUTOINCREMENT,
  "nickname" varchar(255) NOT NULL,
  "password" varchar(100) NOT NULL,
  "avatar" varchar(500),
  "openid" varchar(64) DEFAULT '',
  "role" integer DEFAULT 0,
  "status" integer NOT NULL DEFAULT 1,
  "lastLoginTime" bigint,
  "createTime" bigint NOT NULL,
  "updateTime" bigint
);
CREATE TABLE IF NOT EXISTS "ticket_image" (
  "id" integer PRIMARY KEY AUTOINCREMENT,
  "userId" integer,
  "type" varchar(20),
  "filename" varchar(200),
  "url" varchar(500),
  "thumbUrl" varchar(500) DEFAULT '',
  "ip" varchar(64),
  "createTime" bigint
);
CREATE TABLE IF NOT EXISTS "ticket_template" (
  "id" integer PRIMARY KEY AUTOINCREMENT,
  "userId" integer NOT NULL DEFAULT 0,
  "name" varchar(100) DEFAULT '',
  "url" varchar(500) NOT NULL,
  "titleColor" varchar(20) DEFAULT '#ffffff',
  "textColor" varchar(20) DEFAULT '#ffffff',
  "status" integer NOT NULL DEFAULT 1,
  "createTime" integer NOT NULL
);
`

func ensureSQLiteSchema(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	_, err = sqlDB.Exec(sqliteDDL)
	return err
}
