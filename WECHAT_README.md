# 微信公众号消息接收接口

## 功能概述

本项目已成功集成了微信公众号消息接收接口，支持接收和处理用户发送的各种类型的消息，并能够自动回复。

## 主要功能

### 1. 消息类型支持
- **文本消息**: 支持关键词回复和智能对话
- **图片消息**: 接收用户发送的图片
- **语音消息**: 支持语音识别（如果微信提供）
- **视频消息**: 接收用户发送的视频
- **位置消息**: 接收用户分享的位置信息
- **链接消息**: 接收用户分享的链接

### 2. 智能回复功能
- **关键词回复**: 支持"你好"、"帮助"、"关于"等关键词
- **默认回复**: 对于未识别的消息提供友好的默认回复
- **帮助系统**: 提供详细的功能说明和使用指南

## 配置说明

### 1. 配置文件设置
在 `config.yaml` 文件中配置微信公众号信息：

```yaml
# 微信公众号配置
wechat:
  app_id: "your_app_id_here"
  app_secret: "your_app_secret_here"
  token: "your_token_here"  # 请替换为您的微信公众号Token
```

### 2. 微信公众号后台配置
1. 登录微信公众平台
2. 进入"设置与开发" -> "基本配置"
3. 设置服务器配置：
   - URL: `http://your-domain.com/api/wechat/message`
   - Token: 与配置文件中的token保持一致
   - EncodingAESKey: 可选，用于消息加密

## API接口

### 接口地址
```
POST/GET /api/wechat/message
```

### 请求方式
- **GET**: 用于微信服务器验证
- **POST**: 用于接收用户消息

### 参数说明

#### GET请求参数（验证用）
- `signature`: 微信加密签名
- `timestamp`: 时间戳
- `nonce`: 随机数
- `echostr`: 随机字符串

#### POST请求
- 请求体为XML格式的微信消息

## 使用示例

### 1. 启动服务器
```bash
go run main.go
```

### 2. 测试接口
可以使用以下命令测试接口是否正常工作：

```bash
# 测试GET请求（验证）
curl "http://localhost:8080/api/wechat/message?signature=xxx&timestamp=xxx&nonce=xxx&echostr=xxx"

# 测试POST请求（消息接收）
curl -X POST "http://localhost:8080/api/wechat/message" \
  -H "Content-Type: application/xml" \
  -d '<xml><ToUserName><![CDATA[toUser]]></ToUserName><FromUserName><![CDATA[fromUser]]></FromUserName><CreateTime>1348831860</CreateTime><MsgType><![CDATA[text]]></MsgType><Content><![CDATA[你好]]></Content><MsgId>1234567890123456</MsgId></xml>'
```

### 3. 运行测试
```bash
go run test_wechat_message.go
```

## 消息处理逻辑

### 文本消息处理
- **"你好" 或 "hello"**: 返回欢迎消息
- **"帮助" 或 "help"**: 返回功能说明
- **"关于"**: 返回关于信息
- **其他内容**: 返回默认回复

### 其他消息类型
- **图片**: 返回图片接收确认
- **语音**: 返回语音接收确认（包含识别结果）
- **视频**: 返回视频接收确认
- **位置**: 返回位置信息确认
- **链接**: 返回链接信息确认

## 扩展功能

### 1. 添加新的关键词回复
在 `api/api.go` 文件的 `handleTextMessage` 函数中添加新的case：

```go
case strings.Contains(content, "新关键词"):
    return client.CreateTextReply(msg.FromUserName, msg.ToUserName, "新关键词的回复内容")
```

### 2. 添加数据库存储
可以在消息处理函数中添加数据库操作，记录用户消息：

```go
// 在handleTextMessage函数中添加
messageRecord := model.MessageRecord{
    FromUser: msg.FromUserName,
    Content:  msg.Content,
    Type:     "text",
    Time:     time.Now(),
}
messageRecord.Create(config.GVA_DB)
```

### 3. 添加媒体文件处理
对于图片、语音、视频等媒体文件，可以添加下载和处理逻辑：

```go
// 在handleImageMessage函数中添加
func handleImageMessage(msg *wechat.WXImageMessage, client *wechat.WechatClient) interface{} {
    // 下载图片
    // 保存到本地或云存储
    // 进行图片处理（如压缩、水印等）
    
    return client.CreateTextReply(msg.FromUserName, msg.ToUserName, "图片已处理完成！")
}
```

## 安全注意事项

1. **Token安全**: 确保Token的安全性，不要泄露给他人
2. **签名验证**: 所有来自微信的请求都会进行签名验证
3. **HTTPS**: 生产环境建议使用HTTPS协议
4. **IP白名单**: 可以在微信公众号后台设置IP白名单

## 故障排除

### 1. 签名验证失败
- 检查配置文件中的Token是否正确
- 确认微信公众号后台的Token设置
- 检查时间戳是否正确

### 2. 消息解析失败
- 检查XML格式是否正确
- 确认消息类型是否支持
- 查看服务器日志获取详细错误信息

### 3. 回复消息失败
- 检查XML序列化是否正确
- 确认消息格式符合微信规范
- 验证字符编码是否为UTF-8

## 相关文件

- `package/wechat/wechat.go`: 微信客户端核心功能
- `api/api.go`: 消息处理接口
- `router/router.go`: 路由配置
- `config/config.go`: 配置管理
- `config.yaml`: 配置文件
- `test_wechat_message.go`: 测试文件

## 更新日志

- **v1.0.0**: 初始版本，支持基本的消息接收和回复功能
- 支持文本、图片、语音、视频、位置、链接等多种消息类型
- 实现关键词回复和智能对话
- 添加完整的测试用例
