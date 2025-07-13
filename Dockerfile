FROM registry.cn-shanghai.aliyuncs.com/lesroad/infrastructure:golang_1.23-alpine AS builder

LABEL stage=gobuilder

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

# 修改 Alpine 包管理工具 apk 的软件源为阿里云镜像，加速后续依赖安装（国内环境优化）
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

# apk update：更新软件源索引（--no-cache 不缓存索引文件，减小层体积）
# apk add tzdata：安装时区数据（用于后续设置容器时区）。
# apk add ca-certificates：安装 CA 证书（用于 HTTPS 通信，如访问 HTTPS 接口）。
RUN apk update --no-cache && apk add --no-cache tzdata ca-certificates

# 设置当前工作目录为 /build，后续命令均在此目录执行。
WORKDIR /build

# 优化构建缓存
ADD go.mod .
ADD go.sum .
RUN go mod download

# 将当前目录（宿主机构建上下文）的所有文件复制到容器的 /build 目录（包括代码、配置等）。
COPY . .

# 直接编译Go程序
RUN go build -ldflags="-s -w" -o app main.go

FROM registry.cn-shanghai.aliyuncs.com/lesroad/infrastructure:alpine_3.18

# 安装必要的运行时依赖
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# 复制二进制文件和配置文件
COPY --from=builder /build/app /app/
COPY --from=builder /build/internal/config/config.yaml /app/internal/config/
COPY --from=builder /build/static /app/static

# 设置时区
ENV TZ=Asia/Shanghai
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone

# 创建非root用户（安全最佳实践）
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# 改变文件所有权
RUN chown -R appuser:appgroup /app

# 切换到非root用户
USER appuser

# 暴露端口8090
EXPOSE 8090

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8090/test || exit 1

CMD ["./app"]
