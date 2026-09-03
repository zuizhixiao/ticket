// Package config 负责配置加载与全局共享实例。
// 优先级:默认值 < config.yaml(dev 读取,可选) < 环境变量(prod 必填,env 恒可覆盖)。
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
	"gorm.io/gorm"
)

// C 全局配置(Load 后可用)。
var C *Config

// DB 全局 GORM 实例(database.Init 后可用)。
var DB *gorm.DB

// Config 应用配置结构(config.yaml / config.example.yaml 同构)。
type Config struct {
	Server  Server  `mapstructure:"server"`
	Storage Storage `mapstructure:"storage"`
	Mysql   Mysql   `mapstructure:"mysql"`
	Wechat  Wechat  `mapstructure:"wechat"`
	Jwt     Jwt     `mapstructure:"jwt"`
	Captcha Captcha `mapstructure:"captcha"`
	Admin   Admin   `mapstructure:"admin"`
}

type Server struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type Storage struct {
	Type            string `mapstructure:"type"` // mos | cos
	AccessKeyID     string `mapstructure:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret"`
	Endpoint        string `mapstructure:"endpoint"`
	Bucket          string `mapstructure:"bucket"`
}

type Mysql struct {
	Driver   string `mapstructure:"driver"` // mysql(默认) | sqlite(本地开发,Path=db文件)
	Path     string `mapstructure:"path"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	Dbname   string `mapstructure:"dbname"`
	Config   string `mapstructure:"config"`
}

type Wechat struct {
	AppID     string `mapstructure:"app_id"`
	AppSecret string `mapstructure:"app_secret"`
	Token     string `mapstructure:"token"`
}

type Jwt struct {
	Secret        string `mapstructure:"secret"`
	ExpireSeconds int    `mapstructure:"expire_seconds"`
}

type Captcha struct {
	ExpireMinutes int `mapstructure:"expire_minutes"`
}

type Admin struct {
	BootstrapNicknames string `mapstructure:"bootstrap_nicknames"`
}

// IsProd 生产模式(仅读环境变量)。
func (c *Config) IsProd() bool {
	return strings.EqualFold(os.Getenv("RUN_MODE"), "prod")
}

// Load 初始化全局配置。可重复调用(幂等)。
func Load() error {
	cfg := defaults()
	mergeYAML(cfg)
	mergeEnv(cfg)
	C = cfg
	return nil
}

func defaults() *Config {
	return &Config{
		Server: Server{Port: 8080, Mode: "debug"},
		Jwt:    Jwt{Secret: "ticket", ExpireSeconds: 86400},
		Captcha: Captcha{ExpireMinutes: 10},
	}
}

// mergeYAML 读取 config.yaml(忽略不存在),仅用于非 prod 或作为基础值。
func mergeYAML(cfg *Config) {
	v := viper.New()
	v.SetConfigFile("config.yaml")
	if err := v.ReadInConfig(); err != nil {
		return
	}
	_ = v.Unmarshal(cfg)
}

// mergeEnv 环境变量非空时覆盖配置(prod 模式下必需)。
func mergeEnv(cfg *Config) {
	setInt(&cfg.Server.Port, "SERVER_PORT")
	setStr(&cfg.Server.Mode, "SERVER_MODE")

	setStr(&cfg.Storage.Type, "STORAGE_TYPE")
	setStr(&cfg.Storage.AccessKeyID, "STORAGE_ACCESS_KEY_ID")
	setStr(&cfg.Storage.AccessKeySecret, "STORAGE_ACCESS_KEY_SECRET")
	setStr(&cfg.Storage.Endpoint, "STORAGE_ENDPOINT")
	setStr(&cfg.Storage.Bucket, "STORAGE_BUCKET")

	setStr(&cfg.Mysql.Driver, "MYSQL_DRIVER")
	setStr(&cfg.Mysql.Path, "MYSQL_PATH")
	setStr(&cfg.Mysql.Username, "MYSQL_USERNAME")
	setStr(&cfg.Mysql.Password, "MYSQL_PASSWORD")
	setStr(&cfg.Mysql.Dbname, "MYSQL_DBNAME")
	setStr(&cfg.Mysql.Config, "MYSQL_CONFIG")

	setStr(&cfg.Wechat.AppID, "WECHAT_APP_ID")
	setStr(&cfg.Wechat.AppSecret, "WECHAT_APP_SECRET")
	setStr(&cfg.Wechat.Token, "WECHAT_TOKEN")

	setStr(&cfg.Jwt.Secret, "JWT_SECRET")
	setInt(&cfg.Jwt.ExpireSeconds, "JWT_EXPIRE_SECONDS")

	setInt(&cfg.Captcha.ExpireMinutes, "CAPTCHA_EXPIRE_MINUTES")

	setStr(&cfg.Admin.BootstrapNicknames, "ADMIN_BOOTSTRAP_NICKNAMES")
}

func setStr(dst *string, key string) {
	if v := os.Getenv(key); v != "" {
		*dst = v
	}
}

func setInt(dst *int, key string) {
	if v := os.Getenv(key); v != "" {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err == nil {
			*dst = n
		}
	}
}

// DSN 生成 MySQL 连接串。
func (m Mysql) DSN() string {
	if m.Config == "" {
		m.Config = "charset=utf8mb4&parseTime=True&loc=Local"
	}
	return fmt.Sprintf("%s:%s@(%s)/%s?%s", m.Username, m.Password, m.Path, m.Dbname, m.Config)
}
