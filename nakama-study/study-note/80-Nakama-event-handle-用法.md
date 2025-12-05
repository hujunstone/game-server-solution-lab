### 简要回答

- `internal/runtime/events` 和 `internal/runtime/hooks` 里的代码 **都是服务端代码**，**不会被客户端直接调用**。  
- 它们是由 **Nakama 引擎（或服务端自己的业务代码）回调/触发** 的，用来在特定时机执行逻辑，然后**间接**影响客户端（例如通过 RPC 响应或 Realtime 推送）。

下面分开说一下机制：

### `hooks/`：Nakama 生命周期钩子（服务端被动回调）

- 典型写法是在插件初始化里注册，例如：

```go
func InitModule(ctx context.Context, nk runtime.NakamaModule, logger runtime.Logger) error {
    if err := nk.RegisterBeforeAuthenticateDevice(beforeAuthDeviceHook); err != nil {
        return err
    }
    if err := nk.RegisterAfterAuthenticateDevice(afterAuthDeviceHook); err != nil {
        return err
    }
    return nil
}
```

- **触发链路**：
  1. 客户端调用 Nakama 标准接口（如登录 `AuthenticateDevice`）。  
  2. Nakama 核心处理基础逻辑，然后在流程前/后自动调用你在 `hooks/` 里注册的 Go 函数。  
  3. 这些 Hook 可以校验参数、初始化 Session、加载角色数据等，然后由 Nakama 把最终结果返回给客户端。

> 所以：**客户端只知道自己在调用“登录”等标准接口，不知道有 hooks 存在；hooks 完全是服务端内部扩展点。**

### `events/`：服务端内部事件总线（服务端主动触发）

- 这里是你自己在服务端实现的一层“事件中心”，例如：

```go
type EventBus interface {
    PublishBattleFinished(ctx context.Context, payload BattleResult)
    SubscribeBattleFinished(handler func(BattleResult))
}
```

- **触发链路**（一个例子）：
  1. 某个 RPC 或 Match 逻辑在 `services/` 中完成战斗结算逻辑后，调用 `events.PublishBattleFinished(...)`。  
  2. `events` 模块收到事件，调用你注册的回调：  
     - 更新统计、写日志；  
     - 通过 Nakama 的 `nk.StreamSend` / Realtime API 往 `rt.battle.<encounter_id>` 等频道推送消息给在线客户端。
  3. UE 客户端只是订阅了这些 `rt.*` 频道，收到数据后更新界面。

> 所以：**`events/` 完全由服务端主动调用和分发，客户端只消费结果（推送数据），不会直接触发 `events` 里的函数。**

### 总结一句

- `hooks/`：**Nakama 核心在“登录/存储/匹配等生命周期点”自动回调的服务端扩展函数。**
- `events/`：**服务端业务代码（`services/` 等）主动发布/订阅的内部事件，用来统一做后续推送、日志、统计等。**
- 客户端只跟 **RPC / Realtime 频道** 打交道，**不会直接调用这两个目录里的函数**。