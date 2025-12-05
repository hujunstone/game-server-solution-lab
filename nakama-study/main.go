package main

import (
	"context"
	"database/sql"
	"nakama/demo"  // 引入 demo 包
	"nakama/login" // 引入 login 包

	"github.com/heroiclabs/nakama-common/runtime"
)

// InitModule 是 Nakama 加载插件时的入口函数
func InitModule(ctx context.Context, logger runtime.Logger, db *sql.DB, nk runtime.NakamaModule, initializer runtime.Initializer) error {
	logger.Info("初始化 Nakama Go 模块...")

	// 注册登录相关的 Hook
	if err := login.RegisterLoginHooks(initializer); err != nil {
		logger.Error("注册登录 Hook 失败: %v", err)
		return err
	}

	// 注册 demo 相关的 RPC
	if err := demo.RegisterDemo(logger, initializer); err != nil {
		logger.Error("注册 demo 模块失败: %v", err)
		return err
	}

	logger.Info("Nakama Go 模块加载完成")
	return nil
}
