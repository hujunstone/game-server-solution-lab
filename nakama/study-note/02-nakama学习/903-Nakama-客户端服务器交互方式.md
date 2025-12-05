[10 tools called]

### 简单结论

**Nakama 和客户端（包括蓝图 SDK）之间，既用 Protobuf，也用 JSON：**

- **gRPC 通道**：纯 **Protobuf 二进制**（pb）。
- **HTTP/REST 通道**：是 **JSON**，但格式由 Protobuf 定义（`protojson` 编解码）。
- **WebSocket 实时通道**：同时支持 **JSON 和 Protobuf** 两种格式，客户端用 URL 上的 `format=json|protobuf` 决定；蓝图 SDK 一般走这一层，所以**理论上两种都支持，取决于 SDK 配置**。

下面是源码里“pb vs json”体现最明显的几个点。

---

### 1. gRPC：纯 Protobuf

在 `server/api.go` 里启动 gRPC 服务：

```156:167:server/api.go
	// Register and start GRPC server.
	apigrpc.RegisterNakamaServer(grpcServer, s)
	startupLogger.Info("Starting API server for gRPC requests", zap.Int("port", config.GetSocket().Port-1))
	go func() {
		listener, err := net.Listen("tcp", fmt.Sprintf("%v:%d", config.GetSocket().Address, config.GetSocket().Port-1))
		if err != nil {
			startupLogger.Fatal("API server listener failed to start", zap.Error(err))
		}

		if err := grpcServer.Serve(listener); err != nil {
			startupLogger.Fatal("API server listener failed", zap.Error(err))
		}
	}()
```

- `apigrpc.RegisterNakamaServer` 是由 `apigrpc/apigrpc.proto` 生成的 gRPC Server 接口 → **gRPC 使用的就是 Protobuf 编码**。

---

### 2. HTTP/REST：JSON（由 Protobuf 映射）

同一个文件中，通过 grpc‑gateway 暴露 HTTP API 时，显式指定了 JSON 编解码器：

```192:201:server/api.go
	grpcGateway := grpcgw.NewServeMux(
		grpcgw.WithRoutingErrorHandler(handleRoutingError),
		grpcgw.WithMetadata(func(ctx context.Context, r *http.Request) metadata.MD {
			// 省略中间逻辑...
			return p
		}),
		grpcgw.WithMarshalerOption(grpcgw.MIMEWildcard, &grpcgw.HTTPBodyMarshaler{
			Marshaler: &grpcgw.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					UseProtoNames:  true,
					UseEnumNumbers: true,
				},
				UnmarshalOptions: protojson.UnmarshalOptions{
					DiscardUnknown: true,
				},
			},
		}),
	)
```

- 这里使用 `grpcgw.JSONPb` + `protojson`，说明：
    - **HTTP 请求/响应体是 JSON**；
    - 字段名、枚举值等直接来自 Protobuf 定义（`nakama-common/api/api.proto`）。

---

### 3. WebSocket：JSON / Protobuf 二选一

#### 3.1 通过 `format` 参数选择格式

在 WebSocket 接入层 `server/socket_ws.go`：

```44:53:server/socket_ws.go
		// Check format.
		var format SessionFormat
		switch r.URL.Query().Get("format") {
		case "", "json":
			format = SessionFormatJson
		case "protobuf":
			format = SessionFormatProtobuf
		default:
			// Invalid values are rejected.
			http.Error(w, "Invalid format parameter", 400)
			return
		}
```

- `/ws?format=json` 或不填 → **JSON 文本帧**。
- `/ws?format=protobuf` → **Protobuf 二进制帧**。
- 这正是蓝图 SDK 这类实时客户端与 Nakama 通讯时可选择的格式开关。

#### 3.2 消息结构：都是 Protobuf 定义的 `rtapi.Envelope`

在 `server/session_ws.go` 中，收发时根据 `SessionFormat` 决定用 JSON 还是 Protobuf：

**接收：**

```245:253:server/session_ws.go
		request := &rtapi.Envelope{}
		switch s.format {
		case SessionFormatProtobuf:
			err = proto.Unmarshal(data, request)
		case SessionFormatJson:
			fallthrough
		default:
			err = s.protojsonUnmarshaler.Unmarshal(data, request)
		}
```

**发送：**

```389:399:server/session_ws.go
func (s *sessionWS) Send(envelope *rtapi.Envelope, reliable bool) error {
	var payload []byte
	var err error
	switch s.format {
	case SessionFormatProtobuf:
		payload, err = proto.Marshal(envelope)
	case SessionFormatJson:
		fallthrough
	default:
		if buf, err := s.protojsonMarshaler.Marshal(envelope); err == nil {
			payload = buf
		}
	}
	// ...
}
```

- `rtapi.Envelope` 定义在 `github.com/heroiclabs/nakama-common/rtapi` 的 `.proto` 文件里。
- **无论 JSON 还是 Protobuf，本质上消息结构都由 Protobuf schema 定义，只是编码格式不同。**

---

### 4. 回到“蓝图 SDK”这一点

- 从 **服务器源码角度**：
    - gRPC：只能是 **pb**。
    - HTTP：只能是 **JSON**（按 proto 映射）。
    - WebSocket：支持 **JSON 或 pb**，由 `format` 控制。
- Unreal Blueprint SDK 这类客户端通常用 WebSocket 实时通道，所以：
    - **可以用 JSON（默认）或 Protobuf 两种格式，服务器完全支持，两种处理逻辑都已经在 `socket_ws.go` + `session_ws.go` 里实现。**
    - 具体 SDK 默认用哪一种，要看 SDK 那边的实现/配置，但你要确认协议支持情况，源码已经证明“**两种都支持**”。

如果你希望，我可以帮你总结一份“蓝图 SDK 连接参数/URL 应该怎么写（如何指定 format=protobuf/json）”的示例。