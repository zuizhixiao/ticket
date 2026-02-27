FROM registry.cn-hangzhou.aliyuncs.com/open_images/go-alpine:1.24.1 AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOPROXY https://goproxy.cn,direct

WORKDIR /var/www

ADD . .

RUN go mod tidy

# 安装statik工具
RUN go install github.com/rakyll/statik 

RUN statik -src="./web" -f

RUN go build

FROM registry.cn-hangzhou.aliyuncs.com/open_images/alpine

ENV TZ Asia/Shanghai

WORKDIR /app/project

COPY --from=builder /var/www ./

EXPOSE 8080

RUN chmod +x /app/project/ticket

CMD ["./ticket"]
