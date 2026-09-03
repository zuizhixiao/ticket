# 🎬 电影纪念票根生成器(Cinema Ticket)

在线生成"电影纪念票根"的 Web 应用:选择背景模板、上传电影海报、定制文字信息,前端 Canvas 实时合成 700×1400 票根图,一键保存到云端"我的成品"并下载到本地;支持微信内长按保存与公众号验证码登录,运营者可维护系统模板。

本仓库为 2026 年的全面重构版本:**Go(Gin)分层后端 + Vue 3 暗色影院风格前端,单二进制内嵌前端产物**。旧版代码可从 git 历史回溯。

## ✨ 功能

- **票根编辑器**:系统模板宫格、海报拖拽/点击上传、语言/格式/影院/影厅座位/时间/专属 ID 编辑、标题与正文字号滑杆、Canvas 实时预览
- **账号体系**:图形验证码注册 / 昵称登录 / 找回密码 / JWT;密码 bcrypt(兼容旧 MD5 自动升级)
- **成品管理**:生成票根自动上传对象存储并入"我的成品",支持分页浏览、灯箱大图、下载、删除
- **微信**:公众号消息自动回复;发送"验证码"获取 6 位码 → 网页输入兑换登录(自动绑定 openid)
- **模板管理后台**:管理员新增/编辑/上架/下架系统模板(上传背景图 + 标题/正文颜色)
- **本地零依赖开发**:支持纯 Go SQLite(见下),无需安装 MySQL 即可起服务

## 🧱 技术栈

| 层 | 选型 |
|---|---|
| 后端 | Go 1.24 · Gin · GORM(MySQL / SQLite) · go:embed |
| 对象存储 | MinIO/S3 兼容(`mos`)与腾讯云 COS(`cos`)双实现 |
| 前端 | Vue 3 · TypeScript · Vite · Vue Router · Pinia(自研设计系统,无重型 UI 库) |
| 其他 | base64 图形验证码 · 公众号 XML 消息处理 · JWT(HS256) |

## 📁 目录结构

```
cmd/server/            入口
internal/
  assets/              go:embed 内嵌前端产物(internal/assets/dist)
  config/              配置加载(yaml / 环境变量)
  database/            MySQL 或 SQLite 连接与 AutoMigrate
  model/               user / image / template(GORM)
  repository/          数据访问
  service/             业务编排(密码、验证码、上传、模板、微信登录)
  handler/             HTTP 处理
  middleware/          JWT / AdminOnly / CORS
  router/              路由 + SPA 回退 + 静态资源
  pkg/                 response / jwt / captcha / storage / wechat
web/                   Vue 3 前端(Vite 产物输出至 ../internal/assets/dist)
scripts/               build.ps1 / build.sh 一键构建
Dockerfile             多阶段:node 构建 UI → Go 编译 → alpine 运行
DESIGN.md              重构契约(架构/API/数据模型/设计语言)
```

## 🚀 快速开始

### 本地开发

```bash
# 1) 后端(默认读 config.yaml;可用 SQLite 免装 MySQL)
cp config.example.yaml config.yaml   # 并填入你的配置;或用 sqlite
$env:MYSQL_DRIVER='sqlite'           # PowerShell;bash 用 export
$env:MYSQL_PATH='./data.db'
go run ./cmd/server                  # http://127.0.0.1:8080

# 2) 前端(热更新)
cd web && npm install --ignore-scripts && npm run dev # http://127.0.0.1:5173, /api 代理到 8080
```

### 构建发布(单二进制,UI 内嵌)

```bash
cd web && npm install --ignore-scripts && npm run build   # 产物进 internal/assets/dist
go build -o ticket ./cmd/server
./ticket                                          # 依赖 config.yaml 或环境变量
# 或一键脚本:scripts/build.ps1 / scripts/build.sh
```

### Docker

```bash
docker build -t ticket .
docker run --rm -p 8080:8080 \
  -e RUN_MODE=prod \
  -e MYSQL_DRIVER=mysql -e MYSQL_PATH=127.0.0.1:3306 -e MYSQL_USERNAME=root \
  -e MYSQL_PASSWORD=xxx -e MYSQL_DBNAME=ticket \
  -e STORAGE_TYPE=mos -e STORAGE_ENDPOINT=minio.example.com -e STORAGE_ACCESS_KEY_ID=xxx \
  -e STORAGE_ACCESS_KEY_SECRET=xxx -e STORAGE_BUCKET=dvr \
  -e JWT_SECRET=change-me \
  ticket
```

## ⚙️ 配置

`config.example.yaml` 为模板。加载优先级:**默认值 < config.yaml(可选) < 环境变量**;`RUN_MODE=prod` 时依赖环境变量(生产不落盘密钥)。

| 段 | 环境变量(示例) | 说明 |
|---|---|---|
| server | `SERVER_PORT` `SERVER_MODE` | 监听端口 / gin debug·release |
| mysql | `MYSQL_DRIVER` `MYSQL_PATH` `MYSQL_USERNAME` `MYSQL_PASSWORD` `MYSQL_DBNAME` `MYSQL_CONFIG` | driver=`mysql`(默认)或 `sqlite`(Path=文件) |
| storage | `STORAGE_TYPE` `STORAGE_ACCESS_KEY_ID` `STORAGE_ACCESS_KEY_SECRET` `STORAGE_ENDPOINT` `STORAGE_BUCKET` | `mos`=MinIO/S3,`cos`=腾讯云 |
| wechat | `WECHAT_APP_ID` `WECHAT_APP_SECRET` `WECHAT_TOKEN` | 公众号 |
| jwt | `JWT_SECRET` `JWT_EXPIRE_SECONDS` | HS256 密钥与有效期(秒) |
| admin | `ADMIN_BOOTSTRAP_NICKNAMES` | 启动时将昵称(逗号分隔,已注册用户)提升为管理员 |

> 数据库迁移:首次启动 `AutoMigrate` 为 `user` 表补充 `role`、`openid` 列,为 `image` 表补充 `object` 列。若线上 MySQL 磁盘不足会导致迁移失败(先清理空间)。

## 🔌 API 摘要

统一响应 `{ "code": 0, "message": "ok", "data": … }`;鉴权头 `Authorization: Bearer <token>`。

- 公开:`POST /api/auth/captcha|register|login|reset-password`、`GET /api/templates`、`GET|POST /api/wechat/message`、`POST /api/wechat/login`
- 登录:`GET /api/auth/me`、`PUT /api/auth/profile`、`POST /api/uploads`(multipart: file+type)、`GET /api/user/products`、`DELETE /api/user/products/:id`
- 管理员:`GET|POST /api/admin/templates`、`PUT|DELETE /api/admin/templates/:id`

## 🗄️ 数据模型(单数表名)

- `user`:nickname、password(bcrypt,旧 MD5 兼容)、avatar、openid、role(0 普通/1 管理员)、status、时间戳
- `image`:userId、type(product/poster/avatar/template)、filename、url、object(存储 key)、ip、createTime
- `template`:userId(0=系统)、name、url、titleColor、textColor、status(1 上架/2 下架)、createTime

## 🧪 冒烟验证

```powershell
$env:MYSQL_DRIVER='sqlite'; $env:MYSQL_PATH='./smoke.db'
go run ./cmd/server   # 另开终端:
curl http://127.0.0.1:8080/api/auth/captcha -X POST   # code=0
curl http://127.0.0.1:8080/                           # 首页(构建 UI 后为应用)
curl http://127.0.0.1:8080/products                   # SPA 回退
curl http://127.0.0.1:8080/api/nope                    # JSON 404
```

## 📜 设计

暗色影院 × 金色设计语言、绘制管线坐标与页面交互细节见 `DESIGN.md`。票根画布 700×1400 的模板布局数值沿用既有模板调校结果,移植保持 1:1。

## 📄 License

MIT
