// Package captcha 图形验证码(4 位数字,内存存储)。
package captcha

import (
	"github.com/mojocn/base64Captcha"
)

var store = base64Captcha.DefaultMemStore

// Generate 生成验证码,返回 captchaId 与 base64 PNG 图片。
func Generate() (id, b64 string, err error) {
	driver := base64Captcha.NewDriverDigit(80, 240, 4, 0.7, 80)
	cp := base64Captcha.NewCaptcha(driver, store)
	id, b64, _, err = cp.Generate()
	return
}

// Verify 校验(一次性,失败即销毁)。
func Verify(id, code string) bool {
	return store.Verify(id, code, true)
}
