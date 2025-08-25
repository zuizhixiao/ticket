# 电影纪念票根生成器

一个基于Go语言和HTML5 Canvas的电影纪念票根生成器，支持自定义模板、海报上传和票根信息编辑。

## 功能特性

- 🎬 多种票根模板选择
- 🖼️ 海报图片上传和预览
- ✏️ 自定义票根信息（电影标题、影院、时间等）
- 💾 本地下载票根图片
- ☁️ **新增：自动上传票根图片至服务端**
- 📱 微信浏览器兼容性支持

## 技术架构

- **后端**: Go + Gin框架
- **前端**: HTML5 + JavaScript + Canvas
- **静态资源**: 使用statik打包前端资源
- **图片处理**: 支持PNG、JPG、GIF、BMP、WebP格式

## 新增功能：图片上传至服务端

### 功能说明
当用户下载票根时，系统会自动将生成的票根图片上传至服务端存储，同时保持原有的本地下载功能。

### 技术实现
- **API接口**: `POST /api/upload/image`
- **文件存储**: 自动创建`uploads/`目录存储图片
- **文件命名**: 使用时间戳生成唯一文件名，避免冲突
- **格式支持**: PNG、JPG、JPEG、GIF、BMP、WebP
- **文件大小**: 最大支持10MB

### 上传流程
1. 用户点击下载按钮
2. 系统验证必填字段和海报上传
3. **自动上传票根图片至服务端**
4. 显示上传成功提示
5. 继续执行本地下载功能

### 错误处理
- 上传失败不影响本地下载功能
- 支持的文件格式验证
- 文件大小限制
- 服务端错误友好提示

## 快速开始

### 环境要求
- Go 1.24.1+
- 现代浏览器（支持HTML5 Canvas）

### 安装运行
```bash
# 克隆项目
git clone <repository-url>
cd ticket

# 安装依赖
go mod tidy

# 运行服务
go run main.go
```

### 访问应用
打开浏览器访问 `http://localhost:8080`

## 使用说明

1. **选择模板**: 点击模板缩略图选择票根样式
2. **上传海报**: 点击"上传海报"按钮选择电影海报图片
3. **填写信息**: 输入电影标题、影院名称、时间等信息
4. **生成票根**: 点击"下载票根"按钮
5. **自动上传**: 系统自动将票根上传至服务端
6. **本地下载**: 同时下载票根到本地设备

## 项目结构

```
ticket/
├── api/           # API接口
│   └── api.go     # 图片上传处理
├── initialize/    # 服务初始化
│   └── router.go  # 路由配置
├── router/        # 路由定义
│   └── router.go  # API路由注册
├── web/           # 前端资源
│   ├── index.html # 主页面
│   └── static/    # 静态资源
├── uploads/       # 图片上传目录（自动创建）
├── main.go        # 程序入口
└── go.mod         # Go模块配置
```

## API接口

### 图片上传
- **URL**: `POST /api/upload/image`
- **参数**: `image` (multipart/form-data)
- **响应**: JSON格式的上传结果

#### 成功响应示例
```json
{
  "success": true,
  "message": "图片上传成功",
  "data": {
    "filename": "ticket_1703123456789123456.png",
    "filePath": "./uploads/ticket_1703123456789123456.png",
    "size": 1024000,
    "type": "image/png"
  }
}
```

#### 错误响应示例
```json
{
  "success": false,
  "message": "只支持图片文件上传"
}
```

## 注意事项

- 确保`uploads/`目录有写入权限
- 图片文件会自动按时间戳命名，避免文件名冲突
- 上传失败不会影响本地下载功能
- 支持的最大文件大小为10MB

## 开发说明

### 添加新的图片格式支持
在`api/api.go`中修改`allowedExts`数组：
```go
allowedExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".tiff"}
```

### 修改上传目录
在`api/api.go`中修改`uploadDir`变量：
```go
uploadDir := "./custom_uploads"
```

### 调整文件大小限制
在`api/api.go`中修改`ParseMultipartForm`参数：
```go
c.Request.ParseMultipartForm(20 << 20) // 20MB
```

## 许可证

MIT License
