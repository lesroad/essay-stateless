FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o essay-stateless ./main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/essay-stateless .

EXPOSE 8090

ENV TZ=Asia/Shanghai
ENV GIN_MODE=release

CMD ["./essay-stateless"] 