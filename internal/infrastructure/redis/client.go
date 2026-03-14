package redis

import (
	"context"
	"fmt"
	"strings"

	"github.com/FinnChang-creator/DevopsAssistant/internal/config"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/redis/go-redis/v9"
)

var (
	Client *redis.Client
)

func Init(ctx context.Context) error {
	Client = redis.NewClient(config.RedisConfig.ToRedisOptions().(*redis.Options))
	if err := Client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis连接失败: %w", err)
	}
	g.Log().Info(ctx, "Redis初始化成功")
	
	if has, err := hasRedisStack(ctx); err != nil {
		return err
	} else if !has {
		return fmt.Errorf("没有Redis Stack模块")
	}
	g.Log().Info(ctx, "加载Redis Stack模块成功")
	return nil
}

func hasRedisStack(ctx context.Context) (bool, error) {
	modules, err := Client.Do(ctx, "MODULE", "LIST").Result()
	if err != nil {
		return false, fmt.Errorf("执行 MODULE LIST 失败：%w", err)
	}

	modulesStr := fmt.Sprint(modules)
	if strings.Contains(modulesStr, "redis-stack-server") {
		return true, nil
	}

	return false, fmt.Errorf("没有Redis Stack模块")
}
