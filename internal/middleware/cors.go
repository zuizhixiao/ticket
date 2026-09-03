package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS 允许跨域(dev 默认全放行;可用 CORS_ORIGINS 限定)。
func CORS() gin.HandlerFunc {
	allowed := "*"
	if v := os.Getenv("CORS_ORIGINS"); v != "" {
		allowed = v
	}
	origins := map[string]bool{}
	for _, o := range strings.Split(allowed, ",") {
		origins[strings.TrimSpace(o)] = true
	}
	allowAll := origins["*"]

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		if allowAll || origin != "" && (origins[origin] || origins["*"]) {
			if allowAll {
				c.Header("Access-Control-Allow-Origin", "*")
			} else {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Vary", "Origin")
			}
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization, token")
			c.Header("Access-Control-Max-Age", "86400")
		}
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
