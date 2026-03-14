package config

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

func Load(ctx context.Context) error {
	cfg, err := g.Cfg().Get(ctx, "")
	if err != nil {
		return err
	}
	if err = cfg.Struct(&EmbeddingConfig); err != nil {
		return err
	}
	if err = cfg.Struct(&RedisConfig); err != nil {
		return err
	}
	return nil
}
