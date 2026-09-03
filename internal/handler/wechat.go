package handler

import (
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/gin-gonic/gin"

	"ticket/internal/config"
	"ticket/internal/pkg/response"
	"ticket/internal/pkg/wechat"
	"ticket/internal/service"
)

// newWechatClient 依据配置创建公众号客户端。
func newWechatClient() *wechat.Client {
	cfg := config.C.Wechat
	return wechat.NewClient(cfg.AppID, cfg.AppSecret, cfg.Token)
}

// WechatMessageHandler 公众号服务器验证(GET)与消息接收(POST)。
func WechatMessageHandler(c *gin.Context) {
	client := newWechatClient()

	if c.Request.Method == "GET" {
		if client.VerifySignature(c.Query("signature"), c.Query("timestamp"), c.Query("nonce")) {
			c.String(200, c.Query("echostr"))
		} else {
			c.String(403, "签名验证失败")
		}
		return
	}

	// POST 消息
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.String(500, "读取消息失败")
		return
	}
	message, err := client.ParseMessage(body)
	if err != nil {
		c.String(500, "解析消息失败")
		return
	}
	reply := handleWechatMessage(message, client)
	if reply == nil {
		c.String(200, "success")
		return
	}
	out, err := xml.Marshal(reply)
	if err != nil {
		c.String(500, "生成回复失败")
		return
	}
	c.Data(200, "application/xml; charset=utf-8", out)
}

func handleWechatMessage(message any, client *wechat.Client) any {
	switch msg := message.(type) {
	case *wechat.WXTextMessage:
		return handleWechatText(msg, client)
	case *wechat.WXImageMessage:
		return client.CreateTextReply(msg.FromUserName, msg.ToUserName, "收到您的图片,已记录。回复\"验证码\"可获取登录验证码。")
	case *wechat.WXVoiceMessage:
		return client.CreateTextReply(msg.FromUserName, msg.ToUserName, "收到您的语音消息。回复\"验证码\"可获取登录验证码。")
	case *wechat.WXVideoMessage:
		return client.CreateTextReply(msg.FromUserName, msg.ToUserName, "收到您的视频消息。")
	case *wechat.WXLocationMessage:
		return client.CreateTextReply(msg.FromUserName, msg.ToUserName, fmt.Sprintf("收到位置:%s", msg.Label))
	case *wechat.WXLinkMessage:
		return client.CreateTextReply(msg.FromUserName, msg.ToUserName, "收到您分享的链接:"+msg.Title)
	default:
		base := message.(*wechat.WXMessage)
		return client.CreateTextReply(base.FromUserName, base.ToUserName, "暂不支持该消息类型,回复\"帮助\"查看可用指令。")
	}
}

func handleWechatText(msg *wechat.WXTextMessage, client *wechat.Client) any {
	content := strings.TrimSpace(msg.Content)
	to, from := msg.FromUserName, msg.ToUserName
	switch {
	case strings.Contains(content, "验证码"):
		code := service.GenerateCode(from)
		fmt.Println("[wechat] code ->", from, code)
		reply := fmt.Sprintf("您的登录验证码是:%s\n有效期 10 分钟,请勿泄露。", code)
		return client.CreateTextReply(to, from, reply)
	case strings.Contains(content, "帮助"), strings.Contains(content, "help"):
		return client.CreateTextReply(to, from, "指令列表:\n1. 发送\"验证码\"获取公众号登录验证码\n2. 发送\"你好\"获取问候\n3. 发送\"关于\"了解我们")
	case strings.Contains(content, "你好"), strings.Contains(content, "hello"):
		return client.CreateTextReply(to, from, "您好!欢迎使用电影纪念票根 🎬")
	case strings.Contains(content, "关于"):
		return client.CreateTextReply(to, from, "电影纪念票根:在线生成电影纪念票根,支持模板/海报/自定义信息。")
	default:
		return client.CreateTextReply(to, from, "收到:"+content+"\n回复\"帮助\"查看可用指令。")
	}
}

// WechatLogin 用公众号 6 位验证码登录。
func WechatLogin(c *gin.Context) {
	var req struct {
		Code string `json:"code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数格式错误")
		return
	}
	if len(req.Code) != 6 {
		response.BizError(c, "请输入 6 位验证码")
		return
	}
	cfg := config.C.Wechat
	if cfg.AppID == "" || cfg.AppSecret == "" {
		response.BizError(c, "公众号尚未配置")
		return
	}
	client := wechat.NewClient(cfg.AppID, cfg.AppSecret, cfg.Token)
	u, token, isNew, err := service.WeChatLogin(config.DB, req.Code, client)
	if err != nil {
		replyServiceError(c, err)
		return
	}
	response.OK(c, gin.H{"token": token, "user": u.Profile(), "isNew": isNew})
}
