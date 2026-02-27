package wechat

import (
	"crypto/sha1"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// WechatClient 微信客户端
type WechatClient struct {
	AppID     string
	AppSecret string
	Token     string // 微信公众号Token
}

// NewWechatClient 创建微信客户端
func NewWechatClient(appID, appSecret, token string) *WechatClient {
	return &WechatClient{
		AppID:     appID,
		AppSecret: appSecret,
		Token:     token,
	}
}

// Code2SessionResponse 微信登录响应
type Code2SessionResponse struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// UserInfoResponse 微信用户信息响应
type UserInfoResponse struct {
	OpenID     string `json:"openid"`
	Nickname   string `json:"nickname"`
	Sex        int    `json:"sex"`
	Province   string `json:"province"`
	City       string `json:"city"`
	Country    string `json:"country"`
	HeadImgURL string `json:"headimgurl"`
	UnionID    string `json:"unionid"`
}

// AccessTokenResponse 获取access_token响应
type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

// UserInfoByOpenIDResponse 根据openid获取用户信息响应（公众号）
type UserInfoByOpenIDResponse struct {
	Subscribe     int    `json:"subscribe"`      // 用户是否订阅该公众号标识，值为0时，代表此用户没有关注该公众号
	OpenID        string `json:"openid"`         // 用户的标识
	Nickname      string `json:"nickname"`       // 用户的昵称
	Sex           int    `json:"sex"`            // 用户的性别，值为1时是男性，值为2时是女性，值为0时是未知
	Language      string `json:"language"`       // 用户的语言
	City          string `json:"city"`           // 用户所在城市
	Province      string `json:"province"`       // 用户所在省份
	Country       string `json:"country"`        // 用户所在国家
	HeadImgURL    string `json:"headimgurl"`     // 用户头像
	SubscribeTime int64  `json:"subscribe_time"` // 用户关注时间
	UnionID       string `json:"unionid"`        // 只有在用户将公众号绑定到微信开放平台账号后，才会出现该字段
	Remark        string `json:"remark"`         // 公众号运营者对粉丝的备注
	GroupID       int    `json:"groupid"`        // 用户所在的分组ID
	TagIDList     []int  `json:"tagid_list"`     // 用户被打上的标签ID列表
	ErrCode       int    `json:"errcode"`
	ErrMsg        string `json:"errmsg"`
}

// 微信公众号消息相关结构体

// WXMessage 微信消息基础结构
type WXMessage struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:"ToUserName"`
	FromUserName string   `xml:"FromUserName"`
	CreateTime   int64    `xml:"CreateTime"`
	MsgType      string   `xml:"MsgType"`
	MsgId        int64    `xml:"MsgId,omitempty"`
}

// WXTextMessage 文本消息
type WXTextMessage struct {
	WXMessage
	Content string `xml:"Content"`
}

// WXImageMessage 图片消息
type WXImageMessage struct {
	WXMessage
	PicURL  string `xml:"PicUrl"`
	MediaId string `xml:"MediaId"`
}

// WXVoiceMessage 语音消息
type WXVoiceMessage struct {
	WXMessage
	MediaId     string `xml:"MediaId"`
	Format      string `xml:"Format"`
	Recognition string `xml:"Recognition,omitempty"`
}

// WXVideoMessage 视频消息
type WXVideoMessage struct {
	WXMessage
	MediaId      string `xml:"MediaId"`
	ThumbMediaId string `xml:"ThumbMediaId"`
}

// WXLocationMessage 地理位置消息
type WXLocationMessage struct {
	WXMessage
	LocationX float64 `xml:"Location_X"`
	LocationY float64 `xml:"Location_Y"`
	Scale     int     `xml:"Scale"`
	Label     string  `xml:"Label"`
}

// WXLinkMessage 链接消息
type WXLinkMessage struct {
	WXMessage
	Title       string `xml:"Title"`
	Description string `xml:"Description"`
	Url         string `xml:"Url"`
}

// WXReplyMessage 回复消息基础结构
type WXReplyMessage struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:"ToUserName"`
	FromUserName string   `xml:"FromUserName"`
	CreateTime   int64    `xml:"CreateTime"`
	MsgType      string   `xml:"MsgType"`
}

// WXTextReply 文本回复
type WXTextReply struct {
	WXReplyMessage
	Content string `xml:"Content"`
}

// WXImageReply 图片回复
type WXImageReply struct {
	WXReplyMessage
	Image struct {
		MediaId string `xml:"MediaId"`
	} `xml:"Image"`
}

// WXVoiceReply 语音回复
type WXVoiceReply struct {
	WXReplyMessage
	Voice struct {
		MediaId string `xml:"MediaId"`
	} `xml:"Voice"`
}

// WXVideoReply 视频回复
type WXVideoReply struct {
	WXReplyMessage
	Video struct {
		MediaId     string `xml:"MediaId"`
		Title       string `xml:"Title,omitempty"`
		Description string `xml:"Description,omitempty"`
	} `xml:"Video"`
}

// WXNewsReply 图文回复
type WXNewsReply struct {
	WXReplyMessage
	ArticleCount int `xml:"ArticleCount"`
	Articles     struct {
		Item []WXNewsItem `xml:"item"`
	} `xml:"Articles"`
}

// WXNewsItem 图文消息项
type WXNewsItem struct {
	Title       string `xml:"Title"`
	Description string `xml:"Description"`
	PicURL      string `xml:"PicUrl"`
	URL         string `xml:"Url"`
}

// Code2Session 通过code获取用户openid和session_key
func (wc *WechatClient) Code2Session(code string) (*Code2SessionResponse, error) {
	apiURL := "https://api.weixin.qq.com/sns/jscode2session"

	params := url.Values{}
	params.Set("appid", wc.AppID)
	params.Set("secret", wc.AppSecret)
	params.Set("js_code", code)
	params.Set("grant_type", "authorization_code")

	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	resp, err := http.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("请求微信API失败: %v", err)
	}
	defer resp.Body.Close()

	var result Code2SessionResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析微信API响应失败: %v", err)
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("微信API错误: %d - %s", result.ErrCode, result.ErrMsg)
	}

	return &result, nil
}

// GetAccessToken 获取公众号access_token
func (wc *WechatClient) GetAccessToken() (string, error) {
	apiURL := "https://api.weixin.qq.com/cgi-bin/token"

	params := url.Values{}
	params.Set("grant_type", "client_credential")
	params.Set("appid", wc.AppID)
	params.Set("secret", wc.AppSecret)

	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	resp, err := http.Get(fullURL)
	if err != nil {
		return "", fmt.Errorf("请求微信access_token失败: %v", err)
	}
	defer resp.Body.Close()
	fmt.Println(fullURL)
	var result AccessTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("解析微信access_token响应失败: %v", err)
	}

	if result.ErrCode != 0 {
		return "", fmt.Errorf("微信API错误: %d - %s", result.ErrCode, result.ErrMsg)
	}

	return result.AccessToken, nil
}

// GetUserInfo 获取用户信息（需要用户授权）
func (wc *WechatClient) GetUserInfo(accessToken, openid string) (*UserInfoResponse, error) {
	apiURL := "https://api.weixin.qq.com/sns/userinfo"

	params := url.Values{}
	params.Set("access_token", accessToken)
	params.Set("openid", openid)
	params.Set("lang", "zh_CN")

	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	resp, err := http.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("请求微信用户信息失败: %v", err)
	}
	defer resp.Body.Close()

	var result UserInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析微信用户信息失败: %v", err)
	}

	return &result, nil
}

// GetUserInfoByOpenID 根据openid获取用户信息（需要用户关注公众号）
func (wc *WechatClient) GetUserInfoByOpenID(accessToken, openid string) (*UserInfoByOpenIDResponse, error) {
	apiURL := "https://api.weixin.qq.com/cgi-bin/user/info"

	params := url.Values{}
	params.Set("access_token", accessToken)
	params.Set("openid", openid)
	params.Set("lang", "zh_CN")

	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	resp, err := http.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("请求微信用户信息失败: %v", err)
	}
	defer resp.Body.Close()

	var result UserInfoByOpenIDResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析微信用户信息失败: %v", err)
	}

	if result.ErrCode != 0 {
		return nil, fmt.Errorf("微信API错误: %d - %s", result.ErrCode, result.ErrMsg)
	}

	return &result, nil
}

// VerifySignature 验证微信消息签名
func (wc *WechatClient) VerifySignature(signature, timestamp, nonce string) bool {
	// 将token、timestamp、nonce三个参数进行字典序排序
	params := []string{wc.Token, timestamp, nonce}
	sort.Strings(params)

	// 将三个参数字符串拼接成一个字符串进行sha1加密
	str := strings.Join(params, "")
	h := sha1.New()
	h.Write([]byte(str))
	hash := fmt.Sprintf("%x", h.Sum(nil))

	// 开发者获得加密后的字符串可与signature对比，标识该请求来源于微信
	return hash == signature
}

// ParseMessage 解析微信消息
func (wc *WechatClient) ParseMessage(data []byte) (interface{}, error) {
	// 先解析基础消息结构
	var baseMsg WXMessage
	if err := xml.Unmarshal(data, &baseMsg); err != nil {
		return nil, fmt.Errorf("解析消息失败: %v", err)
	}

	// 根据消息类型解析具体内容
	switch baseMsg.MsgType {
	case "text":
		var msg WXTextMessage
		if err := xml.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("解析文本消息失败: %v", err)
		}
		return &msg, nil
	case "image":
		var msg WXImageMessage
		if err := xml.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("解析图片消息失败: %v", err)
		}
		return &msg, nil
	case "voice":
		var msg WXVoiceMessage
		if err := xml.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("解析语音消息失败: %v", err)
		}
		return &msg, nil
	case "video":
		var msg WXVideoMessage
		if err := xml.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("解析视频消息失败: %v", err)
		}
		return &msg, nil
	case "location":
		var msg WXLocationMessage
		if err := xml.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("解析位置消息失败: %v", err)
		}
		return &msg, nil
	case "link":
		var msg WXLinkMessage
		if err := xml.Unmarshal(data, &msg); err != nil {
			return nil, fmt.Errorf("解析链接消息失败: %v", err)
		}
		return &msg, nil
	default:
		return &baseMsg, nil
	}
}

// CreateTextReply 创建文本回复
func (wc *WechatClient) CreateTextReply(toUser, fromUser, content string) *WXTextReply {
	return &WXTextReply{
		WXReplyMessage: WXReplyMessage{
			ToUserName:   toUser,
			FromUserName: fromUser,
			CreateTime:   time.Now().Unix(),
			MsgType:      "text",
		},
		Content: content,
	}
}

// CreateImageReply 创建图片回复
func (wc *WechatClient) CreateImageReply(toUser, fromUser, mediaId string) *WXImageReply {
	reply := &WXImageReply{
		WXReplyMessage: WXReplyMessage{
			ToUserName:   toUser,
			FromUserName: fromUser,
			CreateTime:   time.Now().Unix(),
			MsgType:      "image",
		},
	}
	reply.Image.MediaId = mediaId
	return reply
}

// CreateNewsReply 创建图文回复
func (wc *WechatClient) CreateNewsReply(toUser, fromUser string, articles []WXNewsItem) *WXNewsReply {
	reply := &WXNewsReply{
		WXReplyMessage: WXReplyMessage{
			ToUserName:   toUser,
			FromUserName: fromUser,
			CreateTime:   time.Now().Unix(),
			MsgType:      "news",
		},
		ArticleCount: len(articles),
	}
	reply.Articles.Item = articles
	return reply
}
