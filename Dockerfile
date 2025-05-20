# 使用官方的 Go 运行时作为构建环境
FROM golang:1.24.3-alpine AS builder

# 设置工作目录
WORKDIR /app

# 将本地代码复制到容器中
COPY . .

# 构建应用
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# 使用一个轻量级的运行时镜像
FROM alpine:latest

# 设置工作目录
WORKDIR /root/

# 从构建阶段复制二进制文件
COPY --from=builder /app/main .

RUN mkdir -p /root/etc && apk add --no-cache git git-lfs && rm -rf /var/cache/apk/*

# 暴露应用的端口
EXPOSE 8080 8081

# 运行应用
CMD ["./main"]