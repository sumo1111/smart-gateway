# 🚀 Smart Gateway

**智能 LLM API 网关** — 统一多 Provider 接入，Auto 模型智能路由，开箱即用。

[![Go](https://img.shields.io/badge/Go-1.22-00ADD8?style=flat&logo=go)](https://go.dev/)
[![React](https://img.shields.io/badge/React-18-61DAFB?style=flat&logo=react)](https://react.dev/)
[![Docker](https://img.shields.io/badge/Docker-Ready-2496ED?style=flat&logo=docker)](https://www.docker.com/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## ✨ 核心特性

- **🔄 统一 API** — 兼容 OpenAI API 格式，一个端点接入所有模型
- **🧠 Auto 模型** — `model="auto"` 自动选择最优模型 + 渠道，基于成功率/延迟/权重实时评分
- **📡 多 Provider** — 支持 OpenAI、Anthropic Claude、自定义端点，轻松扩展
- **🎯 智能路由** — 三种策略：加权随机 / 最低延迟 / 轮询，自动 failover
- **🛡️ 自愈能力** — 连续失败自动屏蔽，定时健康检测，评分自动恢复
- **📊 可观测性** — 请求日志、成功率统计、延迟监控、模型评分看板
- **🔑 令牌管理** — API Key 分发、额度控制、一键刷新
- **📦 单文件部署** — SQLite 内嵌，零依赖，Docker / Windows exe 一键启动

## 🏗️ 架构

```
┌─────────────┐     ┌──────────────────┐     ┌──────────────┐
│   Client     │────▶│  Smart Gateway   │────▶│  Provider A  │
│  (OpenAI SDK)│     │                  │     │  (OpenAI)    │
└─────────────┘     │  ┌────────────┐  │     ├──────────────┤
                    │  │ Auto Router│  │────▶│  Provider B  │
                    │  │  (Scoring) │  │     │  (Anthropic) │
                    │  └────────────┘  │     ├──────────────┤
                    │  ┌────────────┐  │     │  Provider C  │
                    │  │  Relay     │  │────▶│  (Custom)    │
                    │  │  Adaptor   │  │     └──────────────┘
                    │  └────────────┘  │
                    │  ┌────────────┐  │
                    │  │  SQLite DB │  │
                    │  └────────────┘  │
                    └──────────────────┘
```

## 🚀 快速开始

### Docker 部署（推荐）

```bash
docker run -d \
  --name smart-gateway \
  -p 3000:3000 \
  -e ADMIN_PASSWORD=your_password \
  -v sg-data:/app/data \
  sumo1111/smart-gateway:latest
```

### Docker Compose

```bash
git clone https://github.com/sumo1111/smart-gateway.git
cd smart-gateway
docker compose up -d
```

### Windows exe

1. 从 [Releases](https://github.com/sumo1111/smart-gateway/releases) 下载 `smart-gateway.exe`
2. 双击运行或命令行启动：

```powershell
set ADMIN_PASSWORD=your_password
smart-gateway.exe
```

3. 浏览器打开 `http://localhost:3000`

### 从源码构建

```bash
# 前端
cd web && npm install && npm run build && cd ..

# 后端
go mod tidy
make build-linux    # Linux
make build-windows  # Windows (需要 mingw-w64)
./smart-gateway
```

## 📖 使用说明

### 1. 登录管理面板

访问 `http://localhost:3000`，默认密码 `admin123`（通过 `ADMIN_PASSWORD` 环境变量修改）

### 2. 添加渠道

在「渠道管理」页面添加你的 API Provider：
- **类型**：OpenAI 兼容 / Anthropic Claude / 自定义端点
- **Base URL**：Provider 的 API 地址
- **API Key**：你的密钥
- **模型列表**：该渠道支持的模型名称

### 3. 创建令牌

在「令牌管理」页面创建访问令牌，获得 `sk-xxxx` 格式的 Key。

### 4. 调用 API

```python
from openai import OpenAI

client = OpenAI(
    api_key="sk-your-token-key",
    base_url="http://localhost:3000/v1"
)

# 指定模型
response = client.chat.completions.create(
    model="gpt-4o",
    messages=[{"role": "user", "content": "Hello!"}]
)

# ✨ Auto 模型 — 自动选择最优模型和渠道
response = client.chat.completions.create(
    model="auto",
    messages=[{"role": "user", "content": "Hello!"}]
)
```

### Auto 路由算法

当 `model="auto"` 时，网关自动选择最优的模型+渠道组合：

```
score = success_rate × 0.5 + (1 - avg_latency / timeout) × 0.3 + weight × 0.2
```

| 策略 | 说明 |
|------|------|
| `weighted` | 加权随机，按 score 概率选择（默认） |
| `lowest_latency` | 最低延迟优先，选 score 最高的 |
| `round_robin` | 简单轮询，均匀分配 |

**自愈机制**：连续 3 次失败 → 屏蔽 5 分钟 → 评分自动衰减/恢复

## ⚙️ 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `PORT` | 3000 | 服务端口 |
| `ADMIN_PASSWORD` | admin123 | 管理员密码 |
| `DB_PATH` | data/smart-gateway.db | SQLite 数据库路径 |
| `AUTO_STRATEGY` | weighted | 路由策略 |
| `HEALTH_INTERVAL` | 60 | 健康检测间隔（秒） |
| `FAIL_BAN_COUNT` | 3 | 连续失败屏蔽阈值 |
| `FAIL_BAN_DURATION` | 300 | 屏蔽时长（秒） |
| `REQUEST_TIMEOUT` | 120 | 请求超时（秒） |

## 🔧 API 参考

### Relay（兼容 OpenAI）

```
POST /v1/chat/completions   # Chat 补全
POST /v1/completions        # 文本补全
POST /v1/embeddings         # 向量嵌入
GET  /v1/models             # 模型列表
```

### 管理接口

```
GET/POST/PUT/DELETE  /api/channel      # 渠道 CRUD
GET/POST/PUT/DELETE  /api/token        # 令牌 CRUD
GET                   /api/log          # 请求日志
GET                   /api/stats        # 统计信息
GET                   /api/auto/status  # Auto 路由状态
POST                  /api/auto/strategy # 修改路由策略
POST                  /api/auto/refresh  # 刷新评分
POST                  /api/auto/probe    # 全渠道健康检测
```

## 🛠️ 技术栈

| 层 | 技术 |
|----|------|
| 后端 | Go 1.22 + Gin + SQLite |
| 前端 | React 18 + TypeScript + Ant Design 5 + Recharts |
| 部署 | Docker (Alpine) / Windows exe (交叉编译) |

## 📜 License

[MIT](LICENSE)

## 🙏 致谢

灵感来源于 [one-api](https://github.com/songquanpeng/one-api) 和 [one-hub](https://github.com/MartialBE/one-hub)，精简核心功能并加入 Auto 智能路由。
