# 使用阿里云的 Golang 基础镜像
FROM registry.cn-shanghai.aliyuncs.com/library/golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o essay-stateless ./main.go

# 使用阿里云的 Alpine 基础镜像
FROM registry.cn-shanghai.aliyuncs.com/library/alpine:latest
RUN apk --no-cache add ca-certificates tzdata

# 可选：配置 Alpine 使用阿里云源（加速 apk 安装）
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories

WORKDIR /root/
COPY --from=builder /app/essay-stateless .

EXPOSE 8090
ENV TZ=Asia/Shanghai
ENV GIN_MODE=release

CMD ["./essay-stateless"]
