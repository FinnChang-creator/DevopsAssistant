package knowledge

import (
	"context"

	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/markdown"
	"github.com/cloudwego/eino/components/document"
	"github.com/google/uuid"
)

func newTransformer(ctx context.Context) (tfr document.Transformer, err error) {
	return markdown.NewHeaderSplitter(
		ctx,
		&markdown.HeaderConfig{
			Headers: map[string]string{
				"#": "title",
			},
			TrimHeaders: false,
			IDGenerator: func(ctx context.Context, originalID string, splitIndex int) string {
				return uuid.NewString()
			},
		},
	)
}
