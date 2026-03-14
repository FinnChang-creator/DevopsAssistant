package knowledge

import (
	"context"

	em "github.com/FinnChang-creator/DevopsAssistant/app/service/embedding"
	"github.com/FinnChang-creator/DevopsAssistant/bootstrap"
	"github.com/cloudwego/eino/components/embedding"
)

func newEmbedding(ctx context.Context) (eb embedding.Embedder, err error) {
	switch bootstrap.GetEmbeddingConfig().Choose {
	case "dashscope":
		return em.NewDashscopeEmbedding(ctx)
	default:
		return em.NewDashscopeEmbedding(ctx)
	}
}
