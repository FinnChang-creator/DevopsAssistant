package knowledge

import (
	"context"

	"github.com/cloudwego/eino-ext/components/document/loader/file"
	"github.com/cloudwego/eino/components/document"
)

func newLoader(ctx context.Context) (ldr document.Loader, err error) {
	return file.NewFileLoader(
		ctx,
		&file.FileLoaderConfig{
			UseNameAsID: false,
		},
	)
}
