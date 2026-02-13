# LogForge 技术文档

## 1. 系统概述

LogForge（网络日志格式化系统）是一款基于 Wails v2 的 Windows 桌面应用。核心能力是利用大语言模型（LLM）自动分析网络日志样本，生成 Python 解析代码，然后在隔离的 Python 环境中批量执行，将结果导出为 Excel 文件。

### 1.1 架构总览

```
┌─────────────────────────────────────────────┐
│                  Frontend                    │
│         原生 JavaScript SPA (Wails)          │
│  ┌──────┐ ┌──────┐ ┌──────┐ ┌──────────┐   │
│  │样本  │ │批量  │ │项目  │ │  设置    │   │
│  │分析页│ │处理页│ │管理页│ │  页面    │   │
│  └──┬───┘ └──┬───┘ └──┬───┘ └────┬─────┘   │
│     │        │        │          │          │
├─────┼────────┼────────┼──────────┼──────────┤
│     ▼        ▼        ▼          ▼          │
│              app.go (主控制器)                │
│     Wails Binding — Go ↔ JavaScript         │
├─────────────────────────────────────────────┤
│                 Go Backend                   │
│  ┌────────┐ ┌──────────┐ ┌──────────────┐  │
│  │ Agent  │ │ Executor │ │   Project    │  │
│  │(LLM)  │ │(批量处理) │ │  (持久化)    │  │
│  └────────┘ └──────────┘ └──────────────┘  │
│  ┌────────┐ ┌──────────┐ ┌──────────────┐  │
│  │ Config │ │  PyEnv   │ │    Model     │  │
│  │(设置)  │ │(Python)  │ │  (数据模型)   │  │
│  └────────┘ └──────────┘ └──────────────┘  │
├─────────────────────────────────────────────┤
│            Python Runtime (uv)              │
│         隔离虚拟环境 + openpyxl              │
└─────────────────────────────────────────────┘
```

## 2. 核心模块

### 2.1 app.go — 主控制器

应用的中枢，实现了所有暴露给前端的 Wails binding 方法。

**主要职责：**
- 应用生命周期管理（startup 初始化）
- LLM 组件懒加载初始化
- 协调各内部模块完成业务流程
- 目录选择对话框（调用系统原生对话框）

**关键 API：**

| 方法 | 说明 |
|------|------|
| `AnalyzeSample(name, text)` | 分析日志样本，生成并验证 Python 代码 |
| `RunBatch(projectID, inputDir, outputDir)` | 启动批量处理任务 |
| `GetBatchProgress()` | 获取当前批量处理进度 |
| `ListProjects()` / `GetProject(id)` | 项目列表与详情 |
| `UpdateProjectCode(id, code)` | 更新项目代码 |
| `DeleteProject(id)` | 删除项目 |
| `RerunProject(id, inputDir, outputDir)` | 重新执行项目 |
| `GetSettings()` / `SaveSettings(settings)` | 读写全局设置 |
| `TestLLM()` | 测试 LLM 连接 |
| `EnsurePythonEnv()` | 手动触发 Python 环境初始化 |
| `GetPythonEnvReady()` | 查询 Python 环境状态 |
| `SelectDirectory(title)` | 打开系统目录选择对话框 |

### 2.2 internal/agent — LLM 集成层

#### LLMClient (`llm_client.go`)

封装与 OpenAI 兼容 API 的通信，基于字节跳动 Eino 框架。

- 支持任意 OpenAI 兼容的 LLM 服务（如 OpenAI、DeepSeek、本地部署等）
- 提供 `Chat(ctx, messages)` 方法进行多轮对话
- 配置项：BaseURL、APIKey、ModelName

#### SampleAnalyzer (`sample_analyzer.go`)

负责将日志样本发送给 LLM，获取 Python 解析代码。

- 构造包含样本数据的 prompt，引导 LLM 生成完整的 Python 程序
- 生成的代码需要能够：
  - 读取指定目录下的日志文件
  - 解析日志内容为结构化数据
  - 通过 stdout 输出 JSON 格式的进度信息
  - 使用 openpyxl 将结果写入 Excel

#### CodeValidator (`code_validator.go`)

验证生成的 Python 代码语法正确性。

- 使用 Python 的 `py_compile` 模块进行语法检查
- 验证失败时，将错误信息反馈给 LLM 进行自动修复
- 最多重试 3 次

### 2.3 internal/executor — 批量处理引擎

#### BatchExecutor (`batch_executor.go`)

执行生成的 Python 脚本，处理整个目录的日志文件。

**执行流程：**
1. 将项目代码写入临时 Python 脚本文件
2. 通过 `PythonEnvManager.RunScript()` 在隔离环境中执行
3. 实时解析 stdout 中的 JSON 进度信息，更新 `BatchProgress`
4. 监控 stderr 捕获运行时错误
5. 执行失败时，调用 `CodeRepairer` 接口修复代码并重试

**自动修复机制：**
- `CodeRepairer` 接口由 `app.go` 中的 `llmRepairerAdapter` 实现
- 将运行时错误信息和原始代码发送给 LLM
- LLM 返回修复后的代码，重新执行

### 2.4 internal/project — 项目持久化

#### ProjectManager (`project_manager.go`)

基于 JSON 文件的项目存储管理。

- 每个项目存储为独立的 JSON 文件（`{id}.json`）
- 存储路径：`{configDir}/projects/`
- 项目 ID 使用 UUID，文件名经过安全过滤防止路径穿越
- 支持 CRUD 操作和部分更新

**项目状态流转：**
```
draft → validated → executed
  ↓        ↓          ↓
  └────────┴──→ failed
```

### 2.5 internal/config — 设置管理

#### SettingsManager (`settings_manager.go`)

管理全局应用设置的读写。

- 存储路径：`{configDir}/settings.json`
- 配置项包括：
  - LLM 连接配置（BaseURL、APIKey、ModelName）
  - uv 路径
  - 默认输入/输出目录
  - 是否显示启动向导

### 2.6 internal/pyenv — Python 环境管理

#### PythonEnvManager (`env_manager.go`)

通过 uv 管理隔离的 Python 虚拟环境。

- `EnsureEnv()`：创建虚拟环境并安装 openpyxl 依赖
- `RunScript()`：在虚拟环境中执行 Python 脚本，返回 stdout/stderr 管道
- `GetStatus()`：查询环境状态（ready/pending/error）
- `checkUv()`：验证 uv 工具是否可用

### 2.7 internal/model — 数据模型

定义所有跨模块共享的数据结构：

| 类型 | 说明 |
|------|------|
| `LLMConfig` | LLM API 连接配置 |
| `Settings` | 全局应用设置 |
| `Project` | 项目记录（含代码、状态、时间戳） |
| `ProjectUpdate` | 项目部分更新 |
| `GenerateResult` | 代码生成结果 |
| `BatchResult` | 批量处理结果摘要 |
| `BatchProgress` | 批量处理实时进度 |
| `ProgressInfo` | Python 脚本输出的进度 JSON |

## 3. 前端架构

### 3.1 SPA 路由

前端是纯原生 JavaScript 实现的单页应用，通过 hash 路由切换页面。

- `app.js`：路由核心，管理页面注册、导航、LLM 配置状态检查
- 未配置 LLM 时，强制跳转到设置页面，其他导航项禁用

### 3.2 页面模块

| 页面 | 文件 | 功能 |
|------|------|------|
| 样本分析 | `sample.js` | 输入日志样本，调用 AI 生成解析代码 |
| 批量处理 | `batch.js` | 选择项目和目录，执行批量处理，显示实时进度 |
| 项目管理 | `projects.js` | 项目列表、代码编辑、删除、重新执行 |
| 设置 | `settings.js` | LLM 配置、Python 环境状态、默认目录设置 |

### 3.3 Go-JS 绑定

Wails 自动生成 `frontend/wailsjs/go/main/App.js`，前端通过 `window.go.main.App.MethodName()` 调用 Go 后端方法，返回 Promise。

## 4. 数据流

### 4.1 样本分析流程

```
用户粘贴日志样本
    ↓
App.AnalyzeSample(name, text)
    ↓
SampleAnalyzer.Analyze() → LLM API → 返回 Python 代码
    ↓
CodeValidator.Validate() → Python py_compile
    ↓ 失败？→ LLM 自动修复 → 重新验证（最多3次）
    ↓
ProjectManager.Create() → 保存项目
    ↓
返回 GenerateResult 给前端
```

### 4.2 批量处理流程

```
用户选择项目 + 输入/输出目录
    ↓
App.RunBatch(projectID, inputDir, outputDir)
    ↓
BatchExecutor.Execute()
    ├─ 写入临时 Python 脚本
    ├─ PythonEnvManager.RunScript() 执行
    ├─ 实时解析 stdout JSON 进度
    ├─ 前端轮询 GetBatchProgress()
    ↓ 失败？→ CodeRepairer 修复 → 重新执行
    ↓
输出 Excel 文件到指定目录
```

## 5. 构建与部署

### 5.1 构建流程

`build.cmd` 执行以下步骤：
1. 检查 Go 环境
2. 检查并自动安装 Wails CLI
3. `go mod tidy` 整理依赖
4. `wails build -clean -ldflags "-H windowsgui"` 编译

产物：`build/bin/network-log-formatter.exe`

### 5.2 运行时依赖

- uv（Python 包管理器）：用于创建虚拟环境
- 可用的 OpenAI 兼容 LLM API 端点

### 5.3 配置文件位置

应用配置存储在用户本地目录：
- `{configDir}/settings.json` — 全局设置
- `{configDir}/projects/*.json` — 项目数据

## 6. 安全考虑

- 项目 ID 经过安全过滤，防止路径穿越攻击
- 输入/输出目录强制使用绝对路径
- Python 代码在隔离虚拟环境中执行
- API Key 存储在本地配置文件中，用户需自行保护
- 前端对用户输入进行 HTML 转义，防止 XSS
