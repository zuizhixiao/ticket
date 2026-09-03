// Package jwt 签发/校验 HS256 Token。
package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims 自定义载荷。
type Claims struct {
	UserId   int    `json:"userId"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	Role     int    `json:"role"`
	Openid   string `json:"openid"`
	jwt.RegisteredClaims
}

// Config 签发参数。
type Config struct {
	Secret        string
	ExpireSeconds int
}

// Generate 签发 token。
func Generate(cfg Config, userId int, nickname, avatar, openid string, role int) (string, error) {
	expire := cfg.ExpireSeconds
	if expire <= 0 {
		expire = 86400
	}
	now := time.Now()
	claims := Claims{
		UserId:   userId,
		Nickname: nickname,
		Avatar:   avatar,
		Role:     role,
		Openid:   openid,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expire) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(cfg.Secret))
}

// Parse 校验并解析 token。
func Parse(secret, tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
