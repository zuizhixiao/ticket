// Package router 路由装配与 HTTP 服务。
package router

import (
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"

	"ticket/internal/assets"
	"ticket/internal/config"
	"ticket/internal/handler"
	"ticket/internal/middleware"
	"ticket/internal/pkg/response"
)

// Run 启动 HTTP 服务。
func Run(cfg *config.Config) error {
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery(), middleware.CORS())
	_ = r.SetTrustedProxies(nil)

	registerAPI(r, cfg)
	registerStatic(r)
	port := 8080
	if cfg.Server.Port > 0 {
		port = cfg.Server.Port
	}
	return r.Run(":" + strconv.Itoa(port))
}

func registerAPI(r *gin.Engine, cfg *config.Config) {
	api := r.Group("/api")

	// 公开
	api.POST("/auth/captcha", handler.GetCaptcha)
	api.POST("/auth/register", handler.Register)
	api.POST("/auth/login", handler.Login)
	api.POST("/auth/reset-password", handler.ResetPassword)
	api.GET("/templates", handler.PublicTemplates)
	api.Any("/wechat/message", handler.WechatMessageHandler)
	api.POST("/wechat/login", handler.WechatLogin)

	// 游客可上传海报/成品(可选鉴权,带 token 时记录归属)
	api.POST("/uploads", middleware.JWTAuthOptional(cfg.Jwt.Secret), handler.Upload)

	// 登录用户
	user := api.Group("", middleware.JWTAuth(cfg.Jwt.Secret))
	{
		user.GET("/auth/me", handler.Me)
		user.PUT("/auth/profile", handler.UpdateProfile)
		user.GET("/user/products", handler.ListProducts)
		user.DELETE("/user/products/:id", handler.DeleteProduct)
	}

	// 管理员
	admin := api.Group("", middleware.JWTAuth(cfg.Jwt.Secret), middleware.AdminOnly())
	{
		admin.GET("/admin/templates", handler.AdminTemplates)
		admin.POST("/admin/templates", handler.AdminCreateTemplate)
		admin.PUT("/admin/templates/sort", handler.AdminTemplatesSort)
		admin.PUT("/admin/templates/:id", handler.AdminUpdateTemplate)
		admin.DELETE("/admin/templates/:id", handler.AdminDeleteTemplate)
		admin.GET("/admin/images", handler.AdminImages)
		admin.DELETE("/admin/images/:id", handler.AdminDeleteImage)
		admin.GET("/admin/users", handler.AdminUsers)
		admin.PUT("/admin/users/:id/status", handler.AdminUserStatus)
		admin.PUT("/admin/users/:id/password", handler.AdminUserResetPassword)
	}
}

// registerStatic 内嵌前端:首页 /assets 静态 + SPA 回退。
func registerStatic(r *gin.Engine) {
	index := assets.IndexHTML()
	assetFS := assets.FileSystem()

	r.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", index)
	})

	// Vite 产物输出到 dist/assets/,URL /assets/* ↔ dist/assets/*
	r.GET("/assets/*filepath", func(c *gin.Context) {
		name := "/assets" + c.Param("filepath")
		f, err := assetFS.Open(name)
		if err != nil {
			c.Status(http.StatusNotFound)
			return
		}
		defer f.Close()
		stat, err := f.Stat()
		if err != nil {
			c.Status(http.StatusInternalServerError)
			return
		}
		http.ServeContent(c.Writer, c.Request, filepath.Base(name), stat.ModTime(), f)
	})

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/api/") {
			response.NotFound(c, "接口不存在")
			return
		}
		c.Data(http.StatusOK, "text/html; charset=utf-8", index)
	})
}
