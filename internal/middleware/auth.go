// Package middleware Gin 中间件:JWT 鉴权、管理员校验、CORS。
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"ticket/internal/model"
	"ticket/internal/pkg/jwt"
	"ticket/internal/pkg/response"
)

const ctxClaimsKey = "ticket.claims"

// JWT 解析后的载荷。
type JWT struct{ *jwt.Claims }

// JWTAuth 校验 Authorization: Bearer <token>(兼容 token 请求头)。
func JWTAuth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			response.Unauthorized(c, "请先登录")
			return
		}
		claims, err := jwt.Parse(secret, tokenStr)
		if err != nil {
			response.Unauthorized(c, "登录已过期,请重新登录")
			return
		}
		c.Set(ctxClaimsKey, &JWT{Claims: claims})
		c.Next()
	}
}

// JWTAuthOptional 可选鉴权:无 token 或无效 token 时放行为游客
// (用于游客也可上传海报/成品;无效 token 仍返回 401 以便前端刷新登录态)。
func JWTAuthOptional(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractToken(c)
		if tokenStr == "" {
			c.Next()
			return
		}
		claims, err := jwt.Parse(secret, tokenStr)
		if err != nil {
			response.Unauthorized(c, "登录已过期,请重新登录")
			return
		}
		c.Set(ctxClaimsKey, &JWT{Claims: claims})
		c.Next()
	}
}

// AdminOnly 需与 JWTAuth 连用,校验管理员角色。
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, ok := c.Get(ctxClaimsKey)
		if !ok || claims.(*JWT).Role != model.RoleAdmin {
			response.Forbidden(c, "需要管理员权限")
			return
		}
		c.Next()
	}
}

// ClaimsOf 从上下文取当前用户载荷。
func ClaimsOf(c *gin.Context) (*JWT, bool) {
	v, ok := c.Get(ctxClaimsKey)
	if !ok {
		return nil, false
	}
	return v.(*JWT), true
}

func extractToken(c *gin.Context) string {
	if s := c.GetHeader("Authorization"); s != "" {
		if strings.HasPrefix(s, "Bearer ") {
			return strings.TrimSpace(strings.TrimPrefix(s, "Bearer "))
		}
	}
	return strings.TrimSpace(c.GetHeader("token"))
}
