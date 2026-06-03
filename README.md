# beryllium-be-there

课堂签到管理系统 —— 一个全栈演示项目，用于课堂场景下的学生签到管理。

## 项目架构

```
beryllium-be-there/
├── beryllium-server/    # Go 语言后端服务
├── beryllium-manage/    # 管理端前端 (Next.js)
├── beryllium-signin/    # 签到端前端 (Next.js)
├── publish.sh           # 一键构建发布脚本
└── README.md
```

### 架构图

```
浏览器 ──→ Go Server (:8080) ──→ JSON 文件存储 (data/*.json)
                │
                ├── /                  → beryllium-manage 管理页面
                ├── /signin/{classId}  → beryllium-signin 签到页面 (GET) / 签到接口 (POST)
                └── /api/*             → REST API
```

## 技术栈

| 组件 | 技术 |
|------|------|
| 后端 | Go 1.26+, 标准库 `net/http` |
| 管理端 | Next.js 16, React 19, TypeScript, Tailwind CSS 4 |
| 签到端 | Next.js 16, React 19, TypeScript, Tailwind CSS 4 |
| 数据库 | JSON 文件（演示项目） |
| 密码安全 | bcrypt |
| 传输加密 | RSA-2048 OAEP + SHA-256 |

## 核心设计

### 安全机制

1. **管理员认证**：基于 Bearer Token 的会话管理，Token 由 `crypto/rand` 生成，2 小时过期
2. **密码存储**：所有密码使用 bcrypt 哈希后存储，禁止明文
3. **签到密码传输**：每个课堂创建时生成 RSA-2048 密钥对。签到页面获取公钥，使用 Web Crypto API (RSA-OAEP) 加密密码后提交，服务端使用私钥解密后验证

### 签到流程

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│ 学生浏览器    │     │   Go Server   │     │  JSON 存储   │
└──────┬──────┘     └──────┬───────┘     └──────┬──────┘
       │  GET /signin/{id}  │                    │
       │ ──────────────────>│                    │
       │                    │ 查询课堂 + 公钥     │
       │                    │ ──────────────────>│
       │  HTML + 注入公钥    │                    │
       │ <──────────────────│                    │
       │                    │                    │
       │  用户输入账号密码    │                    │
       │  RSA-OAEP 加密密码  │                    │
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
Admin       { Username, PasswordHash }
Student     { Account (PK), Name, PasswordHash }
Class       { ClassID (UUID), Name, PublicKey, PrivateKey, CreatedAt }
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
| GET | `/api/classes/{id}` | Yes | 课堂详情（含签到状态） |
| DELETE | `/api/classes/{id}` | Yes | 删除课堂 |
| GET | `/api/classes/{id}/public-key` | No | 获取课堂 RSA 公钥 |
| GET | `/api/classes/{id}/attendance` | Yes | 签到记录列表 |
| POST | `/api/classes/{id}/attendance` | Yes | 初始化签到表 |
| PATCH | `/api/classes/{id}/attendance` | Yes | 手动修改签到状态 |
| GET | `/api/classes/{id}/export` | Yes | 导出 CSV |
| GET | `/signin/{classId}` | No | 签到页面（注入 classId + 公钥） |
| POST | `/signin/{classId}` | No | 提交签到 |

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
# 一键构建发布
./publish.sh

# 运行
cd publish/
./beryllium-be-there
# 访问 http://localhost:8080
```

## 项目结构

### beryllium-server（Go 后端）

```
beryllium-server/
├── main.go                     # 入口，路由注册
├── go.mod
├── go.sum
├── data/                       # 运行时 JSON 数据库（自动生成）
│   ├── admins.json
│   ├── students.json
│   ├── classes.json
│   └── attendance/
│       └── {classId}.json
└── internal/
    ├── model/models.go         # 数据结构定义
    ├── store/
    │   ├── json.go             # JSON 读写工具
    │   ├── admin_store.go      # 管理员持久化
    │   ├── student_store.go    # 学生持久化
    │   ├── class_store.go      # 课堂持久化（含 RSA 密钥）
    │   └── attendance_store.go # 签到记录持久化
    ├── crypto/
    │   ├── password.go         # bcrypt 密码哈希
    │   ├── rsa.go              # RSA 密钥生成与加解密
    │   └── uuid.go             # UUID v4 生成
    ├── auth/session.go         # Token 会话管理
    └── handler/
        ├── helpers.go          # 通用工具函数
        ├── auth.go             # 登录/登出接口
        ├── student.go          # 学生 CRUD 接口
        ├── class.go            # 课堂管理接口
        ├── attendance.go       # 签到状态接口
        ├── export.go           # CSV 导出接口
        ├── signin.go           # 签到页面 + 签到提交
        └── static.go           # 前端静态文件服务
```

### beryllium-manage（管理端前端）

```
beryllium-manage/
├── app/
│   ├── layout.tsx              # 根布局
│   ├── page.tsx                # 登录页
│   ├── globals.css             # Tailwind 样式
│   ├── student-manage/
│   │   └── page.tsx            # 学生管理
│   ├── class-manage/
│   │   └── page.tsx            # 课堂管理 + 签到详情
│   └── lib/
│       ├── api.ts              # API 请求封装
│       ├── auth.ts             # Token 管理
│       └── types.ts            # TypeScript 类型定义
├── next.config.ts
├── package.json
└── tsconfig.json
```

### beryllium-signin（签到端前端）

```
beryllium-signin/
├── app/
│   ├── layout.tsx              # 根布局
│   ├── page.tsx                # 签到表单
│   ├── globals.css             # Tailwind 样式
│   └── lib/
│       ├── crypto.ts           # RSA-OAEP 加密（Web Crypto API）
│       └── globals.d.ts        # 全局类型声明
├── next.config.ts
├── package.json
└── tsconfig.json
```

## 代码规范

- **类型安全**：Go 禁止 `interface{}`，TypeScript 禁止 `any`
- **导入顺序**：标准库 → 第三方库 → 自定义库
- **注释**：每个导出函数/方法开头简要描述功能
- **安全**：密码使用 bcrypt，禁止 DES/MD5

## License

MIT
