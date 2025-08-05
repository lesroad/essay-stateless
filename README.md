# essay-stateless

基于Gin框架的Go语言无状态服务，提供作文评估、OCR识别和计费功能。

## 功能特性

- **作文评估服务**：提供作文内容的评估和分析功能
- **OCR识别服务**：
  - 常规OCR识别
  - 标题OCR识别
- **计费服务**：用户使用量跟踪和计费功能
- **监控追踪**：
  - 请求日志记录
  - 分布式追踪(OpenTelemetry)

## 技术栈

- **编程语言**: Go
- **Web框架**: Gin
- **数据库**: MongoDB
- **追踪系统**: OpenTelemetry
- **配置管理**: 自定义配置加载

## 项目结构

```
stateless-go/
├── internal/       # 内部实现代码
│   ├── config/     # 配置管理
│   ├── handler/    # HTTP处理器
│   ├── middleware/ # 中间件
│   ├── repository/ # 数据访问层
│   └── service/    # 业务逻辑
├── pkg/            # 可复用组件
│   ├── database/   # 数据库连接
│   ├── logger/     # 日志
│   └── trace/      # 追踪
├── main.go         # 程序入口
└── Dockerfile      # 容器化配置
```

## API文档

### 评估服务
- `POST /evaluate` - 提交作文内容进行评估

### OCR服务
- `POST /sts/ocr` - 常规OCR识别
- `POST /sts/ocr/title` - 标题OCR识别

### 计费服务
- `POST /billing` - 计费相关操作

## 部署说明

### 容器化部署
1. 构建Docker镜像：
   ```bash
   docker build -t essay-stateless .
   ```
2. 运行容器：
   ```bash
   docker run -p 8080:8080 --env-file .env essay-stateless
   ```

### 环境变量
- `MONGO_URI`: MongoDB连接字符串
- `OTEL_EXPORTER_OTLP_ENDPOINT`: OpenTelemetry收集器地址

## 开发指南

### 环境要求
- Go 1.18+
- MongoDB 4.4+
- Docker (可选)

### 运行本地开发
1. 安装依赖：
   ```bash
   go mod download
   ```
2. 启动服务：
   ```bash
   go run main.go
   ```

### 贡献指南
1. Fork项目
2. 创建特性分支
3. 提交Pull Request