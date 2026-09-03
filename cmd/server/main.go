// Command server 是电影纪念票根服务的入口。
package main

import (
	"fmt"
	"log"

	"ticket/internal/config"
	"ticket/internal/database"
	"ticket/internal/router"
	"ticket/internal/service"
	"ticket/internal/storage"
)

func main() {
	if err := config.Load(); err != nil {
		log.Fatalf("load config: %v", err)
	}

	db, err := database.Init(config.C.Mysql)
	if err != nil {
		log.Fatalf("init mysql: %v", err)
	}
	config.DB = db

	sc := config.C.Storage
	store, err := storage.New(sc.Type, sc.AccessKeyID, sc.AccessKeySecret, sc.Endpoint, sc.Bucket)
	if err != nil {
		log.Fatalf("init storage: %v", err)
	}
	storage.Default = store

	if n := config.C.Admin.BootstrapNicknames; n != "" {
		if err := service.BootstrapAdmins(db, n); err != nil {
			log.Printf("bootstrap admins: %v", err)
		}
	}

	fmt.Println("ticket server starting at :" + fmt.Sprint(config.C.Server.Port))
	if err := router.Run(config.C); err != nil {
		log.Fatalf("run server: %v", err)
	}
}
