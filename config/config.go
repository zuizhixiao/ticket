package config

import (
	"fmt"
	"os"
	"ticket/package/fileFactory"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var (
	FileFactoryClient fileFactory.FileFactory
	GVA_CONFIG        Params
	GVA_DB            *gorm.DB
)

type Params struct {
	File   File
	Mysql  Mysql
	Wechat Wechat
	Jwt    Jwt
}

type Jwt struct {
	Secret    string `mapstructure:"secret" json:"secret" yaml:"secret"`
	ExpireSec int    `mapstructure:"expire_time" json:"expire_time" yaml:"expire_time"`
}

type File struct {
	AccessKeyId     string
	AccessKeySecret string
	Endpoint        string
	Bucket          string
	Type            string
}

type Mysql struct {
	Path     string
	Username string
	Password string
	Dbname   string
	Config   string
}

type Wechat struct {
	AppID     string `mapstructure:"app_id" json:"app_id" yaml:"app_id"`
	AppSecret string `mapstructure:"app_secret" json:"app_secret" yaml:"app_secret"`
	Token     string
}

func Init() {
	runMode := os.Getenv("RUN_MODE")
	//若非生产环境，则读取配置文件
	if runMode != "prod" {
		//读取配置文件
		v := viper.New()
		v.SetConfigFile("config.yaml")
		err := v.ReadInConfig()
		if err != nil {
			panic(fmt.Errorf("Fatal error config file: %s \n", err))
		}
		v.WatchConfig()
		v.OnConfigChange(func(e fsnotify.Event) {
			fmt.Println("config file changed:", e.Name)
			if err := v.Unmarshal(&GVA_CONFIG); err != nil {
				fmt.Println(err)
			}
		})
		if err := v.Unmarshal(&GVA_CONFIG); err != nil {
			fmt.Println(err)
		}
	} else {
		GVA_CONFIG = Params{
			File: File{
				AccessKeyId:     os.Getenv("FILE_ACCESS_KEY_ID"),
				AccessKeySecret: os.Getenv("FILE_ACCESS_KEY_SECRET"),
				Endpoint:        os.Getenv("FILE_ENDPOINT"),
				Bucket:          os.Getenv("FILE_BUCKET"),
			},
			Mysql: Mysql{
				Path:     os.Getenv("MYSQL_PATH"),
				Username: os.Getenv("MYSQL_USERNAME"),
				Password: os.Getenv("MYSQL_PASSWORD"),
				Dbname:   os.Getenv("MYSQL_DBNAME"),
				Config:   os.Getenv("MYSQL_CONFIG"),
			},
			Wechat: Wechat{
				AppID:     os.Getenv("WECHAT_APP_ID"),
				AppSecret: os.Getenv("WECHAT_APP_SECRET"),
				Token:     os.Getenv("WECHAT_TOKEN"),
			},
			Jwt: Jwt{
				Secret:    getEnvDefault("JWT_SECRET", "ticket"),
				ExpireSec: 86400,
			},
		}
	}
}

func getEnvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
