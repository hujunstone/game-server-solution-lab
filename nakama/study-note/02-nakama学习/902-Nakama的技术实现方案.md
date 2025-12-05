### 1. 总体架构思路

**Nakama 的总体技术实现方案可以概括为：**

- **语言与运行时**：使用 **Go** 编写的单体服务，内部划分为 API 层、核心业务层、脚本运行时层。
- **通信协议**：对外暴露 **gRPC + HTTP/JSON（通过 grpc‑gateway）**，同时支持 **WebSocket** 实时通信。
- **数据存储**：主要依赖 **PostgreSQL**，使用手写 SQL 和事务来完成全部数据读写。
- **脚本扩展**：通过内嵌 **Lua / JavaScript / Go 模块** 提供可插拔的游戏逻辑。
- **管理后台**：内置 **Console 管理后台**，前端打包为静态文件，后端通过 gRPC/HTTP 提供管理接口。
- **部署方式**：提供 Docker 镜像和 `docker-compose.yml`，方便本地开发和生产部署。

整体上是一个高性能 Go 后端 + Postgres + gRPC/HTTP + 可扩展脚本运行时的“实时游戏服务器”。

---

### 2. 主要技术/框架 & 在项目中的使用方式

#### 2.1 Go + 标准库

- **用途**
    - 实现绝大部分逻辑：连接管理、API 路由、业务逻辑、协程并发、定时任务等。
- **典型位置**
    - `main.go`：启动服务，加载配置，初始化数据库、日志、运行时，最后启动 gRPC/HTTP 监听。
    - `server/*.go`：绝大部分业务逻辑和 API 实现。

#### 2.2 gRPC + Protobuf + grpc‑gateway（API 层）

- **Proto 定义**
    - 公共 API 协议在 `vendor/github.com/heroiclabs/nakama-common/api/api.proto` 以及项目内 `apigrpc/apigrpc.proto`、`console/console.proto`。
    - 通过 `protoc` 生成：
        - Go 服务端 Stub（`apigrpc.pb.go`、`console.pb.go`）
        - gRPC‑Gateway 反向代理（`apigrpc.pb.gw.go`），将 HTTP/JSON 请求转成 gRPC 调用。
- **服务实现**
    - `server/api_*.go` 系列文件实现 gRPC 服务接口（`ApiServer` 方法），每个方法对应一个对外 API。
    - 例如：
        - `server/api_user.go::GetUsers` 实现 “获取用户信息” 的 RPC 方法，入参与出参类型来自 `nakama-common/api` 包。
        - `server/api_account.go::GetAccount/UpdateAccount/DeleteAccount` 实现账号相关接口。
        - `server/api_authenticate.go` 实现各类登录鉴权接口（Email、Username、Device、Facebook、Google、Steam 等）。
- **使用模式**
    - gRPC 层只做：
        - 参数校验；
        - 取当前会话上下文（userID、username 等）；
        - 调用核心逻辑函数（通常在 `core_*.go` 或 `core_user.go` / `core_account.go` / `core_authenticate.go`）。
        - 调 “Before/After Hook”（为脚本扩展预留）。

#### 2.3 PostgreSQL + `database/sql` + `pgx`（数据访问层）

- **Schema 管理**
    - 使用 SQL 迁移脚本位于 `migrate/sql/*.sql`，如：
        - `20180103142001_initial_schema.sql` 定义 `users`、`user_device`、`user_edge`、`storage`、`notification`、`wallet_ledger`、`groups` 等核心表。
- **数据库访问方式**
    - 全局使用 `*sql.DB` 连接池（标准库 `database/sql`）。
    - 查询/更新示例模式：
        - `db.QueryContext(ctx, query, params...)`
        - `db.ExecContext(ctx, query, params...)`
    - 对一些复杂事务，封装了：
        - `ExecuteInTx(ctx, db, func(tx *sql.Tx) error { ... })`
        - `ExecuteInTxPgx(ctx, db, func(tx pgx.Tx) error { ... })`
- **典型例子**
    - 用户查询：
        - `server/core_user.go::GetUsers`：根据 id / username / facebook_id 从 `users` 表批量查询用户，并映射为 `api.User`。
    - 账号详情：
        - `server/core_account.go::GetAccount`：`SELECT ... FROM users u ... array(select ud.id from user_device ud where u.id = ud.user_id)` 拼出账号 + 设备列表。
    - 登录注册：
        - `server/core_authenticate.go::AuthenticateEmail`：
            - `SELECT id, username, password, disable_time FROM users WHERE email = $1`；
            - 使用 bcrypt 校验密码；
            - 如果不存在则 `INSERT INTO users (id, username, email, password, ...)`.
    - 好友关系：
        - `server/core_friend.go` + 部分函数在 `core_authenticate.go`：
            - 读写 `user_edge` 表维护好友/拉黑/邀请状态；
            - 同时 `UPDATE users SET edge_count = edge_count ± 1` 维护关系计数。

#### 2.4 日志：`go.uber.org/zap`

- **用途**
    - 统一结构化日志记录，记录请求 traceID、错误堆栈、关键业务信息。
- **典型使用**
    - 几乎所有 API/核心函数签名都接受 `*zap.Logger`：
        - `LoggerWithTraceId(ctx, s.logger)` 给当前请求打上 traceID。
        - 调用 `logger.Error/Info/Warn` 记录结构化字段（`zap.String`、`zap.Error` 等）。

#### 2.5 脚本运行时：Lua / JavaScript / Go Runtime

- **Lua runtime**
    - 实现基于 `gopher-lua`（在 `internal/gopher-lua`）的 Lua 虚拟机。
    - `server/runtime_lua*.go` 中封装了 Lua 环境、注册 Nakama 内建函数（如存储、排行榜、RPC 调用等）。
- **JavaScript runtime**
    - `server/runtime_javascript*.go`：利用 Go 的 JS 引擎（例如 goja，具体依赖在 `vendor` 中）执行 JS 模块。
- **Go Runtime**
    - `server/runtime_go*.go`：允许通过 Go 插件形式扩展逻辑。
- **整体模式**
    - 运行时通过一套统一接口（`BeforeXxx` / `AfterXxx` Hook + 自定义 RPC 等）与 `ApiServer` 集成。
    - 例如：
        - 在 `server/api_user.go::GetUsers` 中，如果注册了 `BeforeGetUsers` 脚本，会先进入脚本逻辑，可以修改请求或直接拒绝访问。

#### 2.6 Console 管理后台（gRPC + Web 前端）

- **后端**
    - `console/console.proto` + `console.pb.go` 定义 Console 专用 gRPC 接口。
    - `server/console_*.go` 实现：
        - 控制台登录（多因子认证、权限控制）。
        - 用户查询、封号、导出账号、修改配置等管理操作。
- **前端**
    - 打包后的前端资源存放在 `console/ui/dist`（`index.html` + `static`）。
    - 通过 `console/ui.go` 将静态文件嵌入 Go 程序并对外提供 HTTP 路由。
- **使用方式**
    - Console 通过 gRPC/HTTP 调用后台接口（与游戏客户端不同的一套接口），权限更高，面向运营/运维人员。

#### 2.7 其他基础组件

- **定时调度 / Cron 表达式**
    - `internal/cronexpr`：用于排行榜重置、定时任务等。
- **社交平台集成**
    - `social/social.go` + `server/core_authenticate.go`：封装 Facebook/Google/Apple/Steam 等 SDK 访问。
- **配置管理**
    - `server/config.go`：集中读取配置（端口、DB、社交、会话、控制台等），通常通过 YAML/环境变量注入。

---

### 3. “如何使用这些技术”的几个典型流程

#### 3.1 用户登录（以 Email 为例）

- **gRPC/HTTP → API 层**
    - 客户端调用 `AuthenticateEmail`：
        - gRPC 服务在 `server/api_authenticate.go::AuthenticateEmail`。
        - 校验参数 → 调用 `core_authenticate.AuthenticateEmail`。
- **核心逻辑 → DB**
    - `server/core_authenticate.go::AuthenticateEmail`：
        - 先查 `users` 表；
        - 若存在，校验 `disable_time`、密码哈希；
        - 若不存在且允许创建，`INSERT` 新用户。
- **生成 Session & 缓存**
    - 核心函数返回 `userID/username` 后，API 层生成 JWT Token/Refresh Token（`generateToken` 等），写入 in‑memory `sessionCache`。
- **可选脚本扩展**
    - `BeforeAuthenticateEmail` / `AfterAuthenticateEmail` 可以在脚本中实现风控、埋点等逻辑。

#### 3.2 获取用户资料 / 批量查用户

- **单个账号详情**
    - `GetAccount`（客户端接口）：
        - `server/api_account.go::GetAccount` → `server/core_account.go::GetAccount` → `SELECT ... FROM users` + `user_device`。
- **批量用户信息**
    - `GetUsers`（传 ids/usernames/facebook_ids）：
        - `server/api_user.go::GetUsers` → `server/core_user.go::GetUsers`；
        - 使用 `id = ANY($1)` / `username = ANY(...)` 等方式批量查表。

#### 3.3 好友系统

- **列好友 / 加好友 / 拉黑**
    - API 层：`server/api_friend.go` 提供 `ListFriends` / `AddFriends` / `BlockFriends` 等。
    - 核心逻辑：`server/core_friend.go` + `core_authenticate.go` 中的导入好友函数，围绕 `user_edge` 表：
        - 维护 `state`（好友/邀请/拉黑）；
        - 维护 `users.edge_count`；
        - 结合通知/消息系统（`notification` 表、实时推送）。

---

### 4. 总结

- **技术栈**：Go + gRPC/Protobuf + grpc‑gateway + PostgreSQL + Lua/JS 内嵌脚本 + 内置 Console 前端。
- **结构方式**：`api_*.go` 实现对外 gRPC/HTTP 接口，`core_*.go` 负责业务 + SQL，`migrate/sql` 定义 Schema，`runtime_*.go` 提供脚本扩展，`console/*` 提供运营管理。
- **典型例子**：用户认证（`api_authenticate.go` + `core_authenticate.go`）、账号管理（`api_account.go` + `core_account.go`）、用户/好友系统（`api_user.go`、`api_friend.go`、`core_user.go`、`core_friend.go`）。

如果你想针对某一块（例如“脚本运行时如何嵌入 Lua/JS”或“Console 权限模型”）深入，我可以再按源码帮你拆得更细。