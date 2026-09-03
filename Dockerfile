# 多阶段构建:node 构建 Vue 产物 → Go 编译单二进制(内嵌前端)。
# 生产运行配置走环境变量(参见 config.example.yaml / README)。
#
# 底包使用国内可达镜像源:registry.cn-hangzhou.aliyuncs.com/library/*
# (阿里云容器镜像服务对 Docker Hub 官方库的同步镜像)。
# 若你的构建环境可直连 Docker Hub,去掉前缀即可使用官方镜像:
#   node:22-alpine / golang:1.24-alpine / alpine:3.20

FROM registry.cn-hangzhou.aliyuncs.com/open_images/node:22-alpin AS ui
WORKDIR /src
COPY . .
WORKDIR /src/web
# --ignore-scripts:npm 11 默认拦截 esbuild 安装脚本;二进制由可选依赖提供,不影响构建
RUN npm ci --ignore-scripts \
    && npm run build

FROM registry.cn-hangzhou.aliyuncs.com/open_images/go-alpine:1.24.1 AS builder
WORKDIR /build
COPY --from=ui /src/go.mod ./go.mod
COPY --from=ui /src/go.sum ./go.sum
COPY --from=ui /src/cmd ./cmd
COPY --from=ui /src/internal ./internal
ENV CGO_ENABLED=0 GOPROXY=https://goproxy.cn,direct
RUN go mod download \
    && go build -ldflags="-s -w" -o ticket ./cmd/server

FROM registry.cn-hangzhou.aliyuncs.com/open_images/alpine
RUN apk add --no-cache ca-certificates tzdata \
    && cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
WORKDIR /app
COPY --from=builder /build/ticket ./ticket
EXPOSE 8080
ENTRYPOINT ["./ticket"]
