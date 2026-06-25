# 项目架构总结文档

> 本文档由 Trae AI 自动生成，用于快速了解项目全貌。

## 1. 项目简述
- **项目名称**: Moments
- **项目类型**: 全栈应用（Go 后端 + Nuxt Web 前端）
- **简述**: 一个极简朋友圈 / Memo 发布平台，支持多用户、Markdown 内容、评论点赞、文件上传、RSS 与后台配置管理。

## 2. 技术栈
- **核心语言**: Go、TypeScript、Vue
- **主要框架**: Echo、GORM、Nuxt 3、Vue 3
- **关键依赖**:
  - `github.com/labstack/echo/v4`：后端 HTTP 服务与路由
  - `gorm.io/gorm` + `github.com/glebarez/sqlite`：ORM 与 SQLite 存储
  - `github.com/golang-jwt/jwt/v5`：登录态 JWT 鉴权
  - `nuxt`：前端应用框架与构建体系
  - `@nuxt/ui`：前端界面组件库
- **包管理工具**: Go Modules、pnpm

## 3. 目录结构说明
使用树状图展示核心目录结构，并为关键目录添加注释说明。

```text
项目根目录/
├── backend/ # Go 后端服务
│   ├── db/ # 数据模型、数据库初始化与迁移对象
│   ├── docs/ # Swagger 生成产物
│   ├── handler/ # API 处理器，按业务拆分为用户、Memo、评论、文件等模块
│   ├── log/ # 日志初始化
│   ├── middleware/ # 鉴权中间件
│   ├── pkg/ # 邮件、通用能力等内部包
│   ├── util/ # 文件系统等工具函数
│   ├── vo/ # 请求/响应与配置结构体
│   ├── main.go # 后端启动入口
│   ├── router.go # 路由注册
│   └── go.mod # Go 依赖定义
├── front/ # Nuxt 3 前端应用
│   ├── assets/ # 全局样式资源
│   ├── components/ # 复用组件，如 Memo、评论、上传、预览等
│   ├── layouts/ # 页面布局
│   ├── pages/ # 路由页面，包含首页、发布、详情、设置、登录注册等
│   ├── public/ # 静态资源与前端脚本
│   ├── types/ # TypeScript 类型定义
│   ├── utils/ # API 请求、上传、Markdown 渲染等工具
│   ├── app.vue # 前端根组件
│   ├── store.ts # 本地持久化全局状态
│   ├── nuxt.config.ts # Nuxt / Vite 配置
│   └── package.json # 前端依赖与脚本
├── .github/ # CI/CD 工作流与发布配置
├── .devcontainer/ # 开发容器配置
├── Dockerfile # 镜像构建配置
├── docker-compose.yml # 容器编排示例
├── Makefile # 常用开发命令封装
├── README.md # 项目使用说明
└── docs.md # 项目架构总结文档
```

## 4. 核心模块解析
- **入口文件**: `backend/main.go` - 读取环境变量配置，初始化依赖注入、日志、数据库、鉴权中间件与静态资源服务，执行数据库迁移后启动 Echo HTTP 服务。
- **前端入口**: `front/app.vue` - 挂载 NuxtLayout、NuxtPage 与全局消息提示，并引入 Markdown / Fancybox 相关样式。
- **路由注册模块**: `backend/router.go` - 统一注册 `/api` 下的用户、Memo、评论、系统配置、标签、文件、友情链接与 RSS 路由。
- **鉴权模块**: `backend/middleware/auth.go` - 从 `x-api-token` 解析 JWT，为受保护接口注入当前用户，同时放行登录、注册、公开列表和部分公开资源。
- **Memo 核心模块**: `backend/handler/memo.go` - 负责 Memo 列表查询、详情、点赞、标签过滤、豆瓣信息抓取及图片缩略图 URL 处理，是核心业务逻辑所在。
- **文件上传模块**: `backend/handler/file.go` - 支持本地上传、基于 SHA256 的去重、图片缩略图生成以及无引用文件清理；同时项目也支持 S3 预签名上传。
- **数据库模块**: `backend/db/db.go` - 使用 GORM 连接 SQLite，初始化日志并自动迁移 User、Comment、Memo、SysConfig、Friend 等表结构。
- **前端请求与渲染模块**: `front/utils/index.ts` - 封装统一的 `useMyFetch` 请求函数、登录态 Token 注入、上传逻辑与 Markdown 渲染能力。
- **首页模块**: `front/pages/index.vue` - 负责加载 Memo 列表、分页加载、监听刷新事件并驱动首页时间线展示。

## 5. 启动与构建
- **安装依赖**: `cd front && pnpm install`
- **本地运行**:
  - 后端：`cd backend && go build -ldflags="-X main.version=local -X main.commitId=local" -o ./dist/moments && ./dist/moments`
  - 前端：`cd front && pnpm run dev`
  - 或使用 Makefile：`make backend-dev` / `make frontend-install` / `make frontend-dev`
- **构建打包**:
  - 后端：`cd backend && go build -ldflags="-X main.version=local -X main.commitId=local" -o ./dist/moments`
  - 前端：`cd front && pnpm run build`
