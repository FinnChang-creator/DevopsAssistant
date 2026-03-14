package embedding

import (
	"context"

	"github.com/FinnChang-creator/DevopsAssistant/bootstrap"
	"github.com/cloudwego/eino-ext/components/embedding/dashscope"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/gogf/gf/v2/frame/g"
)

func NewDashscopeEmbedding(ctx context.Context) (eb embedding.Embedder, err error) {
	cfg := bootstrap.GetEmbeddingConfig()
	defer g.Log().Info(ctx, "Dashscope Embedder成功")
	// fmt.Print(cfg.Model.Model)
	// fmt.Print(cfg.Model.APIKey)
	// fmt.Print(cfg.Model.BaseURL)
	return dashscope.NewEmbedder(
		ctx,
		&dashscope.EmbeddingConfig{
			Model:      cfg.Dashscope.Model,
			APIKey:     cfg.Dashscope.APIKey,
			Dimensions: &cfg.Dashscope.Dimensions,
		},
	)
}
