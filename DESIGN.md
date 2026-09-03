# 电影纪念票根 · 重构设计契约(DESIGN.md)

> 本文档是本次重构的唯一契约。后端(Go)与前端(Vue3)按此实现;实现与文档冲突时以本文档为准并回改文档。

## 1. 目标与范围

基于现有"电影纪念票根生成器"的功能面,重新设计前后端,要求页面美观。

**保留/重做的功能**
1. 票根编辑器:系统背景模板、海报上传、文字信息(标题/语言/格式/影院/影厅座位/时间/特殊ID)、字号调节、Canvas 实时预览合成、下载/微信内长按保存。
2. 账号体系:图形验证码 + 昵称/密码注册、登录、找回密码、JWT。
3. 成品:生成时上传成品 PNG 至对象存储并入库,个人"我的成品"分页浏览/删除。
4. 微信公众号:消息自动回复(关键词),"发送验证码获取 6 位码 → 输入兑换登录"(upsert 本地用户)。
5. 系统模板管理后台:管理员对 `template` 表增删改(软删)、状态开关、标题/正文颜色配置。

**本次新增(相对现状)**
- `user.role`(0=普通,1=管理员)、`user.openid`(公众号绑定)。
- 统一响应体 `{code,message,data}`、HTTP 状态码语义化(401 未登录/403 无权限)。
- 管理员模板管理接口 + 前端后台页。
- 微信验证码登录接口(`POST /api/wechat/login`)替代旧的"仅返回微信用户信息"。

**不迁移的旧设计**
- 旧的 statik 打包 → 改 stdlib `go:embed`。
- 旧的 `api/config/initialize/model/router/statik/types/package/web(静态页)` 目录 → 删除(重构完成后),git 历史保留。
- 旧上传"秒级文件名 + 协程内写响应"等缺陷 → 重构修复(毫秒+随机后缀、同步处理)。

## 2. 技术栈与仓库布局(同仓库重构)

| 层 | 选型 |
|---|---|
| 后端 | Go 1.24 + Gin;标准 `net/http` 服务;`go:embed` 内嵌前端产物 |
| 前端 | Vue 3 + TypeScript + Vite + Vue Router + Pinia(无重型 UI 库,自研设计系统) |
| 数据 | MySQL(GORM,单数表名,`AutoMigrate` 兼容现有库)+ 对象存储 MinIO(S3)/腾讯云 COS |
| 其它 | 密码 bcrypt(兼容旧 MD5 自动升级);JWT(HS256) |

```
ticket/
├─ cmd/server/main.go             入口:config → db → storage → router → listen
├─ internal/
│  ├─ assets/assets.go            //go:embed dist → http.FS(前端产物,SPA fallback)
│  ├─ config/config.go            配置加载(dev:config.yaml / prod:env)+ 全局单例
│  ├─ database/database.go        MySQL 连接 + AutoMigrate
│  ├─ model/                      user.go image.go template.go(GORM)
│  ├─ repository/                 数据访问(每表一个文件,收口 SQL)
│  ├─ service/                    业务(含验证码 manager、上传编排、微信)
│  ├─ handler/                    HTTP 处理(auth/user/template/image/admin/wechat/upload)
│  ├─ middleware/                 auth.go(admin/user), cors.go
│  ├─ router/router.go            路由装配 + SPA fallback + 404/405 JSON
│  └─ pkg/
│     ├─ response/response.go     统一响应
│     ├─ jwt/jwt.go
│     ├─ captcha/captcha.go       base64 图形验证码
│     ├─ storage/                 接口 + cos/  mos/ 实现
│     └─ wechat/wechat.go         公众号客户端(复用重构版)
├─ web/                           Vue3 源码(vite outDir → ../internal/assets/dist)
├─ scripts/build.ps1 / build.sh   一键:ui 构建 → go build
├─ config.yaml
├─ Dockerfile                     多阶段:node 构建 UI → go 构建 → alpine 运行
└─ README.md                      更新为新架构说明
```

**旧代码处理顺序**:绘制管线移植完成、新代码可编译运行前,保留旧 `web/static/js/script.js` 等作为移植参照;全部就绪后在最后一轮删除 `api/ config/ initialize/ model/ router/ statik/ types/ package/` 与旧 `web/` 静态页,并 `git rm` 提交。

## 3. 构建与运行

```bash
# UI 产物(开发时热更)
cd web && npm install --ignore-scripts && npm run dev          # http://localhost:5173, /api 代理到 :8080

# 后端(需先有 UI 产物,否则 assets 为占位页)
go run ./cmd/server                          # dev:读取 config.yaml;prod:RUN_MODE=prod + env
npm --prefix web run build                    # 产物输出到 internal/assets/dist
go build ./... && go vet ./...

# 生产镜像
docker build -t ticket .
```

- Vite `build.outDir = ../internal/assets/dist`,`emptyOutDir: true`;Vue Router 用 **history 模式**,后端对非 `/api` 请求回退 `index.html`(刷新不 404)。
- `internal/assets` 预置占位 `dist/index.html`(提示先构建 UI),保证 `go build` 始终可编译。

## 4. 数据模型(GORM,单数表名,兼容现有表)

```go
// user  (AutoMigrate 自动补列)
type User struct {
  Id            int    `gorm:"primaryKey;autoIncrement"`
  Nickname      string `gorm:"size:255;not null;uniqueIndex"`   // 登录名(保留旧 unique 语义,旧表无唯一索引需手工留意)
  Password      string `gorm:"size:100"`        // bcrypt;兼容 32 位 hex=旧 MD5,登录命中后自动升级
  Avatar        string `gorm:"size:500"`
  Openid        string `gorm:"size:64;default:''"`   // 公众号 openid(新)
  Role          int    `gorm:"default:0"`            // 0 普通 / 1 管理员(新)
  Status        int    `gorm:"default:1"`            // 1 正常 / 0 禁用
  LastLoginTime *int64
  CreateTime    int64
  UpdateTime    *int64
}

// image  (成品/海报/头像/模板图 统一)
type Image struct {
  Id         int    `gorm:"primaryKey;autoIncrement"`
  UserId     int    `gorm:"index"`
  Type       string `gorm:"size:20;index"`  // product|poster|avatar|template
  Filename   string `gorm:"size:200"`
  Url        string `gorm:"size:500"`
  Ip         string `gorm:"size:64"`
  CreateTime int64
}

// template
type Template struct {
  Id         int    `gorm:"primaryKey;autoIncrement"`
  UserId     int    `gorm:"default:0"`    // 0=系统模板
  Name       string `gorm:"size:100"`     // 展示名(新,可空)
  Url        string `gorm:"size:500;not null"`
  TitleColor string `gorm:"size:20"`
  TextColor  string `gorm:"size:20"`
  Status     int    `gorm:"default:1"`    // 1 上架 / 2 下架(软删)
  CreateTime int
}
```

管理员首启引导:环境变量 `APP_ADMIN_NICKNAME`(可逗号分隔)启动时把对应昵称 `role=1`(幂等);文档说明也可直接改库。

## 5. API 契约

**统一响应**:`{ "code": 0, "message": "ok", "data": {...} }`;`code≠0` 失败。
**HTTP 语义**:401 未登录/token 失效;403 非管理员或账号禁用;400 参数错;404 不存在;500 服务器错。响应体仍带上述 JSON。
**鉴权**:请求头 `Authorization: Bearer <token>`(登录接口返回 token);JWT claims 含 `userId, nickname, avatar, role, openid?`。
**时间**:一律毫秒或秒?沿用旧约定 **秒**(Unix),前端自行格式化。

### 公开
| 方法/路径 | 说明 |
|---|---|
| POST `/api/auth/captcha` | → `{captchaId, captchaImg(base64 png)}` |
| POST `/api/auth/register` | `{nickname,password,captchaId,captcha}` → 注册成功(≥3 位昵称,密码 ≥6 位) |
| POST `/api/auth/login` | `{nickname,password}` → `{token,user:{id,nickname,avatar,role}}` |
| POST `/api/auth/reset-password` | `{nickname,password,captchaId,captcha}` → 找回 |
| GET `/api/templates` | 系统上架模板列表 `{list:[{id,name,url,titleColor,textColor}], total}`(编辑器用,不分页) |
| POST `/api/wechat/login` | `{code}` 6 位公众号验证码 → 校验→openid→取用户信息→upsert 本地用户(昵称=微信昵称,头像=headimgurl)→ `{token,user,isNew}` |
| GET `/api/wechat/message` | 公众号服务器验证(签名) |
| POST `/api/wechat/message` | 公众号消息(文本/图片/语音/视频/位置/链接)→ XML 自动回复;含"验证码"关键词生成 6 位码 |

### 需登录(user)
| 方法/路径 | 说明 |
|---|---|
| GET `/api/auth/me` | 当前用户 `{id,nickname,avatar,role,openid,createTime}` |
| PUT `/api/auth/profile` | `{nickname?, avatar?}` 更新资料 |
| POST `/api/uploads` | multipart `file` + `type`(poster/product/avatar)→ `{url,filename,size}`(对象存储) |
| GET `/api/user/products` | `?page=1&size=12` → `{list:[image...], total}`(type=product,新→旧) |
| DELETE `/api/user/products/{id}` | 删除自己的成品记录与存储对象 |

### 需管理员(role=1)
| 方法/路径 | 说明 |
|---|---|
| GET `/api/admin/templates` | `?status=all|1|2` → 全量(默认 all) |
| POST `/api/admin/templates` | `{name,url,titleColor,textColor}` 新增 |
| PUT `/api/admin/templates/{id}` | `{name?,url?,titleColor?,textColor?,status?}` 编辑/上下架 |
| DELETE `/api/admin/templates/{id}` | 软删(status=2) |

## 6. 前端路由与页面

| 路由 | 页面 | 说明 |
|---|---|---|
| `/` | 编辑器 | 主界面(核心页) |
| `/login` `/register` `/forgot` | AuthView 单组件三态 | 卡片式登录/注册/找回 + 公众号验证码登录 tab |
| `/me` | 个人中心 | 头像(可上传)、昵称、资料编辑、快速入口 |
| `/products` | 我的成品 | 瀑布/网格画廊、灯箱预览、下载、删除;未登录引导 |
| `/admin/templates` | 模板管理(admin 守卫) | 表格/卡片管理 + 新增/编辑弹窗(传图→调 uploads type=template) |

导航:统一 AppHeader(品牌字标 + 链接 + 用户头像下拉:我的成品/个人中心/模板管理(仅 admin)/退出)。
路由守卫:需要登录的页面未登录跳 `/login?redirect=...`;admin 路由 403 提示。

## 7. 设计语言(暗色影院 × 金色)

**主题**:深夜片场氛围——近黑底、暖金高光、玻璃拟态面板、电影标题衬线。

Token(`web/src/styles/tokens.css`):
```
--bg-0:#0b0a0f;  --bg-1:#12111a;  --bg-2:#1a1824;
--surface:rgba(255,255,255,.045); --surface-2:rgba(255,255,255,.08);
--border:rgba(214,183,110,.18);   --border-strong:rgba(214,183,110,.45);
--gold:#d4af37; --gold-2:#f2dd96;  --gold-grad:linear-gradient(135deg,#f5e08c,#d4af37 55%,#9a7430);
--text:#efe9db; --text-muted:#a49d8c;
--danger:#e5685c; --ok:#5fbf8f;
--radius:16px; --radius-sm:10px;
--font-display:"Noto Serif SC","Source Han Serif SC","Songti SC","SimSun",serif;
--font-body:system-ui,-apple-system,"PingFang SC","Microsoft YaHei",sans-serif;
--shadow-glow:0 10px 40px rgba(212,175,55,.12);
```
细节:细腻噪点/星芒背景;标题字标用衬线+金色渐变;按钮主次两级(金色实心/玻璃描边);入场淡入上移;票券元素(裁切线、打孔圆)点缀;空状态用电影符号插画(纯 CSS/内联 SVG)。

**布局**:桌面两栏编辑器(左 预览舞台:暗色底+金色细框+聚光氛围;右 控制面板,分区卡片:模板·海报·信息·排版·操作);`≤900px` 单栏纵向,预览吸顶可选。整体最大宽 1280 居中。

**编辑器交互**:模板以缩略图宫格(加载中骨架);海报拖拽/点击上传 + 预览 + 移除;字号滑杆实时值;字段必填缺省给出优雅 toast;Canvas 预览 debounce 实时重绘;保存按钮三态(空闲/上传中/成功),成功后展示可点开的大图(灯箱)与"下载/微信长按保存"说明。

## 8. 票根绘制管线移植说明(编辑器核心)

- 画布 **700×1400**;模板为远程背景图(存储 URL,`crossOrigin=anonymous`);`titleColor/textColor` 来自模板记录。
- **源码基准**:旧 `web/static/js/script.js` 中 `drawBackground / drawPoster / drawTicketInfo / drawDecorations / drawCornerDecoration / drawGeometricDecoration / drawVintageDecoration / drawVintagePattern / drawElegantDecoration / renderTicket / isWeChatBrowser / saveForWeChat`。移植时**保留坐标/字体/间距语义 1:1**(这些是为现有模板图调校的),仅做类型化与清理,不得臆改布局数值;若需改观感只动"非模板图元素"并记录。
- 新结构:`web/src/lib/ticket/types.ts`(字段/默认值)、`drawing.ts`(纯函数管线,输入 state+assets 输出绘制)、`editor.ts`(DOM/事件/上传编排)。
- 微信内保存:canvas → PNG dataURL → 新开预览层,提示长按保存(微信会阻止 `a[download]`)。
- 非微信:Blob → objectURL → `<a download>`。
- 上传语义:海报选择后即异步传 `type=poster`(失败静默,不影响本地预览);点"生成票根":校验→上传 `type=product`(必登录,未登录先弹登录)→ 成功后提示并可下载;上传失败不阻断本地下载。

## 9. 配置与安全

`config.yaml`(dev)与 env(prod,`RUN_MODE=prod`)同结构:
```yaml
server: { port: 8080, mode: debug }
storage: { type: mos|cos, access_key_id, access_key_secret, endpoint, bucket, public_base: "" }
mysql: { path, username, password, dbname, config }
wechat: { app_id, app_secret, token }
jwt: { secret, expire_seconds: 86400 }
captcha: { expire_minutes: 10 }
admin: { bootstrap_nicknames: "" }
```
- bcrypt cost 10;旧 MD5 兼容 + 登录成功自动升级。
- 上传白名单图片类型与 ≤10MB;文件路径 `{type}/{yyyyMMdd}/{ts_ms}{rand}{ext}`,禁止路径穿越。
- 对象存储 URL 需支持 CORS(编辑器跨域取图);`mos`/`cos` 均提供 `Put` 与 `Delete`。
- 配置中含明文凭据属历史遗留,README 提示生产改用 env 并轮换密钥。

## 10. 验收清单

1. `go vet ./...`、`go build ./...` 零错误;`pnpm --dir web build` 零错误。
2. 本地起服:`curl` 冒烟登录/模板/上传(可选真实存储,失败降级仅验证路由与响应壳)。
3. SPA fallback:`GET /products` 返回 index.html;`GET /api/nope` 返回 JSON 404。
4. 页面:五类页面可访问,editor 能画模板+海报+文字并下载;暗色金色视觉统一、无横向滚动。
5. 文档:README 重写(架构/API 表/构建/配置/首次运行),DESIGN.md 归档说明。

## 11. 里程碑

M1 后端骨架可编译(本契约 + config/model/pkg/db 就绪)
M2 后端全接口 + 中间件 + embed 静态服务,冒烟通过
M3 前端脚手架 + 设计系统 + 路由/守卫
M4 编辑器页(移植绘制管线)
M5 其余页面(账号/个人/成品/管理后台/微信登录)
M6 集成构建、Docker、文档、端到端验收、清理旧代码
