package main

import (
	"ticket/config"
	"ticket/initialize"
	"ticket/package/cos"
)

func main() {
	config.Init()

	//初始化mysql
	initialize.Mysql()

	conf := config.GVA_CONFIG.Cos

	config.CosClient = cos.SetUp(conf.AccessKeyId, conf.AccessKeySecret, conf.Endpoint, conf.Bucket)

	initialize.RunServer()
}
