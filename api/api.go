package api

import (
	"encoding/xml"
	"fmt"
	"io"
	"math/rand"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"strconv"

	"ticket/config"
	"ticket/model"
	"ticket/package/jwt"
	"ticket/package/wechat"

	"ticket/types/response"

	"github.com/gin-gonic/gin"
)

// VerificationCodeInfo 验证码信息
type VerificationCodeInfo struct {
	OpenID     string
	Code       string
	CreateTime int64
	ExpireTime int64
	Used       bool
}

// VerificationCodeManager 验证码管理器（内存存储）
type VerificationCodeManager struct {
	codes sync.Map // map[string]*VerificationCodeInfo，key为验证码
}

var codeManager = &VerificationCodeManager{}

// init 初始化验证码管理器，启动过期清理协程
func init() {
	// 启动定时清理过期验证码的协程
	go codeManager.cleanExpiredCodes()
}

// GenerateCode 生成验证码并存储
func (m *VerificationCodeManager) GenerateCode(openID string) string {
	rand.Seed(time.Now().UnixNano())
	code := fmt.Sprintf("%06d", rand.Intn(1000000))

	now := time.Now().Unix()
	expireTime := now + 600 // 10分钟后过期

	codeInfo := &VerificationCodeInfo{
		OpenID:     openID,
		Code:       code,
		CreateTime: now,
		ExpireTime: expireTime,
		Used:       false,
	}

	m.codes.Store(code, codeInfo)
	return code
}

// VerifyCode 验证验证码
func (m *VerificationCodeManager) VerifyCode(code string) (string, error) {
	value, ok := m.codes.Load(code)
	if !ok {
		return "", fmt.Errorf("验证码不存在")
	}

	codeInfo := value.(*VerificationCodeInfo)

	// 检查是否已使用
	if codeInfo.Used {
		return "", fmt.Errorf("验证码已使用")
	}

	// 检查是否过期
	if time.Now().Unix() > codeInfo.ExpireTime {
		m.codes.Delete(code)
		return "", fmt.Errorf("验证码已过期")
	}

	// 标记为已使用
	codeInfo.Used = true
	m.codes.Store(code, codeInfo)

	return codeInfo.OpenID, nil
}

// cleanExpiredCodes 清理过期的验证码
func (m *VerificationCodeManager) cleanExpiredCodes() {
	ticker := time.NewTicker(5 * time.Minute) // 每5分钟清理一次
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now().Unix()
		m.codes.Range(func(key, value interface{}) bool {
			codeInfo := value.(*VerificationCodeInfo)
			if now > codeInfo.ExpireTime {
				m.codes.Delete(key)
			}
			return true
		})
	}
}

// SavePic 处理图片上传
func SavePic(c *gin.Context) {
	// 设置最大文件大小为10MB
	c.Request.ParseMultipartForm(10 << 20)

	// 获取上传的文件
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		response.Result(2, nil, "获取上传文件失败: "+err.Error(), c)
		return
	}
	defer file.Close()

	// 获取图片类型参数，默认为poster
	imageType := c.PostForm("type")
	if imageType == "" {
		imageType = "poster" // 默认为海报类型
	}

	// 验证文件类型
	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		response.Result(3, nil, "只支持图片文件上传", c)
		return
	}

	// 验证文件扩展名
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowedExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp"}
	isValidExt := false
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			isValidExt = true
			break
		}
	}
	if !isValidExt {
		response.Result(3, nil, "不支持的图片格式，请上传JPG、PNG、GIF、BMP或WebP格式的图片", c)
		return
	}

	// 异步上传图片
	go func() {
		dateFormat := time.Unix(time.Now().Unix(), 0).Format("20060102")
		filename := fmt.Sprintf("images/ticket/%s/%s/%d%s", imageType, dateFormat, time.Now().Unix(), ext)
		bufferBytes, _ := io.ReadAll(file)

		imageUrl, err := config.FileFactoryClient.PutFileWithBody(filename, bufferBytes)
		if err != nil {
			response.Result(500, nil, err.Error(), c)
			return
		}

		// 解析 header 中的 token，获取用户 id
		userId := 0
		if tokenStr := getTokenFromRequest(c); tokenStr != "" {
			if claims, err := jwt.ParseToken(tokenStr); err == nil {
				userId = claims.UserId
			}
		}

		image := model.Image{
			UserId:   userId,
			Type:     imageType,
			Filename: header.Filename,
			Url:      imageUrl,
			Ip:       c.ClientIP(),
		}
		image.Create(config.GVA_DB)
	}()
	// 返回成功响应
	response.Success("上传成功", c)
}

// GetUserProductList 获取登录用户成品列表
func GetUserProductList(c *gin.Context) {
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
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	num, _ := strconv.Atoi(c.DefaultQuery("num", "10"))
	if page < 1 {
		page = 1
	}
	if num < 1 || num > 100 {
		num = 10
	}
	list, err := model.GetImageListByUserAndType(config.GVA_DB, claims.UserId, "product", page, num)
	if err != nil {
		response.Result(500, nil, err.Error(), c)
		return
	}
	count, _ := model.GetImageCountByUserAndType(config.GVA_DB, claims.UserId, "product")
	response.Success(gin.H{
		"list":  list,
		"count": count,
	}, c)
}

func GetTemplate(c *gin.Context) {
	list, err := model.GetTemplateList(config.GVA_DB, 0, 1, 10)
	if err != nil {
		response.Result(500, nil, err.Error(), c)
		return
	}

	//获取模板总数
	count, _ := model.GetTemplateCount(config.GVA_DB, 0)
	response.Success(gin.H{
		"list":  list,
		"count": count,
	}, c)
}

func GetTypeSettingTemplate(c *gin.Context) {
	userId := 1
	list, err := model.GetTemplateList(config.GVA_DB, userId, 1, 10)
	if err != nil {
		response.Result(500, nil, err.Error(), c)
		return
	}

	//获取模板总数
	count, _ := model.GetTemplateCount(config.GVA_DB, userId)
	response.Success(gin.H{
		"list":  list,
		"count": count,
	}, c)
}

// WechatMessageHandler 微信公众号消息处理接口
func WechatMessageHandler(c *gin.Context) {
	// 创建微信客户端
	wechatClient := wechat.NewWechatClient(
		config.GVA_CONFIG.Wechat.AppID,
		config.GVA_CONFIG.Wechat.AppSecret,
		config.GVA_CONFIG.Wechat.Token,
	)

	// 处理GET请求 - 微信服务器验证
	if c.Request.Method == "GET" {
		signature := c.Query("signature")
		timestamp := c.Query("timestamp")
		nonce := c.Query("nonce")
		echostr := c.Query("echostr")

		// 验证签名
		if wechatClient.VerifySignature(signature, timestamp, nonce) {
			c.String(200, echostr)
		} else {
			c.String(403, "签名验证失败")
		}
		return
	}

	// 处理POST请求 - 接收消息
	if c.Request.Method == "POST" {
		// 读取请求体
		body, err := io.ReadAll(c.Request.Body)
		fmt.Println(string(body))
		if err != nil {
			c.String(500, "读取请求体失败")
			return
		}

		// 解析消息
		message, err := wechatClient.ParseMessage(body)
		if err != nil {
			c.String(500, "解析消息失败")
			return
		}

		// 处理不同类型的消息
		reply := handleWechatMessage(message, wechatClient)

		// 返回回复
		if reply != nil {
			replyXML, err := xml.Marshal(reply)
			if err != nil {
				c.String(500, "生成回复失败")
				return
			}
			c.Data(200, "application/xml; charset=utf-8", replyXML)
		} else {
			c.String(200, "success")
		}
	}
}

// handleWechatMessage 处理微信消息并生成回复
func handleWechatMessage(message interface{}, client *wechat.WechatClient) interface{} {
	switch msg := message.(type) {
	case *wechat.WXTextMessage:
		return handleTextMessage(msg, client)
	case *wechat.WXImageMessage:
		return handleImageMessage(msg, client)
	case *wechat.WXVoiceMessage:
		return handleVoiceMessage(msg, client)
	case *wechat.WXVideoMessage:
		return handleVideoMessage(msg, client)
	case *wechat.WXLocationMessage:
		return handleLocationMessage(msg, client)
	case *wechat.WXLinkMessage:
		return handleLinkMessage(msg, client)
	default:
		// 对于其他类型的消息，返回默认回复
		wxMsg := msg.(*wechat.WXMessage)
		return client.CreateTextReply(wxMsg.FromUserName, wxMsg.ToUserName, "收到您的消息，正在处理中...")
	}
}

// handleTextMessage 处理文本消息
func handleTextMessage(msg *wechat.WXTextMessage, client *wechat.WechatClient) interface{} {
	content := msg.Content

	// 根据消息内容进行不同的处理
	switch {
	case strings.Contains(content, "验证码"):
		// 生成6位数字验证码并存储到内存
		code := codeManager.GenerateCode(msg.FromUserName)
		fmt.Println("code:", code)

		replyText := fmt.Sprintf("您的验证码是：%s\n验证码有效期为10分钟，请妥善保管。", code)
		return client.CreateTextReply(msg.FromUserName, msg.ToUserName, replyText)

	case strings.Contains(content, "你好") || strings.Contains(content, "hello"):
		return client.CreateTextReply(msg.FromUserName, msg.ToUserName, "您好！欢迎使用我们的服务！")

	case strings.Contains(content, "帮助") || strings.Contains(content, "help"):
		helpText := "欢迎使用我们的服务！\n\n功能列表：\n1. 发送\"你好\" - 获取问候语\n2. 发送\"帮助\" - 获取帮助信息\n3. 发送\"关于\" - 了解我们\n4. 发送\"验证码\" - 获取验证码\n5. 发送图片 - 我们会回复您的图片\n\n如有其他问题，请联系客服。"
		return client.CreateTextReply(msg.FromUserName, msg.ToUserName, helpText)

	case strings.Contains(content, "关于"):
		return client.CreateTextReply(msg.FromUserName, msg.ToUserName, "我们是一个专业的服务平台，致力于为用户提供优质的服务体验。")

	default:
		// 默认回复
		return client.CreateTextReply(msg.FromUserName, msg.ToUserName, "收到您的消息："+content+"\n\n如需帮助，请发送\"帮助\"获取更多信息。")
	}
}

// handleImageMessage 处理图片消息
func handleImageMessage(msg *wechat.WXImageMessage, client *wechat.WechatClient) interface{} {
	// 可以在这里添加图片处理逻辑，比如保存图片、分析图片等
	return client.CreateTextReply(msg.FromUserName, msg.ToUserName, "收到您的图片！图片链接："+msg.PicURL)
}

// handleVoiceMessage 处理语音消息
func handleVoiceMessage(msg *wechat.WXVoiceMessage, client *wechat.WechatClient) interface{} {
	replyText := "收到您的语音消息！"
	if msg.Recognition != "" {
		replyText += "\n语音识别结果：" + msg.Recognition
	}
	return client.CreateTextReply(msg.FromUserName, msg.ToUserName, replyText)
}

// handleVideoMessage 处理视频消息
func handleVideoMessage(msg *wechat.WXVideoMessage, client *wechat.WechatClient) interface{} {
	return client.CreateTextReply(msg.FromUserName, msg.ToUserName, "收到您的视频消息！")
}

// handleLocationMessage 处理位置消息
func handleLocationMessage(msg *wechat.WXLocationMessage, client *wechat.WechatClient) interface{} {
	locationText := fmt.Sprintf("收到您的位置信息！\n位置：%s\n坐标：%.6f, %.6f",
		msg.Label, msg.LocationX, msg.LocationY)
	return client.CreateTextReply(msg.FromUserName, msg.ToUserName, locationText)
}

// handleLinkMessage 处理链接消息
func handleLinkMessage(msg *wechat.WXLinkMessage, client *wechat.WechatClient) interface{} {
	linkText := fmt.Sprintf("收到您分享的链接！\n标题：%s\n描述：%s\n链接：%s",
		msg.Title, msg.Description, msg.Url)
	return client.CreateTextReply(msg.FromUserName, msg.ToUserName, linkText)
}

// VerifyCodeAndGetUserInfo 验证验证码并获取用户信息
func VerifyCodeAndGetUserInfo(c *gin.Context) {
	// 获取验证码参数
	code := c.PostForm("code")
	if code == "" {
		code = c.Query("code")
	}

	if code == "" {
		response.ParamError("验证码不能为空", c)
		return
	}

	// 验证验证码格式（6位数字）
	if len(code) != 6 {
		response.Result(2, nil, "验证码格式错误，请输入6位数字", c)
		return
	}

	// 从内存验证验证码
	openID, err := codeManager.VerifyCode(code)
	if err != nil {
		response.Result(2, nil, "验证码无效或已过期: "+err.Error(), c)
		return
	}

	// 创建微信客户端
	wechatClient := wechat.NewWechatClient(
		config.GVA_CONFIG.Wechat.AppID,
		config.GVA_CONFIG.Wechat.AppSecret,
		config.GVA_CONFIG.Wechat.Token,
	)

	// 获取access_token
	accessToken, err := wechatClient.GetAccessToken()
	if err != nil {
		response.Result(500, nil, "获取微信access_token失败: "+err.Error(), c)
		return
	}

	fmt.Println("openID:", openID)
	// 根据openid获取用户信息
	userInfo, err := wechatClient.GetUserInfoByOpenID(accessToken, openID)
	if err != nil {
		response.Result(500, nil, "获取用户信息失败: "+err.Error(), c)
		return
	}

	// 检查用户是否关注公众号
	if userInfo.Subscribe == 0 {
		response.Result(2, nil, "用户未关注公众号，无法获取完整信息", c)
		return
	}

	// 返回用户信息
	response.Success(gin.H{
		"openid":     userInfo.OpenID,
		"nickname":   userInfo.Nickname,
		"sex":        userInfo.Sex,
		"province":   userInfo.Province,
		"city":       userInfo.City,
		"country":    userInfo.Country,
		"headimgurl": userInfo.HeadImgURL,
		"unionid":    userInfo.UnionID,
		"subscribe":  userInfo.Subscribe,
		"language":   userInfo.Language,
		"remark":     userInfo.Remark,
	}, c)
}
