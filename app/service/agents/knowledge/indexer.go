package knowledge

import (
	"context"

	"github.com/FinnChang-creator/DevopsAssistant/bootstrap"
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
		Client:    bootstrap.GetRedisClient(),
		Embedding: em,
		KeyPrefix: bootstrap.GetEmbeddingConfig().KeyPrefix,
		BatchSize: bootstrap.GetEmbeddingConfig().BatchSize,
		// 自定义 DocumentToHashes，使字段名与 Redis 索引配置匹配
		DocumentToHashes: func(ctx context.Context, doc *schema.Document) (*redis.Hashes, error) {
			if doc.ID == "" {
				doc.ID = generateDocID()
			}
			return &redis.Hashes{
				Key: doc.ID,
				Field2Value: map[string]redis.FieldValue{
					// content 字段存储原始内容，embedding 字段存储向量
					"content": {
						Value:    doc.Content,
						EmbedKey: "embedding", // 向量将存储在 embedding 字段，与索引配置匹配
					},
					// 将 title 从 metadata 中提取存储到 name 字段
					"name": {
						Value: doc.MetaData["title"],
					},
					// description 字段存储文档内容摘要
					"description": {
						Value: getDescription(doc.Content),
					},
					// price 字段，默认为 0
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

// generateDocID 生成文档 ID
func generateDocID() string {
	return uuid.NewString()
}

// getDescription 获取内容摘要
func getDescription(content string) string {
	if len(content) > 100 {
		return content[:100] + "..."
	}
	return content
}
