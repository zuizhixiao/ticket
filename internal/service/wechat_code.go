package service

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// codeInfo 验证码信息
type codeInfo struct {
	OpenID     string
	Code       string
	ExpireTime int64
	Used       bool
}

// verifyCodeManager 公众号 6 位验证码(内存,10 分钟过期,5 分钟清扫)。
type verifyCodeManager struct {
	mu    sync.Mutex
	codes map[string]*codeInfo // key: code
}

var codeMgr = &verifyCodeManager{codes: make(map[string]*codeInfo)}

func init() {
	go codeMgr.cleanup()
}

func (m *verifyCodeManager) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		now := time.Now().Unix()
		m.mu.Lock()
		for k, v := range m.codes {
			if now > v.ExpireTime {
				delete(m.codes, k)
			}
		}
		m.mu.Unlock()
	}
}

// GenerateCode 为 openid 生成验证码并存储。
func GenerateCode(openID string) string {
	code := fmt.Sprintf("%06d", rand.Intn(1000000))
	expire := time.Now().Add(10 * time.Minute).Unix()
	m := codeMgr
	m.mu.Lock()
	defer m.mu.Unlock()
	m.codes[code] = &codeInfo{OpenID: openID, Code: code, ExpireTime: expire}
	return code
}

// VerifyCode 校验并消费验证码,返回 openid。
func VerifyCode(code string) (string, error) {
	m := codeMgr
	m.mu.Lock()
	defer m.mu.Unlock()
	info, ok := m.codes[code]
	if !ok {
		return "", errors.New("验证码不存在")
	}
	if info.Used {
		return "", errors.New("验证码已使用")
	}
	if time.Now().Unix() > info.ExpireTime {
		delete(m.codes, code)
		return "", errors.New("验证码已过期")
	}
	info.Used = true
	return info.OpenID, nil
}
