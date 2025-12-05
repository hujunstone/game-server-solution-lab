package demo

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/heroiclabs/nakama-common/runtime"
)

// 请求参数：客户端传 {"code":"JINGSHANG"}
type FactionQuery struct {
	Code string `json:"code"`
}

// 查询结果：返回给客户端的 JSON
type Faction struct {
	// 势力主键
	FactionID int32 `db:"faction_id" json:"faction_id"`
	// 势力编码
	Code string `db:"code" json:"code"`
	// 势力名称
	Name string `db:"name" json:"name"`
	// 势力徽记资源引用
	BadgeIcon string `db:"badge_icon" json:"badge_icon"`
	// 势力说明文本
	Description string `db:"description" json:"description"`
}

// RegisterDemo 由主插件入口调用，用来注册 demo 相关的 RPC。
func RegisterDemo(logger runtime.Logger, initializer runtime.Initializer) error {
	// 注册一个自定义 RPC：id = "demo_faction_get"
	if err := initializer.RegisterRpc("demo_faction_get", rpcGetFaction); err != nil {
		return err
	}
	logger.Info("demo_faction_get RPC registered")
	return nil
}

// 具体 RPC 处理逻辑。
// payload 是客户端传来的字符串（一般是 JSON 字符串）。
// 返回值 string 会原样作为 payload 返回给客户端（外面再包一层 rtapi.Envelope，由 Nakama 处理）。
func rpcGetFaction(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, payload string) (string, error) {
	logger.Info("demo_faction_get called with payload: %s", payload)

	// 1. 解析请求 JSON
	var req FactionQuery
	if payload != "" {
		if err := json.Unmarshal([]byte(payload), &req); err != nil {
			return "", fmt.Errorf("invalid payload json: %w", err)
		}
	}
	if req.Code == "" {
		return "", errors.New("code is required")
	}

	// 2. 查询数据库 demo_faction 表
	const query = `
SELECT faction_id, code, name, badge_icon, description
FROM demo_faction
WHERE code = $1
`
	row := db.QueryRowContext(ctx, query, req.Code)

	var f Faction
	if err := row.Scan(&f.FactionID, &f.Code, &f.Name, &f.BadgeIcon, &f.Description); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("faction not found: %s", req.Code)
		}
		return "", fmt.Errorf("db query error: %w", err)
	}

	// 3. 序列化为 JSON 字符串返回（Nakama 会把这个 string 放到 Rpc 响应的 payload 字段）。
	out, err := json.Marshal(f)
	if err != nil {
		return "", fmt.Errorf("marshal response error: %w", err)
	}

	// 返回 JSON 字符串；外层 pb/JSON Envelope 由 Nakama 负责。
	return string(out), nil
}
