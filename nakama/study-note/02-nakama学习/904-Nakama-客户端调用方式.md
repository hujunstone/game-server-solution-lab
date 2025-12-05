### 基本规则

- WebSocket 地址固定是：`ws://<host>:<port>/ws` 或 `wss://<host>:<port>/ws`。
- 关键查询参数：
    - **`token`**：登录/认证接口返回的会话 token（必填）。
    - **`format`**：`json` 或 `protobuf`（不填默认 `json`）。
    - **`lang`**：语言代码，如 `en` / `zh-CN`（可选）。
    - **`status`**：`true/false`，是否同时注册状态流（可选）。

---

### JSON 示例（蓝图 SDK 常用默认形式）

- **URL 示例**

```text
ws://127.0.0.1:7350/ws?token=<SESSION_TOKEN>&format=json&lang=zh-CN&status=true
```

或省略 `format`（默认就是 JSON）：

```text
ws://127.0.0.1:7350/ws?token=<SESSION_TOKEN>&lang=zh-CN&status=true
```

- **说明**
    - WebSocket 帧类型为 **TextMessage**，内容为 JSON，结构是 `rtapi.Envelope` 的 JSON 版。

---

### Protobuf 示例（高性能、二进制）

- **URL 示例**

```text
ws://127.0.0.1:7350/ws?token=<SESSION_TOKEN>&format=protobuf&lang=zh-CN&status=true
```

- **说明**
    - WebSocket 帧类型为 **BinaryMessage**，内容为 Protobuf 编码的 `rtapi.Envelope`。
    - 蓝图 SDK 需要在客户端侧也启用 Protobuf 序列化/反序列化，结构与 `rtapi.Envelope` 对应。

---

### 蓝图 SDK 里典型拼接方式（伪代码思路）

- 先通过 REST/gRPC 登录拿到 `session.Token`。
- 在蓝图中拼接字符串变量，例如：

```text
BaseUrl = "ws://127.0.0.1:7350/ws"
Query   = "?token=" + SessionToken + "&format=protobuf&lang=zh-CN&status=true"
Final   = BaseUrl + Query
```

然后把 `Final` 作为 WebSocket 连接 URL 传给蓝图的 WebSocket 组件即可；要用 JSON 就把 `format=protobuf` 改成 `format=json` 或直接去掉。

