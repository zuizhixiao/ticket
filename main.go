package main

import (
	"ticket/config"
	"ticket/initialize"
	"ticket/package/fileFactory"
)

func main() {
	config.Init()

	//初始化mysql
	initialize.Mysql()

	conf := config.GVA_CONFIG.File

	var err error
	config.FileFactoryClient, err = fileFactory.SetUp(conf.AccessKeyId, conf.AccessKeySecret, conf.Endpoint, conf.Bucket, conf.Type)
	if err != nil {
		panic(err)
	}

	initialize.RunServer()
}
