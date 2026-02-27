package jwt

import (
	"errors"
	"time"

	"ticket/config"

	"github.com/golang-jwt/jwt/v5"
)

// Claims 自定义 JWT 载荷
type Claims struct {
	UserId   int    `json:"userId"`
	Nickname string `json:"nickname"`
	Avatar   string `json:"avatar"`
	jwt.RegisteredClaims
}

// GenerateToken 生成 JWT
func GenerateToken(userId int, nickname, avatar string) (string, error) {
	cfg := config.GVA_CONFIG.Jwt
	expireSec := cfg.ExpireSec
	if expireSec <= 0 {
		expireSec = 86400
	}
	now := time.Now()
	claims := Claims{
		UserId:   userId,
		Nickname: nickname,
		Avatar:   avatar,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(expireSec) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.Secret))
}

// ParseToken 解析 JWT
func ParseToken(tokenString string) (*Claims, error) {
	cfg := config.GVA_CONFIG.Jwt
	secret := cfg.Secret
	if secret == "" {
		secret = "ticket"
	}
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(t *jwt.Token) (interface{}, error) {
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
