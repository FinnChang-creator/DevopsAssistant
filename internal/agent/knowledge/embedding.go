package knowledge

import (
	"context"

	"github.com/FinnChang-creator/DevopsAssistant/internal/external/embedding"
	einocomponents "github.com/cloudwego/eino/components/embedding"
)

func newEmbedding(ctx context.Context) (eb einocomponents.Embedder, err error) {
	return embedding.NewDashscopeEmbedding(ctx)
}
