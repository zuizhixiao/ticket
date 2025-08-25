package initialize

import (
	"fmt"
	"io"
	"net/http"
	"ticket/router"
	_ "ticket/statik"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rakyll/statik/fs"
)

func RunServer() {
	//加载路由
	router := Routers()

	//http
	s := &http.Server{
		Addr:           ":8080",
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	s.ListenAndServe()
}

func Routers() *gin.Engine {
	Router := gin.Default()

	//处理跨域
	Router.Use(cors())
	var statikFS http.FileSystem
	var err error
	statikFS, err = fs.New()
	if err != nil {
		fmt.Println(err)
	}

	//注册前端页面 避免刷新丢失
	Router.GET("/", func(c *gin.Context) {
		file, _ := statikFS.Open("/index.html")
		io.Copy(c.Writer, file)
	})

	Router.Static("/static", "./web/static")

	ApiGroup := Router.Group("/api")
	router.InitRouter(ApiGroup) //注册用户相关接口路由

	//处理404
	Router.NoMethod(HandleNotFind)
	Router.NoRoute(HandleNotFind)

	return Router
}

// 处理404
func HandleNotFind(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{"message": "404 not found"})
}

// 跨域
func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Headers", "content-type, token, x-ca-stage")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type")
		c.Header("Access-Control-Allow-Credentials", "true")

		//过滤options请求
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		//处理请求
		c.Next()
	}
}
