package login

import (
	"context"
	"database/sql"
	"strings"

	"github.com/heroiclabs/nakama-common/api"
	"github.com/heroiclabs/nakama-common/runtime"
)

// RegisterLoginHooks 注册钩子函数
func RegisterLoginHooks(initializer runtime.Initializer) error {
	// 1. 注册 Before Hook：在 Nakama 处理登录前执行（用于校验）
	if err := initializer.RegisterBeforeAuthenticateEmail(BeforeAuthenticateEmail); err != nil {
		return err
	}

	// 2. 注册 After Hook：在 Nakama 处理登录成功后执行（用于初始化数据）
	if err := initializer.RegisterAfterAuthenticateEmail(AfterAuthenticateEmail); err != nil {
		return err
	}

	return nil
}

// BeforeAuthenticateEmail: 登录前的校验逻辑
func BeforeAuthenticateEmail(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, in *api.AuthenticateEmailRequest) (*api.AuthenticateEmailRequest, error) {
	// 防御式编程：避免空指针
	if in == nil || in.Account == nil {
		logger.Warn("登录被拒绝，邮箱账户信息为空")
		return nil, runtime.NewError("邮箱账户信息无效", 3)
	}

	email := strings.TrimSpace(in.Account.Email)
	logger.Info("收到邮箱登录请求: %v", email)

	// 示例逻辑：只允许 @example.com 的邮箱登录或注册
	// if email == "" || !strings.HasSuffix(email, "@example.com") {
	// 	logger.Warn("登录被拒绝，非内部邮箱或邮箱为空: %s", email)
	// 	// 返回错误，客户端会收到 HTTP 400 或 500
	// 	return nil, runtime.NewError("仅允许 @example.com 用户登录", 3) // 3 = Invalid Argument
	// }

	// 允许通过
	return in, nil
}

// AfterAuthenticateEmail: 登录后的初始化逻辑
func AfterAuthenticateEmail(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, out *api.Session, in *api.AuthenticateEmailRequest) error {
	// 从上下文中获取用户 ID（Nakama 在钩子上下文里注入了该值）
	userID, ok := ctx.Value(runtime.RUNTIME_CTX_USER_ID).(string)
	if !ok || userID == "" {
		logger.Error("无法从上下文中获取用户 ID")
		return runtime.NewError("internal error: missing user id", 13)
	}
	logger.Info("用户登录成功，ID: %s. 正在检查是否需要初始化数据...", userID)

	// 检查该用户是否是新用户 (out.Created 在注册时为 true)
	if out.Created {
		logger.Info("检测到新用户注册，发放初始奖励...")

		// 定义初始钱包内容
		changes := map[string]int64{
			"gold": 1000,
			"gems": 50,
		}

		// 定义写操作元数据
		metadata := map[string]interface{}{"source": "welcome_bonus"}

		// 更新钱包
		_, _, err := nk.WalletUpdate(ctx, userID, changes, metadata, true)
		if err != nil {
			logger.Error("发放初始奖励失败: %v", err)
			return err
		}
		logger.Info("初始奖励发放完毕")
	}

	return nil
}
