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
	Client = redis.NewClient(config.RedisConfig.ToRedisOptions())
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
	result, err := Client.Do(ctx, "MODULE", "LIST").Result()
	if err != nil {
		return false, fmt.Errorf("获取Redis模块列表失败: %w", err)
	}

	modules, ok := result.([]interface{})
	if !ok {
		return false, fmt.Errorf("Redis模块列表格式错误")
	}

	for _, module := range modules {
		moduleInfo, ok := module.([]interface{})
		if !ok || len(moduleInfo) < 2 {
			continue
		}

		name, ok := moduleInfo[1].(string)
		if !ok {
			continue
		}

		if strings.EqualFold(name, "searchlight") || strings.EqualFold(name, "redisearch") {
			return true, nil
		}
	}

	return false, nil
}
