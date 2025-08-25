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

	conf := config.GVA_CONFIG

	config.CosClient = cos.SetUp(
		conf.CosAccessKeyId,
		conf.CosAccessKeySecret,
		conf.CosEndpoint,
		conf.CosBucket)
	initialize.RunServer()
}
