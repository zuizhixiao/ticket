// Package wechat 微信公众号客户端:签名校验、消息收发、access_token、用户信息。
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

// Client 微信公众号客户端。
type Client struct {
	AppID     string
	AppSecret string
	Token     string
}

func NewClient(appID, appSecret, token string) *Client {
	return &Client{AppID: appID, AppSecret: appSecret, Token: token}
}

// ---------- 消息结构 ----------

type WXMessage struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:"ToUserName"`
	FromUserName string   `xml:"FromUserName"`
	CreateTime   int64    `xml:"CreateTime"`
	MsgType      string   `xml:"MsgType"`
	MsgId        int64    `xml:"MsgId,omitempty"`
}

type WXTextMessage struct {
	WXMessage
	Content string `xml:"Content"`
}

type WXImageMessage struct {
	WXMessage
	PicURL  string `xml:"PicUrl"`
	MediaId string `xml:"MediaId"`
}

type WXVoiceMessage struct {
	WXMessage
	MediaId     string `xml:"MediaId"`
	Format      string `xml:"Format"`
	Recognition string `xml:"Recognition,omitempty"`
}

type WXVideoMessage struct {
	WXMessage
	MediaId      string `xml:"MediaId"`
	ThumbMediaId string `xml:"ThumbMediaId"`
}

type WXLocationMessage struct {
	WXMessage
	LocationX float64 `xml:"Location_X"`
	LocationY float64 `xml:"Location_Y"`
	Scale     int     `xml:"Scale"`
	Label     string  `xml:"Label"`
}

type WXLinkMessage struct {
	WXMessage
	Title       string `xml:"Title"`
	Description string `xml:"Description"`
	Url         string `xml:"Url"`
}

type WXReplyMessage struct {
	XMLName      xml.Name `xml:"xml"`
	ToUserName   string   `xml:"ToUserName"`
	FromUserName string   `xml:"FromUserName"`
	CreateTime   int64    `xml:"CreateTime"`
	MsgType      string   `xml:"MsgType"`
}

type WXTextReply struct {
	WXReplyMessage
	Content string `xml:"Content"`
}

type WXNewsReply struct {
	WXReplyMessage
	ArticleCount int `xml:"ArticleCount"`
	Articles     struct {
		Item []WXNewsItem `xml:"item"`
	} `xml:"Articles"`
}

type WXNewsItem struct {
	Title       string `xml:"Title"`
	Description string `xml:"Description"`
	PicURL      string `xml:"PicUrl"`
	URL         string `xml:"Url"`
}

// ---------- 用户 / token ----------

type AccessTokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	ErrCode     int    `json:"errcode"`
	ErrMsg      string `json:"errmsg"`
}

type UserInfoResponse struct {
	Subscribe     int    `json:"subscribe"`
	OpenID        string `json:"openid"`
	Nickname      string `json:"nickname"`
	Sex           int    `json:"sex"`
	Language      string `json:"language"`
	City          string `json:"city"`
	Province      string `json:"province"`
	Country       string `json:"country"`
	HeadImgURL    string `json:"headimgurl"`
	SubscribeTime int64  `json:"subscribe_time"`
	UnionID       string `json:"unionid"`
	Remark        string `json:"remark"`
	ErrCode       int    `json:"errcode"`
	ErrMsg        string `json:"errmsg"`
}

// ---------- 方法 ----------

// VerifySignature 校验微信服务器签名。
func (c *Client) VerifySignature(signature, timestamp, nonce string) bool {
	params := []string{c.Token, timestamp, nonce}
	sort.Strings(params)
	h := sha1.Sum([]byte(strings.Join(params, "")))
	return fmt.Sprintf("%x", h) == signature
}

// ParseMessage 解析消息 XML,返回对应消息结构体指针。
func (c *Client) ParseMessage(data []byte) (any, error) {
	var base WXMessage
	if err := xml.Unmarshal(data, &base); err != nil {
		return nil, err
	}
	switch base.MsgType {
	case "text":
		var m WXTextMessage
		return &m, xml.Unmarshal(data, &m)
	case "image":
		var m WXImageMessage
		return &m, xml.Unmarshal(data, &m)
	case "voice":
		var m WXVoiceMessage
		return &m, xml.Unmarshal(data, &m)
	case "video":
		var m WXVideoMessage
		return &m, xml.Unmarshal(data, &m)
	case "location":
		var m WXLocationMessage
		return &m, xml.Unmarshal(data, &m)
	case "link":
		var m WXLinkMessage
		return &m, xml.Unmarshal(data, &m)
	default:
		return &base, nil
	}
}

func (c *Client) CreateTextReply(toUser, fromUser, content string) *WXTextReply {
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

func (c *Client) CreateNewsReply(toUser, fromUser string, items []WXNewsItem) *WXNewsReply {
	r := &WXNewsReply{
		WXReplyMessage: WXReplyMessage{
			ToUserName:   toUser,
			FromUserName: fromUser,
			CreateTime:   time.Now().Unix(),
			MsgType:      "news",
		},
		ArticleCount: len(items),
	}
	r.Articles.Item = items
	return r
}

// GetAccessToken 获取公众号全局 access_token。
func (c *Client) GetAccessToken() (string, error) {
	u := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s",
		url.QueryEscape(c.AppID), url.QueryEscape(c.AppSecret))
	resp, err := http.Get(u)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var r AccessTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", err
	}
	if r.ErrCode != 0 {
		return "", fmt.Errorf("wechat errcode=%d %s", r.ErrCode, r.ErrMsg)
	}
	return r.AccessToken, nil
}

// GetUserInfoByOpenID 根据 openid 获取公众号用户信息(需已关注)。
func (c *Client) GetUserInfoByOpenID(accessToken, openid string) (*UserInfoResponse, error) {
	u := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/user/info?access_token=%s&openid=%s&lang=zh_CN",
		url.QueryEscape(accessToken), url.QueryEscape(openid))
	resp, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var r UserInfoResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	if r.ErrCode != 0 {
		return nil, fmt.Errorf("wechat errcode=%d %s", r.ErrCode, r.ErrMsg)
	}
	return &r, nil
}
