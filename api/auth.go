package api

import (
	"strings"
	"time"

	"ticket/config"
	"ticket/model"
	"ticket/package/jwt"
	"ticket/types/response"

	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"
	"gorm.io/gorm"
)

var captchaStore = base64Captcha.DefaultMemStore

// GetCaptcha 获取图形验证码
func GetCaptcha(c *gin.Context) {
	driver := base64Captcha.NewDriverDigit(80, 240, 4, 0.7, 80)
	captcha := base64Captcha.NewCaptcha(driver, captchaStore)
	id, b64s, _, err := captcha.Generate()
	if err != nil {
		response.Result(500, nil, "生成验证码失败", c)
		return
	}
	response.Success(gin.H{
		"captchaId":  id,
		"captchaImg": b64s,
	}, c)
}

// verifyCaptcha 验证图形验证码
func verifyCaptcha(captchaId, code string) bool {
	return captchaStore.Verify(captchaId, code, true)
}

// LoginRequest 登录请求
type LoginRequest struct {
	Nickname string `json:"nickname" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 登录接口
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError("参数错误: "+err.Error(), c)
		return
	}
	user, err := model.GetUserByNickname(config.GVA_DB, req.Nickname)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Result(2, nil, "昵称或密码错误", c)
			return
		}
		response.Result(500, nil, err.Error(), c)
		return
	}
	if user.Status == 0 {
		response.Result(2, nil, "账号已被禁用", c)
		return
	}
	if !user.CheckPassword(req.Password) {
		response.Result(2, nil, "昵称或密码错误", c)
		return
	}
	// 更新最后登录时间
	now := time.Now().Unix()
	config.GVA_DB.Model(user).Update("lastLoginTime", now)
	// 生成 JWT
	token, err := jwt.GenerateToken(user.Id, user.Nickname, user.Avatar)
	if err != nil {
		response.Result(500, nil, "生成凭证失败", c)
		return
	}
	response.Success(gin.H{
		"token":    token,
		"id":       user.Id,
		"nickname": user.Nickname,
		"avatar":   user.Avatar,
	}, c)
}

// getTokenFromRequest 从请求中获取 token（Header Authorization: Bearer、Header token、Query token）
func getTokenFromRequest(c *gin.Context) string {
	if s := c.GetHeader("Authorization"); s != "" {
		if strings.HasPrefix(s, "Bearer ") {
			return strings.TrimPrefix(s, "Bearer ")
		}
	}
	if s := c.GetHeader("token"); s != "" {
		return s
	}
	return c.Query("token")
}

// GetUserInfo 获取当前登录用户信息
func GetUserInfo(c *gin.Context) {
	tokenStr := getTokenFromRequest(c)
	if tokenStr == "" {
		response.Result(401, nil, "请先登录", c)
		return
	}
	claims, err := jwt.ParseToken(tokenStr)
	if err != nil {
		response.Result(401, nil, "登录已过期，请重新登录", c)
		return
	}
	// 从数据库获取最新用户信息
	user, err := model.GetUserById(config.GVA_DB, claims.UserId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Result(401, nil, "用户不存在", c)
			return
		}
		response.Result(500, nil, err.Error(), c)
		return
	}
	if user.Status == 0 {
		response.Result(2, nil, "账号已被禁用", c)
		return
	}
	response.Success(gin.H{
		"id":       user.Id,
		"nickname": user.Nickname,
		"avatar":   user.Avatar,
	}, c)
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Nickname  string `json:"nickname" binding:"required"`
	CaptchaId string `json:"captchaId" binding:"required"`
	Code      string `json:"code" binding:"required"`
	Password  string `json:"password" binding:"required"`
}

// Register 注册接口
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError("参数错误: "+err.Error(), c)
		return
	}
	if !verifyCaptcha(req.CaptchaId, req.Code) {
		response.Result(2, nil, "图形验证码错误", c)
		return
	}
	// 检查昵称是否已注册
	_, err := model.GetUserByNickname(config.GVA_DB, req.Nickname)
	if err == nil {
		response.Result(2, nil, "该昵称已注册", c)
		return
	}
	if err != gorm.ErrRecordNotFound {
		response.Result(500, nil, err.Error(), c)
		return
	}
	now := time.Now().Unix()
	user := &model.User{
		Nickname:   req.Nickname,
		Status:     1,
		CreateTime: now,
		UpdateTime: &now,
	}
	if err := user.SetPassword(req.Password); err != nil {
		response.Result(500, nil, "密码加密失败", c)
		return
	}
	if err := user.Create(config.GVA_DB); err != nil {
		response.Result(500, nil, err.Error(), c)
		return
	}
	response.Success(gin.H{
		"id": user.Id,
	}, c)
}

// ResetPasswordRequest 重置密码请求
type ResetPasswordRequest struct {
	Nickname  string `json:"nickname" binding:"required"`
	CaptchaId string `json:"captchaId" binding:"required"`
	Code      string `json:"code" binding:"required"`
	Password  string `json:"password" binding:"required"`
}

// ResetPassword 重置密码接口
func ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError("参数错误: "+err.Error(), c)
		return
	}
	if !verifyCaptcha(req.CaptchaId, req.Code) {
		response.Result(2, nil, "图形验证码错误", c)
		return
	}
	user, err := model.GetUserByNickname(config.GVA_DB, req.Nickname)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Result(2, nil, "该昵称未注册", c)
			return
		}
		response.Result(500, nil, err.Error(), c)
		return
	}
	if err := user.SetPassword(req.Password); err != nil {
		response.Result(500, nil, "密码加密失败", c)
		return
	}
	now := time.Now().Unix()
	user.UpdateTime = &now
	if err := user.Update(config.GVA_DB, "password", "updateTime"); err != nil {
		response.Result(500, nil, err.Error(), c)
		return
	}
	response.Success("密码重置成功", c)
}
