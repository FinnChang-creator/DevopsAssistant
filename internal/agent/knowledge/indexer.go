package knowledge

import (
	"context"

	"github.com/FinnChang-creator/DevopsAssistant/internal/config"
	redisext "github.com/FinnChang-creator/DevopsAssistant/internal/infrastructure/redis"
	"github.com/cloudwego/eino-ext/components/indexer/redis"
	"github.com/cloudwego/eino/components/indexer"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
)

func newIndexer(ctx context.Context) (idx indexer.Indexer, err error) {
	em, err := newEmbedding(ctx)
	if err != nil {
		return nil, err
	}
	cfg := redis.IndexerConfig{
		Client:    redisext.Client,
		Embedding: em,
		KeyPrefix: config.EmbeddingConfig.KeyPrefix,
		BatchSize: config.EmbeddingConfig.BatchSize,
		DocumentToHashes: func(ctx context.Context, doc *schema.Document) (*redis.Hashes, error) {
			if doc.ID == "" {
				doc.ID = generateDocID()
			}
			return &redis.Hashes{
				Key: doc.ID,
				Field2Value: map[string]redis.FieldValue{
					"content": {
						Value:    doc.Content,
						EmbedKey: "embedding",
					},
					"name": {
						Value: doc.MetaData["title"],
					},
					"description": {
						Value: getDescription(doc.Content),
					},
					"price": {
						Value: 0,
					},
				},
			}, nil
		},
	}
	idx, err = redis.NewIndexer(ctx, &cfg)
	if err != nil {
		return nil, err
	}
	return idx, err
}

func generateDocID() string {
	return uuid.NewString()
}

func getDescription(content string) string {
	if len(content) > 100 {
		return content[:100] + "..."
	}
	return content
}
