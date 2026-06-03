# beryllium-be-there

课堂签到管理系统 —— 一个全栈演示项目，用于课堂场景下的学生签到管理。

## 项目架构

```
beryllium-be-there/
├── beryllium-server/       # Go 语言后端服务
├── beryllium-manage/       # 管理端前端 (Next.js)
├── beryllium-signin/       # 签到端前端 (Next.js)
├── publish.sh              # 一键多平台构建发布脚本
└── README.md
```

### 架构图

```
浏览器 ──→ Go Server (:8080) ──→ JSON 文件存储 (data/*.json)
                │
                ├── /                       → beryllium-manage 管理页面
                ├── /signin/{classId}       → beryllium-signin 签到页面 (GET) / 签到接口 (POST)
                ├── /api/*                  → REST API
                └── /ws/classes/{id}/...    → WebSocket 实时推送
```

## 技术栈

| 组件 | 技术 |
|------|------|
| 后端 | Go 1.26+, 标准库 `net/http` |
| 管理端 | Next.js 16, React 19, TypeScript, Tailwind CSS 4 |
| 签到端 | Next.js 16, React 19, TypeScript, Tailwind CSS 4 |
| 数据库 | JSON 文件（演示项目） |
| 密码存储 | bcrypt |
| 传输加密 | RSA-2048 PKCS#1 v1.5（jsencrypt 纯 JS 实现） |
| 实时推送 | WebSocket（gorilla/websocket），签到状态即时同步 |
| 静态资源 | Go `embed.FS` 编译时嵌入 |
| 配置管理 | TOML 配置文件 + 环境变量覆盖 |

## 核心设计

### 安全机制

1. **管理员认证**：基于 Bearer Token 的会话管理，Token 由 `crypto/rand` 生成，2 小时过期
2. **密码存储**：所有密码使用 bcrypt 哈希后存储
3. **签到密码传输**：每个课堂创建时生成 RSA-2048 密钥对（PEM 格式）。签到页面获取公钥后使用 [jsencrypt](https://github.com/travist/jsencrypt)（纯 JavaScript 实现，无需 HTTPS）加密密码提交，服务端使用私钥解密后 bcrypt 验证

> **为什么用 jsencrypt 而不是 Web Crypto API？**
> 浏览器 `crypto.subtle` 仅在安全上下文（HTTPS 或 localhost）中可用。课堂场景中，学生通过 `http://192.168.x.x:8080` 访问签到页面属于非安全上下文，`crypto.subtle` 不可用。jsencrypt 是纯 JS 的 RSA 实现，不受此限制。

### 配置系统

配置优先级（由低到高）：

1. **默认值** — 端口 `8080`，主机名自动检测
2. **config.toml 文件** — `port` 和 `hostname` 配置项
3. **环境变量** — `PORT`、`HOSTNAME` 可覆盖文件配置

```toml
# config.toml
port = 8080
hostname = ""
```

`hostname` 留空时，服务器自动遍历本机网卡获取局域网 IPv4 地址，用于生成签到链接和二维码。

### 实时推送机制

当学生在签到页面完成签到时，服务端通过 WebSocket 实时将签到状态变更推送给所有正在查看该课堂管理页面的管理员：

```
学生签到 → POST /signin/{id}
              │
              ▼
       SigninHandler.Submit()
              │
              ├── 更新 attendance JSON
              └── Hub.Broadcast(classID, "attendance_update")
                        │
                        ▼
              ┌─────────────────┐
              │   WebSocket Hub  │  (per-class rooms)
              └───┬─────────┬───┘
                  │         │
          conn1 (PC)    conn2 (PC)
                  │         │
                  ▼         ▼
         管理端页面即时更新签到状态
```

- 管理端连接 `ws://{host}:{port}/ws/classes/{id}/attendance`
- 断线自动重连（3 秒间隔）
- 可随时关闭自动刷新，切换为手动刷新模式

### 签到流程

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│ 学生浏览器    │     │   Go Server   │     │  JSON 存储   │
└──────┬──────┘     └──────┬───────┘     └──────┬──────┘
       │  GET /signin/{id}  │                    │
       │ ──────────────────>│                    │
       │                    │ 查询课堂 + PEM公钥  │
       │                    │ ──────────────────>│
       │  HTML + 注入公钥    │                    │
       │ <──────────────────│                    │
       │                    │                    │
       │  用户输入账号密码    │                    │
       │  jsencrypt 加密密码 │                    │
       │  (PKCS#1 v1.5)     │                    │
       │                    │                    │
       │  POST /signin/{id} │                    │
       │ ──────────────────>│                    │
       │                    │ RSA 私钥解密       │
       │                    │ bcrypt 验证密码    │
       │                    │ ──────────────────>│
       │                    │ 标记签到状态       │
       │                    │ <──────────────────│
       │  签到成功/失败      │                    │
       │ <──────────────────│                    │
```

### 数据模型

```
Admin        { Username, PasswordHash }
Student      { Account (PK), Name, PasswordHash }
Class        { ClassID (UUID), Name, PublicKey, PrivateKey, CreatedAt }
SignInRecord { StudentAccount, IsPresent, SignedAt }
```

### 路由设计

| Method | Path | Auth | 说明 |
|--------|------|------|------|
| GET | `/` | No | 管理端 SPA |
| POST | `/api/login` | No | 管理员登录 |
| POST | `/api/logout` | No | 管理员登出 |
| GET | `/api/students` | Yes | 学生列表 |
| POST | `/api/students` | Yes | 添加学生 |
| DELETE | `/api/students/{account}` | Yes | 删除学生 |
| GET | `/api/classes` | Yes | 课堂列表 |
| POST | `/api/classes` | Yes | 创建课堂（生成 UUID + RSA 密钥对） |
| GET | `/api/classes/{id}` | Yes | 课堂详情（含签到状态和签到链接） |
| DELETE | `/api/classes/{id}` | Yes | 删除课堂 |
| GET | `/api/classes/{id}/public-key` | No | 获取课堂 RSA 公钥（PEM 格式） |
| GET | `/api/classes/{id}/attendance` | Yes | 签到记录列表 |
| POST | `/api/classes/{id}/attendance` | Yes | 初始化签到表 |
| PATCH | `/api/classes/{id}/attendance` | Yes | 手动修改签到状态 |
| GET | `/api/classes/{id}/export` | Yes | 导出 CSV |
| GET | `/signin/{classId}` | No | 签到页面（注入 classId + PEM 公钥） |
| POST | `/signin/{classId}` | No | 提交签到 |
| GET | `/ws/classes/{id}/attendance` | No | WebSocket 签到状态实时推送 |

## 快速开始

### 环境要求

- Go 1.26+
- Node.js 24+
- npm 11+

### 开发模式

```bash
# 1. 启动后端（终端 1）
cd beryllium-server
go run .

# 2. 启动管理端开发服务器（终端 2）
cd beryllium-manage
npm install
npm run dev
# 访问 http://localhost:3000

# 3. 启动签到端开发服务器（终端 3）
cd beryllium-signin
npm install
npm run dev
# 访问 http://localhost:3001
```

开发模式下，前端通过 CORS 跨域访问 `localhost:8080` 的后端 API。默认管理员账号：`admin` / `admin`。

### 生产构建

```bash
# 构建所有平台（darwin/linux/windows, amd64/arm64）
./publish.sh

# 仅构建指定平台
./publish.sh linux/amd64

# 清理构建产物
./publish.sh clean
```

发布目录结构：

```
publish/
├── darwin-amd64/          # macOS Intel
│   ├── beryllium-be-there
│   ├── config.toml
│   ├── manage/
│   └── signin/
├── darwin-arm64/          # macOS Apple Silicon
├── linux-amd64/           # Linux x86_64
├── linux-arm64/           # Linux ARM64
└── windows-amd64/         # Windows x64
    ├── beryllium-be-there.exe
    ├── config.toml
    ├── manage/
    └── signin/
```

运行方法：

```bash
cd publish/<platform>
./beryllium-be-there
# 访问 http://{host}:8080
```

## 项目结构

### beryllium-server（Go 后端）

```
beryllium-server/
├── main.go                     # 入口，路由注册
├── go.mod / go.sum
├── config.toml                 # 服务器配置文件
├── web/                        # 编译时嵌入的静态资源
│   ├── embed.go                # //go:embed 声明
│   └── signin_fallback.html    # 回退签到页面（当构建产物不可用时）
├── data/                       # 运行时 JSON 数据库（自动生成）
│   ├── admins.json
│   ├── students.json
│   ├── classes.json
│   └── attendance/
│       └── {classId}.json
└── internal/
    ├── config/config.go        # TOML 配置加载 + IP 自动检测
    ├── model/models.go         # 数据结构定义
    ├── store/
    │   ├── json.go             # JSON 读写工具
    │   ├── admin_store.go      # 管理员持久化
    │   ├── student_store.go    # 学生持久化
    │   ├── class_store.go      # 课堂持久化（含 RSA 密钥）
    │   └── attendance_store.go # 签到记录持久化
    ├── crypto/
    │   ├── password.go         # bcrypt 密码哈希
    │   ├── rsa.go              # RSA 密钥生成 + PEM 格式导入导出 + PKCS#1 v1.5 解密
    │   └── uuid.go             # UUID v4 生成
    ├── auth/session.go         # Token 会话管理
    ├── ws/hub.go               # WebSocket Hub（per-class 连接管理 + 广播）
    └── handler/
        ├── helpers.go          # 通用工具函数
        ├── auth.go             # 登录/登出接口
        ├── student.go          # 学生 CRUD 接口
        ├── class.go            # 课堂管理接口（含签到 URL 生成）
        ├── attendance.go       # 签到状态接口
        ├── export.go           # CSV 导出接口
        ├── signin.go           # 签到页面（PEM 公钥注入）+ 签到提交 + 广播
        ├── ws.go               # WebSocket 升级处理
        └── static.go           # 前端静态文件服务（多路径探测）
```

### beryllium-manage（管理端前端）

```
beryllium-manage/
├── app/
│   ├── layout.tsx              # 根布局
│   ├── page.tsx                # 登录页
│   ├── globals.css             # Tailwind 主题定义
│   ├── student-manage/
│   │   └── page.tsx            # 学生管理（添加/删除/列表）
│   ├── class-manage/
│   │   └── page.tsx            # 课堂管理 + 签到详情（QR 码 + CSV 导出）
│   └── lib/
│       ├── api.ts              # API 请求封装（自动检测生产/开发模式）
│       ├── auth.ts             # Token 本地存储管理
│       └── types.ts            # TypeScript 类型定义
├── next.config.ts              # output: 'export'
├── package.json
└── tsconfig.json
```

### beryllium-signin（签到端前端）

```
beryllium-signin/
├── app/
│   ├── layout.tsx              # 根布局
│   ├── page.tsx                # 签到表单（使用 window.location.origin）
│   ├── globals.css             # Tailwind 主题定义
│   └── lib/
│       ├── crypto.ts           # jsencrypt RSA 加密（纯 JS，无需 HTTPS）
│       └── globals.d.ts        # 全局类型声明（注入变量）
├── next.config.ts              # output: 'export', assetPrefix: '/signin'
├── package.json
└── tsconfig.json
```

## 代码规范

- **类型安全**：Go 使用 `any` 替代 `interface{}`，TypeScript 禁止 `any`
- **导入顺序**：标准库 → 第三方库 → 自定义库
- **注释**：每个导出函数/方法开头简要描述功能
- **安全**：密码使用 bcrypt，禁止 DES/MD5

## License

MIT
