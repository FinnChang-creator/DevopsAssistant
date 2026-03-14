package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gogf/gf/v2/frame/g"
)

var (
	vectorIndexes    map[string]VectorIndex
	vectorOperations VectorOperations
	vectorTuning     VectorTuning
)

func initRedisVector(ctx context.Context) {
	cfg, err := g.Cfg().Get(ctx, "redis_vector_indexes")
	if err != nil {
		panic(fmt.Sprintf("redis_vector_indexes读取失败: %v", err))
	}
	if err = cfg.Struct(&vectorIndexes); err != nil {
		panic(fmt.Sprintf("redis_vector_indexes结构化失败: %v", err))
	}
	cfg, err = g.Cfg().Get(ctx, "redis_vector_indexes")
	if err != nil {
		panic(fmt.Sprintf("redis_vector_operations读取失败: %v", err))
	}
	if err = cfg.Struct(&vectorOperations); err != nil {
		panic(fmt.Sprintf("redis_vector_operations结构化失败: %v", err))
	}
	cfg, err = g.Cfg().Get(ctx, "redis_vector_tuning")
	if err != nil {
		panic(fmt.Sprintf("redis_vector_tuning读取失败: %v", err))
	}
	if err = cfg.Struct(&vectorTuning); err != nil {
		panic(fmt.Sprintf("redis_vector_tuning结构化失败: %v", err))
	}
	g.Log().Info(ctx, "Redis Vector模块加载成功")
	setVector(ctx)
}

func setVector(ctx context.Context) {
	// 前置校验：检查是否有重复索引名（全局去重）
	indexNameMap := make(map[string]int)
	for alias, index := range vectorIndexes {
		if index.IndexName == "" {
			panic(fmt.Sprintf("索引别名[%s]的index_name为空", alias))
		}
		indexNameMap[index.IndexName]++
		if indexNameMap[index.IndexName] > 1 {
			panic(fmt.Sprintf("配置中存在重复的索引名: %s", index.IndexName))
		}
	}

	// 遍历所有向量索引（key=索引别名，value=索引配置）
	for indexAlias, indexConfig := range vectorIndexes {
		g.Log().Infof(ctx, "开始处理索引[别名:%s, 名称:%s]", indexAlias, indexConfig.IndexName)

		// 步骤1：检查Redis中是否已存在该索引
		exists, err := checkIndexExists(ctx, indexConfig.IndexName)
		if err != nil {
			panic(fmt.Sprintf("检查索引[%s]是否存在失败: %v", indexConfig.IndexName, err))
		}

		// 步骤2：如果存在，根据drop_if_exists决定是否删除
		if exists {
			if indexConfig.DropIfExists {
				g.Log().Warning(ctx, "索引[%s]已存在，删除旧索引", indexConfig.IndexName)
				if err := deleteRedisIndex(ctx, indexConfig.IndexName); err != nil {
					panic(fmt.Sprintf("删除索引[%s]失败: %v", indexConfig.IndexName, err))
				}
			} else {
				panic(fmt.Sprintf("索引[%s]已存在且drop_if_exists=false，终止初始化", indexConfig.IndexName))
			}
		}

		// 步骤3：创建向量索引
		if err := createVectorIndex(ctx, indexConfig); err != nil {
			panic(fmt.Sprintf("创建索引[%s]失败: %v", indexConfig.IndexName, err))
		}

		g.Log().Infof(ctx, "索引[别名:%s, 名称:%s]创建成功", indexAlias, indexConfig.IndexName)
	}
}

func checkIndexExists(ctx context.Context, indexName string) (bool, error) {
	// FT._LIST 获取所有索引
	indexList, err := redisClient.Do(ctx, "FT._LIST").Result()
	if err != nil {
		if strings.Contains(err.Error(), "unknown command") {
			return false, errors.New("Redis未加载RediSearch模块,不支持向量索引")
		}
		return false, err
	}

	// 遍历索引列表判断是否存在
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

// deleteRedisIndex 删除Redis中的指定索引
func deleteRedisIndex(ctx context.Context, indexName string) error {
	// FT.DROPINDEX [索引名] DD ：删除索引并删除关联文档
	return redisClient.Do(ctx, "FT.DROPINDEX", indexName, "DD").Err()
}
func createVectorIndex(ctx context.Context, indexConfig VectorIndex) error {
	// ====================== 核心修正：合并JSON路径映射和向量类型，避免重复定义 ======================
	// 支持 JSON/HASH 切换（根据你的数据类型选择）
	dataType := "HASH"                    // HASH 改为 "HASH"
	prefix := indexConfig.IndexName + ":" // 或改为你示例中的 "item:"

	// 1. 基础命令构建
	cmdBuilder := []string{
		fmt.Sprintf("FT.CREATE %s", indexConfig.IndexName),
		fmt.Sprintf("ON %s", dataType),
		fmt.Sprintf("PREFIX 1 %s", prefix),
		"SCHEMA",
	}

	// 2. 字段定义（区分JSON/HASH）
	var fields []string
	if dataType == "JSON" {
		// JSON类型：合并JSON路径映射 + 向量类型（无重复）
		fields = []string{
			"$.name AS name TEXT",
			"$.description AS description TEXT",
			"$.price AS price NUMERIC",
			// 核心：JSON路径映射和向量类型合并为一行
			fmt.Sprintf("$.%s AS %s VECTOR %s %s DIM %d DISTANCE_METRIC %s TYPE %s",
				indexConfig.VectorField.FieldName,                // JSON路径：$.embedding
				indexConfig.VectorField.FieldName,                // 别名：embedding
				indexConfig.VectorField.Algorithm,                // FLAT/HNSW
				getVecTypeCode(indexConfig.VectorField.DataType), // 6(FLOAT32)/7(FLOAT64)
				indexConfig.VectorField.Dim,                      // 维度
				indexConfig.VectorField.DistanceMetric,           // 距离度量
				indexConfig.VectorField.DataType,                 // 数据类型
			),
		}
	} else {
		// HASH类型：直接定义向量字段
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

	// 3. HNSW算法补充M参数（仅旧版语法需要）
	if indexConfig.VectorField.Algorithm == "HNSW" {
		if indexConfig.VectorField.HnswParams.M <= 0 {
			return fmt.Errorf("HNSW算法[%s]：M参数必须>0", indexConfig.IndexName)
		}
		// 找到向量字段行，追加M参数
		for i, field := range fields {
			if strings.Contains(field, indexConfig.VectorField.FieldName+" VECTOR") {
				fields[i] = fmt.Sprintf("%s M %d", field, indexConfig.VectorField.HnswParams.M)
				break
			}
		}
	}

	// 4. 合并字段到命令
	cmdBuilder = append(cmdBuilder, fields...)

	// 5. 拼接最终命令
	cmdStr := strings.Join(cmdBuilder, " ")
	g.Log().Infof(ctx, "修正后最终命令：%s", cmdStr)

	// 6. 转换参数类型并执行
	strArgs := strings.Fields(cmdStr)
	cmdArgs := make([]interface{}, len(strArgs))
	for i, s := range strArgs {
		cmdArgs[i] = s
	}

	_, err := redisClient.Do(ctx, cmdArgs...).Result()
	if err != nil {
		return fmt.Errorf(
			"创建索引[%s]失败！\n最终命令：%s\n错误详情：%w\n【手动验证】：复制上述命令到redis-cli执行",
			indexConfig.IndexName, cmdStr, err,
		)
	}

	g.Log().Infof(ctx, "索引[%s]创建成功（数据类型：%s）", indexConfig.IndexName, dataType)
	return nil
}

// 辅助函数：数据类型转旧版编码（FLOAT32=6，FLOAT64=7）
func getVecTypeCode(dataType string) string {
	switch dataType {
	case "FLOAT32":
		return "6"
	case "FLOAT64":
		return "7"
	default:
		return "6" // 默认FLOAT32
	}
}

// ====================== 根级配置结构体（匹配整个YAML） ======================

type RedisVectorConfig struct {
	RedisVectorIndexes    map[string]VectorIndex `yaml:"redis_vector_indexes"`          // 向量索引配置（key为索引别名：vector_index/text_semantic_index等）
	RedisVectorOperations VectorOperations       `yaml:"redis_vector_operations"`       // 向量操作配置
	RedisVectorTuning     VectorTuning           `yaml:"redis_vector_tuning,omitempty"` // 向量性能调优（可选）
}

// ====================== 向量索引核心结构体（匹配redis_vector_indexes下的每个索引）

type VectorIndex struct {
	IndexName    string      `yaml:"index_name"`     // 向量索引名称
	DropIfExists bool        `yaml:"drop_if_exists"` // 创建前是否删除已有索引
	VectorField  VectorField `yaml:"vector_field"`   // 向量字段核心配置
}

// VectorField 向量字段配置（匹配vector_field节点）
type VectorField struct {
	FieldName      string     `yaml:"field_name"`            // 向量字段名（如embedding）
	Algorithm      string     `yaml:"algorithm"`             // 索引算法：HNSW/FLAT
	Dim            int        `yaml:"dim"`                   // 向量维度
	DataType       string     `yaml:"data_type"`             // 数据类型：FLOAT32/FLOAT64
	DistanceMetric string     `yaml:"distance_metric"`       // 距离计算方式：COSINE/L2/IP
	HnswParams     HnswParams `yaml:"hnsw_params,omitempty"` // HNSW算法参数（仅HNSW生效）
}

// HnswParams HNSW算法调优参数（匹配hnsw_params节点）
type HnswParams struct {
	M              int `yaml:"M"`               // 每个节点最大邻居数
	EfConstruction int `yaml:"ef_construction"` // 索引构建时的探索范围
	EfRuntime      int `yaml:"ef_runtime"`      // 查询时的探索范围
}

// ====================== 向量操作配置结构体（匹配redis_vector_operations节点）

type VectorOperations struct {
	Insert InsertConfig `yaml:"insert"` // 向量写入配置
	Search SearchConfig `yaml:"search"` // 向量查询配置
}

// InsertConfig 向量写入配置（匹配insert节点）
type InsertConfig struct {
	BatchSize    int    `yaml:"batch_size"`    // 批量写入大小
	VectorEncode string `yaml:"vector_encode"` // 向量编码方式：bytes/list
}

// SearchConfig 向量查询配置（匹配search节点）
type SearchConfig struct {
	TopK            int  `yaml:"top_k"`             // 默认返回TOP-K相似向量
	Timeout         int  `yaml:"timeout"`           // 查询超时时间（ms）
	ReturnDistance  bool `yaml:"return_distance"`   // 是否返回距离值
	ReturnRawVector bool `yaml:"return_raw_vector"` // 是否返回原始向量
}

// ====================== 向量性能调优结构体（匹配redis_vector_tuning节点） ======================

type VectorTuning struct {
	HnswMemory     HnswMemoryConfig `yaml:"hnsw_memory,omitempty"`     // HNSW内存优化
	DistanceFilter DistanceFilter   `yaml:"distance_filter,omitempty"` // 距离过滤阈值
}

// HnswMemoryConfig HNSW内存优化参数（匹配hnsw_memory节点）
type HnswMemoryConfig struct {
	MaxConnections int `yaml:"max_connections"` // 同hnsw_params.M
	GCInterval     int `yaml:"gc_interval"`     // 索引垃圾回收间隔（秒）
}

// DistanceFilter 距离过滤阈值（匹配distance_filter节点）
type DistanceFilter struct {
	CosineMax float64 `yaml:"cosine_max"` // 余弦距离最大阈值
	L2Max     float64 `yaml:"l2_max"`     // L2距离最大阈值
	IPMin     float64 `yaml:"ip_min"`     // IP内积最小阈值
}
