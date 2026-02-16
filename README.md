# LogForge — 智能网络日志格式化系统
<img width="427" height="479" alt="image" src="https://github.com/user-attachments/assets/83637c52-a03e-463d-ab97-cc47b2ecb5bd" />

<img width="1044" height="695" alt="image" src="https://github.com/user-attachments/assets/873cc306-6768-4f87-8d4f-5fce3472487b" />


LogForge 是一款基于 LLM（大语言模型）的桌面应用，能够自动分析网络日志样本、生成 Python 解析代码，并批量处理日志文件导出为 Excel。

## 功能特性

- **AI 驱动的代码生成**：粘贴日志样本，LLM 自动生成完整的 Python 解析程序
- **自动代码验证**：语法检查 + LLM 自动修复（最多 3 次重试）
- **批量处理**：一键处理整个目录的日志文件，合并输出到单个 Excel 文件
- **运行时错误恢复**：检测执行失败后自动调用 LLM 修复代码并重试
- **项目管理**：历史项目持久化存储，支持查看、编辑代码、重新执行
- **Python 环境隔离**：通过 [uv](https://docs.astral.sh/uv/) 自动创建独立虚拟环境
- **实时进度监控**：批量处理时通过 JSON stdout 实时显示进度

## 技术栈

| 层级 | 技术 |
|------|------|
| 桌面框架 | [Wails v2](https://wails.io/) |
| 后端 | Go 1.21 |
| 前端 | 原生 JavaScript SPA |
| LLM 集成 | [Eino](https://github.com/cloudwego/eino)（字节跳动），兼容 OpenAI API |
| Python 环境 | [uv](https://docs.astral.sh/uv/) |
| 数据导出 | openpyxl（Excel） |

## 快速开始

### 环境要求

- Go 1.21+
- [Wails CLI v2](https://wails.io/docs/gettingstarted/installation)
- [uv](https://docs.astral.sh/uv/getting-started/installation/)（Python 包管理器）
- 可用的 OpenAI 兼容 LLM API

### 构建

```cmd
build.cmd
```

构建产物输出到 `build/bin/` 目录。

### 开发模式

```bash
wails dev
```

### 首次使用

1. 启动应用后，进入「设置」页面配置 LLM 连接（Base URL、API Key、模型名称）
2. 应用会自动初始化 Python 虚拟环境
3. 进入「样本分析」页面，粘贴日志样本，点击生成
4. 生成的代码验证通过后，进入「批量处理」选择输入/输出目录执行

## 工作流程

```
日志样本 → LLM 分析 → 生成 Python 代码 → 语法验证 → 批量处理 → Excel 输出
                                              ↓ 失败
                                         LLM 自动修复（最多3次）
```

## 项目结构

```
├── main.go                     # 程序入口
├── app.go                      # 主控制器，桥接前后端
├── internal/
│   ├── agent/                  # LLM 集成层
│   │   ├── llm_client.go       # LLM API 客户端
│   │   ├── sample_analyzer.go  # 样本分析与代码生成
│   │   └── code_validator.go   # 代码语法验证
│   ├── executor/
│   │   └── batch_executor.go   # 批量处理引擎
│   ├── project/
│   │   └── project_manager.go  # 项目持久化（JSON 文件存储）
│   ├── config/
│   │   └── settings_manager.go # 全局设置管理
│   ├── pyenv/
│   │   └── env_manager.go      # Python 虚拟环境管理（uv）
│   └── model/
│       └── model.go            # 共享数据模型
├── frontend/
│   ├── index.html
│   └── src/
│       ├── app.js              # SPA 路由与导航
│       ├── main.js             # 入口
│       ├── style.css
│       └── pages/
│           ├── sample.js       # 样本分析页
│           ├── batch.js        # 批量处理页
│           ├── projects.js     # 项目管理页
│           └── settings.js     # 设置页
├── build.cmd                   # Windows 构建脚本
└── wails.json                  # Wails 配置
```

## 许可证

私有项目，保留所有权利。
