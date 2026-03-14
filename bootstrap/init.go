package bootstrap

import (
	"context"
	"sync"

	"github.com/FinnChang-creator/DevopsAssistant/internal/config"
	"github.com/FinnChang-creator/DevopsAssistant/internal/infrastructure/logger"
	"github.com/FinnChang-creator/DevopsAssistant/internal/infrastructure/redis"
)

var once sync.Once

func Init(ctx context.Context) {
	once.Do(func() {
		logger.Init(ctx)
		if err := config.Init(ctx); err != nil {
			panic(err)
		}
		if err := redis.Init(ctx); err != nil {
			panic(err)
		}
		if err := redis.InitVector(ctx); err != nil {
			panic(err)
		}
	})
}
