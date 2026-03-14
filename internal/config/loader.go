package config

import (
	"context"
	"fmt"

	"github.com/gogf/gf/v2/frame/g"
)

func Load(ctx context.Context) error {
	// 获取 embedding_model 配置
	embeddingCfg, err := g.Cfg().Get(ctx, "embedding_model")
	if err != nil {
		return fmt.Errorf("获取 embedding_model 配置失败: %w", err)
	}
	fmt.Printf("[debug] embedding_model 原始配置: %v\n", embeddingCfg)

	// 解析到结构体
	if err = embeddingCfg.Struct(&EmbeddingConfig); err != nil {
		return fmt.Errorf("解析 EmbeddingConfig 失败: %w", err)
	}
	fmt.Printf("[debug] EmbeddingConfig 解析结果: Choose=%s, Model=%s, APIKey=%s\n",
		EmbeddingConfig.Choose,
		EmbeddingConfig.Dashscope.Model,
		EmbeddingConfig.Dashscope.APIKey)

	// 获取 redis 配置
	redisCfg, err := g.Cfg().Get(ctx, "redis")
	if err != nil {
		return fmt.Errorf("获取 redis 配置失败: %w", err)
	}
	if err = redisCfg.Struct(&RedisConfig); err != nil {
		return fmt.Errorf("解析 RedisConfig 失败: %w", err)
	}

	return nil
}
