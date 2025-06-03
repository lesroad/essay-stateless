# 第一阶段：构建阶段
FROM golang:1.23-alpine AS builder

# 设置环境变量
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOPROXY=https://goproxy.cn,direct

# 更换阿里云镜像源（适合阿里云ACR构建）
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 安装必要的构建工具
RUN apk update --no-cache && apk add --no-cache git ca-certificates tzdata

# 设置工作目录
WORKDIR /build

# 复制go模块文件并下载依赖
COPY go.mod go.sum ./
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN go build -ldflags="-s -w" -o main .

# 第二阶段：运行阶段
FROM alpine:latest

# 更换阿里云镜像源
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# 安装ca-certificates和tzdata
RUN apk --no-cache add ca-certificates tzdata

# 设置时区
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# 创建非root用户
RUN addgroup -g 1001 -S appgroup && \
    adduser -S appuser -u 1001 -G appgroup

# 设置工作目录
WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/main ./

# 复制配置文件目录
COPY --from=builder /build/configs ./configs

# 更改文件所有权
RUN chown -R appuser:appgroup /app && \
    chmod +x /app/main

# 切换到非root用户
USER appuser

# 暴露端口
EXPOSE 8090

# 设置健康检查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8090/ || exit 1

# 启动应用
CMD ["./main"] 