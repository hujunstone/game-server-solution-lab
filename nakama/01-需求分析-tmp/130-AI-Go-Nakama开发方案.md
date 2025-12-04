## Nakama Go 服务端开发与部署方案

### 1. 总览
- **技术栈**：Go 1.21+、Nakama 3.x（Go Runtime 模块）、CockroachDB、可选 Redis。  
- **当前阶段**：以 **Demo / 原型验证** 为主，优先使用 **单机或少量服务器 + Docker / docker-compose / systemd**，保证部署简单、可快速迭代；Kubernetes 仅作为未来正式服/大规模并发时的扩展选项。  
- **项目目标**：支撑角色/战斗/跑商玩法，暴露 RPC、Realtime、Match 官方接口，并确保后续可平滑演进到滚动升级架构。  
- **现有仓库结构**：根目录 `nakama/` 下包含 `main.go`, `login/`, `02-需求分析-精华输出/`, `docker-compose.yml` 等文件，可直接扩展。

### 2. 工程目录规划

```
dmst-server/                     # 游戏服务端工程根目录（与 UE 客户端项目分离）
├── cmd/                         # 可执行程序或插件入口（main 包）
│   └── nakama-plugin/           # Nakama 插件入口，编译为共享库 (.so/.dll)
├── internal/                    # 应用核心代码（不对外暴露的 Go 模块实现）
│   ├── runtime/                 # Nakama Go Runtime 整合层（注册 RPC、Match、Hook）
│   │   ├── rpc/                 # Nakama 自定义 RPC 实现（角色、商队、战斗等），按领域拆分文件
│   │   ├── match/               # 战斗/战役的实时匹配与回合逻辑（Nakama Match Handler）
│   │   ├── events/              # 服务器内部事件总线（任务完成、战斗结算等回调触发）
│   │   └── hooks/               # 登录、登出等生命周期钩子处理（鉴权、初始化 Session 等）
│   │       └── login/           # 迁移自现有 `login/` 目录的登录 Hook 实现
│   ├── services/                # 领域服务层（角色、战斗、商队、贸易、战役等业务规则）
│   ├── repositories/            # 数据访问层，对接 CockroachDB（SQL、事务、查询封装）
│   ├── cache/                   # 本地缓存与 Redis 适配（配置缓存、Session 缓存等）
│   ├── config/                  # 配置模型与加载逻辑（YAML/ENV/Secrets，支持热更新）
│   └── observability/           # 日志、指标与链路追踪（logger、metrics、tracing）
├── api/                         # 对外接口契约定义
│   └── contracts/               # JSON schema / protobuf（生成 UE SDK 或用于校验）
├── migrations/                  # 数据库迁移脚本目录
│   └── *.sql                    # CockroachDB 建表与变更 SQL（按版本拆分）
├── deployments/                 # 部署相关文件
│   ├── docker/                  # Dockerfile、docker-compose 配置
│   └── k8s/                     # 可选的 K8s/Helm 配置（未来正式服扩展用）
├── scripts/                     # 辅助脚本
│   ├── build_plugin.ps1/sh      # CI / 本地构建 Nakama 插件的脚本
│   ├── test.sh                  # 统一执行单元 & 集成测试的脚本
│   └── hot_reload.sh            # 开发/测试环境的热重载与滚动发布辅助脚本
├── go.mod                       # Go module 定义（模块路径与依赖声明）
├── go.sum                       # Go 依赖版本锁定文件
├── README.md                    # 工程总览、快速启动说明
└── docs/                        # 设计文档、接口说明、存储方案等
```

> 迁移指南：  
> 1. 将现有 `login/login_hooks.go` 移至 `internal/runtime/hooks/login/hooks.go`，更新包名与引用路径。  
> 2. 原 `main.go` 中的注册逻辑迁至 `cmd/nakama-plugin/main.go`，并使用 `internal/runtime` 暴露的注册函数。  
> 3. 根目录保留 `docker-compose.yml`、`Dockerfile`，对应 `deployments/docker` 中的模板。

### 3. 关键模块规划

| 模块 | 说明 | 主要文件 |
| --- | --- | --- |
| `cmd/nakama-plugin/main.go` | 插件入口，注册所有 RPC、Match、Hook。 | `nk.RegisterRpc`, `nk.RegisterMatch` |
| `internal/runtime/rpc` | 对应接口文档的 `rpc_*`。每个文件一个子领域，封装参数校验、调用服务层。 | `rpc_character.go`, `rpc_battle.go` 等 |
| `internal/runtime/match` | 战斗/战役的实时逻辑，利用 Nakama Match Handler。 | `battle_match.go`, `campaign_match.go` |
| `internal/services` | 业务规则（战斗模拟、战术冷却、商路风险）。可无状态，注入 repo/cache。 | `character_service.go`, `route_service.go` |
| `internal/repositories` | Go database/sql + pgx，封装 CockroachDB CRUD，支持事务。 | `character_repo.go`, `battle_repo.go` |
| `internal/cache` | LRU/segment cache + 可选 Redis 客户端，提供统一接口。 | `config_cache.go`, `session_cache.go` |
| `internal/config` | 读取 YAML/ENV，提供结构化配置；支持热更新（SIGHUP / API）。 | `config.go` |
| `internal/observability` | Zap/Logrus 日志、Prometheus 指标、OpenTelemetry Trace。 | `logger.go`, `metrics.go` |

### 4. 开发流程
1. **数据模型**：依据 `03-数据结构设计.md` 同步定义 Go struct（`/internal/repositories/dto`），保持 `snake_case` 映射。  
2. **接口实现**：先在 `internal/services` 实现业务，再在 `internal/runtime/rpc` 暴露给 Nakama；新增 RPC 时同步更新 `api/contracts/*.json`。  
3. **缓存策略**：  
   - 静态配置：启动时加载到 `cache.ConfigStore`。  
   - Session 状态：存入 `cache.SessionStore`（map+RWMutex），Match/Travel/Battle 使用。  
4. **测试**：  
   - 单元测试：`go test ./...`。  
   - 集成测试：使用 `docker-compose` 启动 CockroachDB + Nakama + plugin，运行 Postman/Newman 或 UE 自动化脚本。  
5. **代码规范**：`golangci-lint`、`gofumpt`、`staticcheck`；提交前 `make lint test build`。  
6. **依赖管理**：使用 Go modules，固定 Nakama SDK 版本（如 `github.com/heroiclabs/nakama-common v1.34.0`），通过 `go.work` 或 `GONOPROXY` 指定私有模块代理。

### 5. 构建与部署

#### 5.1 构建（含从现有结构迁移）
- 迁移步骤：  
  1. 新建 `cmd/nakama-plugin/main.go`，引用 `internal/runtime` 中注册逻辑。  
  2. 将原 `main.go` 内容拆为 `cmd/...` 入口与 `internal/runtime/bootstrap.go`。  
  3. 更新 `go.mod`，确保所有包路径指向新目录。  
- 生成 Nakama 插件：
  ```bash
  // 生成名为 build/nakama_plugin.so 的共享库文件，用于 Nakama 加载
  // -buildmode=plugin  指定输出为 Go 插件（.so 文件，供宿主进程动态加载）
  // -o build/nakama_plugin.so  指定输出文件路径和文件名
  // ./cmd/nakama-plugin   指定插件主程序目录（含 main.go）
  go build -buildmode=plugin -o build/nakama_plugin.so ./cmd/nakama-plugin
  ```
- Docker 镜像（引用 repo 已有 `Dockerfile`，若需自定义，建议多阶段构建）：
  ```dockerfile
  FROM golang:1.21 as build
  WORKDIR /app
  COPY . .
  RUN go build -buildmode=plugin -o build/nakama_plugin.so ./cmd/nakama-plugin
  
  FROM heroiclabs/nakama:3.21
  COPY --from=build /app/build/nakama_plugin.so /nakama/data/modules/
  COPY config.yml /nakama/data/
  ```

#### 5.2 部署模式（Demo 阶段优先）
- **本地/测试/Demo 服**：`docker-compose up`（CockroachDB + Nakama + Console）。支持两种调试模式：  
  1. **插件模式**：执行 `scripts/build_plugin.sh`，compose 挂载 `nakama_plugin.so`。  
  2. **源码模式**：在 Nakama 配置中设置 `--runtime.path /workspace/build/dev`，使用 `go run cmd/nakama-plugin/main.go` 热重载。  
- **小规模线上 Demo**：1–2 台云主机，使用 Docker + `docker-compose` 或 `systemd` 管理 Nakama 进程：  
  - 通过 Nginx/SLB 将流量打到正在运行的实例；  
  - 发布时先在备用实例上拉起新镜像，验证健康后切换流量，再平滑停掉旧实例（简化版 blue/green）。  

#### 5.3 面向生产/未来的 K8s 部署（可选）
- 当并发量和可用性要求提升后，可引入 Kubernetes/Helm，至少 2 个 Nakama 实例 + 1 个 CockroachDB 集群（3 节点）。`deployments/k8s` 中提供：
  - Deployment（RollingUpdate），配置 `readinessProbe`（HTTP `/v2/health`）与 `livenessProbe`（自定义 `/healthz/runtime`），控制 `maxUnavailable=0`, `maxSurge=1`。  
  - Node Affinity：对战 Match Pod 通过 `spec.affinity` 绑定到同一可用区，减少跨区延迟。  
  - ConfigMap/Secret：数据库连接串、Redis、Feature Flag。  
  - HPA：根据 CPU/QPS 自动伸缩。

### 6. 无停机更新（软件包滚动发布）

1. **镜像分层版本**：插件和配置打包成 `heroiclabs/nakama:<base>-game:vX.Y.Z`。  
2. **滚动策略**（Demo 阶段以 compose/systemd 为主，K8s 为可选扩展）：  
   1. **预检**：在 staging 环境执行 `go test`、`docker-compose up` 验证。  
   2. **迁移**：运行 `migrations`（如 `goose up`）到目标版本。  
   3. **构建镜像**：打包新版本并推送镜像仓库。  
   4. **部署**：  
      - Docker/docker-compose/systemd：  
        - 在备用主机或备用容器上拉起新版本，执行 smoke 测试；  
        - 通过负载均衡（Nginx/SLB）切换到新实例；  
        - 平滑关闭旧实例，确保没有活跃战斗/请求。  
      - （可选）Kubernetes：`kubectl rollout restart deployment/nakama`，RollingUpdate 控制 `maxUnavailable=0`, `maxSurge=1`。  
      - （可选）裸机/VM Blue-Green：使用 `systemd` + 目录布局：  
        - `/srv/nakama/releases/vX.Y.Z`（新版本）  
        - `/srv/nakama/current` 符号链接。  
        - 新版本启动新进程（不同端口或 upstream 权重 0），健康后切换负载均衡。  
   5. **验证**：对新 Pod 执行 smoke 测试（调用关键 RPC）、确认 metrics 正常。  
   6. **清理**：旧版本观察 24h 无异常后清理。
3. **会话保持**：  
   - 通过 Nakama 内建的 **runtime RPC 幂等** + Session Reconnect；  
   - 对战/战术 Match：  
     - 使用 `--matchmaker.max_tickets` 控制分配，结合 `MatchLabelFilter`（在 `nk.MatchCreate` 时传 label）确保同一战斗固定在原节点；  
     - K8s 层设置 `podDisruptionBudget`，确保滚动过程中至少有一个旧 Pod 等待所有 Match 结束后再终止；  
     - `--session.token_expiry_ms` 设置比滚动时间更长，允许客户端重连。
4. **数据迁移**：  
   - 采用 `migrations/` + Goose/Atlas 执行 schema 变更；  
   - 先运行 `ALTER` 等非阻塞 SQL，确认成功后再滚动服务。  
5. **包热加载（可选）**：Nakama 支持 `--runtime.path` 下的模块热重载（通过 REST `POST /v2/runtimes/reload` 或 SIGHUP）。若需在线调试，可：  
   - 在 staging 或单节点测试环境启用；  
   - 使用脚本 `scripts/hot_reload.sh` 调用 API，观察 `nk.Logger` 输出确认加载成功；  
   - 生产环境仍推荐滚动替换，避免中途状态不一致。

### 7. 配置与密钥管理
- `config/config.yml`（示例）：
  ```yaml
  database:
    cockroach:
      url: postgresql://user:pwd@cockroach:26257/game?sslmode=disable
  cache:
    redis_url: redis://redis:6379/0
  runtime:
    max_concurrent_matches: 2000
  features:
    enable_campaign_battle: true
  ```
- 环境变量覆盖：`NAKAMA_DATABASE__COCKROACH__URL` 等。  
- Secrets：在 Demo / 非 K8s 环境下，可使用 **Docker secrets、受权限控制的 .env 文件、或云厂商密钥管理服务（如 AWS Secrets Manager / Azure KeyVault）** 注入敏感信息；若未来接入 K8s，可再切换为 K8s Secret / Vault 注入。

### 8. 可观测性与回滚
- **日志**：结构化 JSON（包含 `player_id`, `rpc`, `route_segment_code`），集中到 ELK/Loki。示例：
  ```go
  observability.Logger.Info("rpc_execute_trade",
      zap.String("player_id", playerID),
      zap.String("node", nk.Node()),
      zap.Duration("duration_ms", since))
  ```
- **指标**：暴露 Prometheus 端点 `/metrics`，关键指标：  
  - `nakama_rpc_latency_seconds{rpc="rpc_execute_trade"}` (Histogram，告警阈值 p95 > 200ms)。  
  - `nakama_match_active_total` (Gauge)。  
  - `nakama_cache_hit_ratio` (Gauge)。  
- **Trace**：OpenTelemetry 导出到 Jaeger，使用 `otelhttp` 包围 RPC Handler。  
- **回滚**：K8s `kubectl rollout undo` 或切换 `current` 符号链接；数据库迁移使用 down 脚本或备份。  
- **健康检查**：`/v2/health`（Nakama 内置）+ 自定义 `/healthz/runtime` 检查依赖（Cockroach、Redis），示例：
  ```go
  func RuntimeHealth(w http.ResponseWriter, r *http.Request) {
      if err := repositories.PingDB(); err != nil {
          w.WriteHeader(http.StatusServiceUnavailable)
          return
      }
      w.WriteHeader(http.StatusOK)
  }
  ```

### 9. 本地开发与协作
- **调试**：  
  - VS Code/Goland：使用 `launch.json` 或 Run Configuration，设置 `GOOS=linux`、`GOARCH=amd64`，运行 `go run cmd/nakama-plugin/main.go`，并在 `nk.RegisterRpc` 处打断点；Nakama 以 `--allow_exec` 模式启动以便外部进程注册。  
  - UE 客户端联调：通过 Nakama Dev Console 的 `RPC` 工具发送请求，或使用自动化蓝图脚本。  
- **文档**：`docs/` 中维护本方案、接口文档、存储方案，CI 自动生成 PDF/HTML。  
- **Git 流程**：采用 Git Flow 或 trunk-based，设 `develop`、`main` 分支；CI（GitHub Actions）执行 lint/test/build/push image，构建产物附带版本标签。  
- **版本兼容**：每次 Nakama 升级前先在 `feature/nakama-x.y` 分支验证 SDK/API 兼容性，更新 `go.mod` 后再合入主干。  
- **客户端契约**：UE 客户端依赖 `api/contracts` 生成的 JSON schema，保障 RPC 字段同步。

---
通过以上目录和部署规划，可确保 Nakama Go 工程结构清晰、易于扩展；利用 RollingUpdate/Blue-Green + 幂等 RPC，能够在不停机的情况下完成插件与依赖更新。

