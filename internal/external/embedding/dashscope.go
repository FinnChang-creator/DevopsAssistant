package embedding

import (
	"context"

	"github.com/FinnChang-creator/DevopsAssistant/internal/config"
	"github.com/cloudwego/eino-ext/components/embedding/dashscope"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/gogf/gf/v2/frame/g"
)

func NewDashscopeEmbedding(ctx context.Context) (eb embedding.Embedder, err error) {
	cfg := config.EmbeddingConfig
	defer g.Log().Info(ctx, "Dashscope Embedder成功")
	return dashscope.NewEmbedder(
		ctx,
		&dashscope.EmbeddingConfig{
			Model:      cfg.Dashscope.Model,
			APIKey:     cfg.Dashscope.APIKey,
			Dimensions: &cfg.Dashscope.Dimensions,
		},
	)
}
