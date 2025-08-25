package config

import (
	"fmt"
	"os"
	"ticket/package/cos"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

var (
	CosClient  *cos.Cos
	GVA_CONFIG Params
	GVA_DB     *gorm.DB
)

type Params struct {
	CosAccessKeyId     string
	CosAccessKeySecret string
	CosEndpoint        string
	CosBucket          string
	SystemPort         string
	Mysql              Mysql
}

type Mysql struct {
	Path     string
	Username string
	Password string
	Dbname   string
	Config   string
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
			CosAccessKeyId:     os.Getenv("COS_ACCESS_KEY_ID"),
			CosAccessKeySecret: os.Getenv("COS_ACCESS_KEY_SECRET"),
			CosEndpoint:        os.Getenv("COS_ENDPOINT"),
			CosBucket:          os.Getenv("COS_BUCKET"),
			Mysql: Mysql{
				Path:     os.Getenv("MYSQL_PATH"),
				Username: os.Getenv("MYSQL_USERNAME"),
				Password: os.Getenv("MYSQL_PASSWORD"),
				Dbname:   os.Getenv("MYSQL_DBNAME"),
				Config:   os.Getenv("MYSQL_CONFIG"),
			},
		}

	}
}
