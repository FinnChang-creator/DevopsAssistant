# DevopsAssistant

基于 Eino 框架的智能知识索引系统，使用 Redis Stack 实现高效的向量搜索和文档管理。

## 特性

- **向量索引**：基于 Redis Stack 的 HNSW 算法，提供高性能向量搜索
- **文档处理**：智能文档分块，支持 Markdown 格式
- **灵活配置**：YAML 配置文件管理，支持多种嵌入服务
- **批量处理**：优化的批量处理机制，减少 API 调用
- **可扩展**：基于 Eino Graph 模型，支持复杂流程编排

## 技术栈

- **Go 1.25+**：主要编程语言
- **Eino**：AI 应用流水线框架
- **Redis Stack**：向量存储和搜索
- **Dashscope**：阿里云通义千问嵌入服务
- **GoFrame**：配置管理和日志

## 快速开始

### 环境要求

- Go 1.20+
- Redis Stack 7.0+
- 网络连接（用于调用 Embedding API）

### 安装

```bash
# 克隆仓库
git clone https://github.com/FinnChang-creator/DevopsAssistant.git
cd DevopsAssistant

# 安装依赖
go mod download

# 配置
cp config/config.yaml.example config/config.yaml
# 编辑 config/config.yaml，配置 Redis 和 Embedding 服务

# 运行
cd cmd && go run main.go
```

### 配置

编辑 `config/config.yaml` 文件：

```yaml
# Redis 连接配置
redis:
  addr: "localhost:6379"
  password: ""
  db: 0

# Embedding 服务配置
embedding_model:
  choose: "dashscope"
  batch_size: 100
  key_prefix: "default_vec_idx:"
  dashscope:
    api_key: "your-api-key"
    model: "text-embedding-v4"
    dimensions: 1024

# 向量索引配置
redis_vector_indexes:
  vector_index:
    index_name: "default_vec_idx"
    vector_field:
      dim: 1024
      algorithm: "HNSW"
```

## 使用

### 添加文档

将 Markdown 文档放入 `docs` 目录，系统会自动处理并建立索引。

```bash
# 查看索引状态
./scripts/list.sh
```

## 文档

- [架构设计](docs/架构设计.md)
- [使用指南](docs/使用指南.md)
- [API 参考](docs/API参考.md)

## 项目结构

```
DevopsAssistant/
├── cmd/
│   └── main.go                    # 入口点
├── internal/
│   ├── agent/                     # 业务逻辑层（智能代理）
│   │   └── knowledge/
│   │       ├── knowledge.go       # 知识代理编排
│   │       ├── indexer.go         # 索引器
│   │       ├── loader.go          # 文档加载
│   │       └── transformer.go     # 文档转换
│   ├── config/                    # 配置管理
│   │   ├── config.go              # 配置结构体
│   │   └── loader.go              # 配置加载
│   ├── external/                  # 外部服务层
│   │   └── embedding/
│   │       └── dashscope.go       # Dashscope 实现
│   ├── infrastructure/            # 基础设施层
│   │   ├── redis/
│   │   │   ├── client.go          # Redis 客户端
│   │   │   └── index.go           # Redis 索引
│   │   └── logger/
│   │       └── logger.go          # 日志初始化
│   └── pkg/                       # 工具包
│       └── callback/
│           └── callback.go
├── bootstrap/                     # 系统初始化
│   └── init.go
├── config/
│   └── config.yaml                # 配置文件
├── docs/                          # 文档目录
├── scripts/
│   └── list.sh                    # 索引查看脚本
├── go.mod
└── go.sum
```

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

## 联系方式

- GitHub: https://github.com/FinnChang-creator/DevopsAssistant
