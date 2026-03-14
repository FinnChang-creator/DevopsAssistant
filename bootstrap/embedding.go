package bootstrap

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
)

var embeddingConfig EmbeddingModelConfig

func initEmbeddingConfig(ctx context.Context) {
	cfg := g.Cfg().MustGet(ctx, "embedding_model")
	if cfg.IsEmpty() {
		panic("Embedding配置加载失败")
	}
	if err := cfg.Struct(&embeddingConfig); err != nil {
		panic("Embedding配置失败")
	}
}
func GetEmbeddingConfig() EmbeddingModelConfig {
	return embeddingConfig
}

type EmbeddingModelConfig struct {
	Choose    string    `yaml:"choose"`
	BatchSize int       `yaml:"batch_size"`
	KeyPrefix string    `yaml:"key_prefix"`
	Dashscope DashScope `yaml:"dashscope"`
}

type DashScope struct {
	APIKey     string `yaml:"api_key"`    // 阿里云DashScope API密钥
	BaseURL    string `yaml:"base_url"`   // API请求基础地址
	Model      string `yaml:"model"`      // 嵌入模型名称
	Dimensions int    `yaml:"dimensions"` // 嵌入向量维度
}
