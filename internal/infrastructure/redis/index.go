package redis

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

var (
	VectorIndexes    map[string]VectorIndex
	VectorOperations VectorOperationsConfig
	VectorTuning     VectorTuningConfig
)

func InitVector(ctx context.Context) error {
	cfg, err := g.Cfg().Get(ctx, "redis_vector_indexes")
	if err != nil {
		return fmt.Errorf("redis_vector_indexes读取失败: %v", err)
	}
	if err = cfg.Struct(&VectorIndexes); err != nil {
		return fmt.Errorf("redis_vector_indexes结构化失败: %v", err)
	}

	cfg, err = g.Cfg().Get(ctx, "redis_vector_operations")
	if err != nil {
		return fmt.Errorf("redis_vector_operations读取失败: %v", err)
	}
	if err = cfg.Struct(&VectorOperations); err != nil {
		return fmt.Errorf("redis_vector_operations结构化失败: %v", err)
	}

	cfg, err = g.Cfg().Get(ctx, "redis_vector_tuning")
	if err != nil {
		return fmt.Errorf("redis_vector_tuning读取失败: %v", err)
	}
	if err = cfg.Struct(&VectorTuning); err != nil {
		return fmt.Errorf("redis_vector_tuning结构化失败: %v", err)
	}

	g.Log().Info(ctx, "Redis Vector模块加载成功")
	return setVector(ctx)
}

func setVector(ctx context.Context) error {
	indexNameMap := make(map[string]int)
	for alias, index := range VectorIndexes {
		if index.IndexName == "" {
			return fmt.Errorf("索引别名[%s]的index_name为空", alias)
		}
		indexNameMap[index.IndexName]++
		if indexNameMap[index.IndexName] > 1 {
			return fmt.Errorf("配置中存在重复的索引名: %s", index.IndexName)
		}
	}

	for indexAlias, indexConfig := range VectorIndexes {
		g.Log().Infof(ctx, "开始处理索引[别名:%s, 名称:%s]", indexAlias, indexConfig.IndexName)

		exists, err := checkIndexExists(ctx, indexConfig.IndexName)
		if err != nil {
			return fmt.Errorf("检查索引[%s]是否存在失败: %v", indexConfig.IndexName, err)
		}

		if exists {
			if indexConfig.DropIfExists {
				g.Log().Warningf(ctx, "索引[%s]已存在，删除旧索引", indexConfig.IndexName)
				if err := deleteRedisIndex(ctx, indexConfig.IndexName); err != nil {
					return fmt.Errorf("删除索引[%s]失败: %v", indexConfig.IndexName, err)
				}
			} else {
				return fmt.Errorf("索引[%s]已存在且drop_if_exists=false，终止初始化", indexConfig.IndexName)
			}
		}

		if err := createVectorIndex(ctx, indexConfig); err != nil {
			return fmt.Errorf("创建索引[%s]失败: %v", indexConfig.IndexName, err)
		}

		g.Log().Infof(ctx, "索引[别名:%s, 名称:%s]创建成功", indexAlias, indexConfig.IndexName)
	}
	return nil
}

func checkIndexExists(ctx context.Context, indexName string) (bool, error) {
	indexList, err := Client.Do(ctx, "FT._LIST").Result()
	if err != nil {
		if strings.Contains(err.Error(), "unknown command") {
			return false, errors.New("Redis未加载RediSearch模块,不支持向量索引")
		}
		return false, err
	}

	switch v := indexList.(type) {
	case []interface{}:
		for _, item := range v {
			if item.(string) == indexName {
				return true, nil
			}
		}
	case string:
		if v == indexName {
			return true, nil
		}
	}
	return false, nil
}

func deleteRedisIndex(ctx context.Context, indexName string) error {
	return Client.Do(ctx, "FT.DROPINDEX", indexName, "DD").Err()
}

func createVectorIndex(ctx context.Context, indexConfig VectorIndex) error {
	dataType := "HASH"
	prefix := indexConfig.IndexName + ":"

	cmdBuilder := []string{
		fmt.Sprintf("FT.CREATE %s", indexConfig.IndexName),
		fmt.Sprintf("ON %s", dataType),
		fmt.Sprintf("PREFIX 1 %s", prefix),
		"SCHEMA",
	}

	var fields []string
	if dataType == "JSON" {
		fields = []string{
			"$.name AS name TEXT",
			"$.description AS description TEXT",
			"$.price AS price NUMERIC",
			fmt.Sprintf("$.%s AS %s VECTOR %s %s DIM %d DISTANCE_METRIC %s TYPE %s",
				indexConfig.VectorField.FieldName,
				indexConfig.VectorField.FieldName,
				indexConfig.VectorField.Algorithm,
				getVecTypeCode(indexConfig.VectorField.DataType),
				indexConfig.VectorField.Dim,
				indexConfig.VectorField.DistanceMetric,
				indexConfig.VectorField.DataType,
			),
		}
	} else {
		fields = []string{
			"name TEXT",
			"description TEXT",
			"price NUMERIC",
			fmt.Sprintf("%s VECTOR %s %s DIM %d DISTANCE_METRIC %s TYPE %s",
				indexConfig.VectorField.FieldName,
				indexConfig.VectorField.Algorithm,
				getVecTypeCode(indexConfig.VectorField.DataType),
				indexConfig.VectorField.Dim,
				indexConfig.VectorField.DistanceMetric,
				indexConfig.VectorField.DataType,
			),
		}
	}

	if indexConfig.VectorField.Algorithm == "HNSW" {
		if indexConfig.VectorField.HnswParams.M <= 0 {
			return fmt.Errorf("HNSW算法[%s]：M参数必须>0", indexConfig.IndexName)
		}
		for i, field := range fields {
			if strings.Contains(field, indexConfig.VectorField.FieldName+" VECTOR") {
				fields[i] = fmt.Sprintf("%s M %d", field, indexConfig.VectorField.HnswParams.M)
				break
			}
		}
	}

	cmdBuilder = append(cmdBuilder, fields...)

	cmdStr := strings.Join(cmdBuilder, " ")
	g.Log().Infof(ctx, "创建索引命令：%s", cmdStr)

	strArgs := strings.Fields(cmdStr)
	cmdArgs := make([]interface{}, len(strArgs))
	for i, s := range strArgs {
		cmdArgs[i] = s
	}

	_, err := Client.Do(ctx, cmdArgs...).Result()
	if err != nil {
		return fmt.Errorf(
			"创建索引[%s]失败！\n最终命令：%s\n错误详情：%w\n【手动验证】：复制上述命令到redis-cli执行",
			indexConfig.IndexName, cmdStr, err,
		)
	}

	g.Log().Infof(ctx, "索引[%s]创建成功（数据类型：%s）", indexConfig.IndexName, dataType)
	return nil
}

func getVecTypeCode(dataType string) string {
	switch dataType {
	case "FLOAT32":
		return "6"
	case "FLOAT64":
		return "7"
	default:
		return "6"
	}
}

type VectorIndex struct {
	IndexName    string      `yaml:"index_name"`
	DropIfExists bool        `yaml:"drop_if_exists"`
	VectorField  VectorField `yaml:"vector_field"`
}

type VectorField struct {
	FieldName      string     `yaml:"field_name"`
	Algorithm      string     `yaml:"algorithm"`
	Dim            int        `yaml:"dim"`
	DataType       string     `yaml:"data_type"`
	DistanceMetric string     `yaml:"distance_metric"`
	HnswParams     HnswParams `yaml:"hnsw_params,omitempty"`
}

type HnswParams struct {
	M              int `yaml:"M"`
	EfConstruction int `yaml:"ef_construction"`
	EfRuntime      int `yaml:"ef_runtime"`
}

type VectorOperationsConfig struct {
	Insert InsertConfig `yaml:"insert"`
	Search SearchConfig `yaml:"search"`
}

type InsertConfig struct {
	BatchSize    int    `yaml:"batch_size"`
	VectorEncode string `yaml:"vector_encode"`
}

type SearchConfig struct {
	TopK            int  `yaml:"top_k"`
	Timeout         int  `yaml:"timeout"`
	ReturnDistance  bool `yaml:"return_distance"`
	ReturnRawVector bool `yaml:"return_raw_vector"`
}

type VectorTuningConfig struct {
	HnswMemory     HnswMemoryConfig `yaml:"hnsw_memory,omitempty"`
	DistanceFilter DistanceFilter   `yaml:"distance_filter,omitempty"`
}

type HnswMemoryConfig struct {
	MaxConnections int `yaml:"max_connections"`
	GCInterval     int `yaml:"gc_interval"`
}

type DistanceFilter struct {
	CosineMax float64 `yaml:"cosine_max"`
	L2Max     float64 `yaml:"l2_max"`
	IPMin     float64 `yaml:"ip_min"`
}
