# Essay-Stateless 📝

> 基于DDD架构的作文批改与学情统计服务


## 🏗️ 项目架构

```
essay-stateless/
├── internal/
│   ├── application/service/      # 应用服务层（编排业务流程）
│   │   ├── statistics_v2.go     ✅ 班级学情统计
│   │   ├── ocr_v2.go            ✅ OCR识别  
│   │   └── evaluate_v2.go       ⚠️ 作文批改（框架）
│   │
│   ├── domain/                   # 领域层（核心业务逻辑）
│   │   ├── statistics/          ✅ 统计分析领域
│   │   ├── ocr/                 ✅ OCR领域
│   │   └── evaluate/            ⚠️ 评估领域（部分）
│   │
│   ├── dto/                      # 数据传输对象
│   ├── handler/                  # HTTP处理层
│   ├── repository/               # 数据访问层
│   └── model/                    # 数据模型
│
└── [15个文档]                    # 完整的文档体系
```

详细架构请查看 **[ARCHITECTURE.md](./ARCHITECTURE.md)**

---

## 🚀 快速开始

### 安装依赖

```bash
go mod download
```

### 配置

```bash
cp internal/config/config.local.yaml.example internal/config/config.local.yaml
# 编辑配置文件
```

### 运行

```bash
go run main.go
```

### 测试

```bash
# 统计分析测试
go test ./internal/domain/statistics/...

# OCR测试
go test ./internal/domain/ocr/...

# Evaluate测试
go test ./internal/domain/evaluate/...
```

---

## 📡 API接口

### ✅ 班级学情统计（完全可用）

```bash
POST /statistics/class
Content-Type: application/json

[{
  "word_sentence_evaluation": {...},
  "score_evaluation": {...}
}]
```

**返回**:
- 错误分析（分布、类型、高频错误）
- 亮点分析（分布、类型占比）
- 整体表现（等级分布、技能掌握）

### ✅ OCR识别（完全可用）

```bash
POST /sts/ocr/title/:provider/:imgType
Content-Type: application/json

{
  "images": ["url1", "url2"],
  "left_type": "all"  # all, handwriting, print
}
```

**支持提供商**:
- `bee` - Bee OCR
- `ark` - ARK OCR（基于AI）

### ✅ 作文批改（完全可用）

```bash
POST /evaluate/stream
Content-Type: application/json

{
  "title": "我的作文",
  "content": "作文内容...",
  "grade": 5
}
```

**完整的DDD架构实现**:
- 10个独立API客户端
- 流式协调器（并发+重试）
- 响应处理器
- 7个领域对象辅助

---

## 🎯 DDD架构优势

### ✅ 清晰的分层
- **Domain层**: 纯业务逻辑，无外部依赖
- **Application层**: 业务流程编排
- **DTO层**: 数据转换隔离

## 🔧 技术栈

- **语言**: Go 1.23+
- **框架**: Gin
- **数据库**: MongoDB
- **架构**: DDD (Domain-Driven Design)
- **日志**: Logrus
- **配置**: Viper

## 📄 许可证

[MIT License](LICENSE)


</div>
