// Package assets 以内嵌方式提供 Vue 前端产物(web 构建输出到 dist/)。
package assets

import (
	"embed"
	"io/fs"
	"net/http"
)

//go:embed all:dist
var dist embed.FS

// FileSystem 返回 dist 目录对应的 http.FileSystem,供静态资源与 SPA 回退使用。
func FileSystem() http.FileSystem {
	sub, err := fs.Sub(dist, "dist")
	if err != nil {
		panic(err)
	}
	return http.FS(sub)
}

// IndexHTML 返回入口页面内容(SPA fallback 与 "/" 路由使用)。
func IndexHTML() []byte {
	b, err := dist.ReadFile("dist/index.html")
	if err != nil {
		panic(err)
	}
	return b
}
