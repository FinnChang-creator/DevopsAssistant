// Package bootstrap 程序启动初始化包
// 负责统一初始化日志、Redis/Redis Stack 等核心组件
package bootstrap

import (
	"context"
	"fmt"
	"sync"
)

var once sync.Once

func Init(ctx context.Context) {
	once.Do(func() {
		initLogger(ctx)
		initRedis(ctx)
		initRedisVector(ctx)
		initEmbeddingConfig(ctx)
	})
	fmt.Print(embeddingConfig)
}
