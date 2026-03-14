package bootstrap

import (
	"context"
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/redis/go-redis/v9"
)

var (
	redisClient *redis.Client
	redisConfig RedisConfig
)

func initRedis(ctx context.Context) {
	cfg, err := g.Cfg().Get(ctx, "redis")
	if err != nil {
		panic(err)
	}
	err = cfg.Struct(&redisConfig)
	if err != nil {
		panic(err)
	}
	redisClient = redis.NewClient(redisConfig.ToRedisOptions())
	if err = redisClient.Ping(ctx).Err(); err != nil {
		panic(err)
	}
	g.Log().Info(ctx, "Redis初始化成功")
	has, err := hasRedisStack(ctx)
	if err != nil {
		panic(err)
	}
	if !has {
		panic("没有Redis Stack模块")
	}
	g.Log().Info(ctx, "加载Redis Stack模块成功")
}

func GetRedisClient() *redis.Client {
	return redisClient
}

func CloseRedisClient() error {
	if redisClient != nil {
		return redisClient.Close()
	}
	return nil
}

type RedisConfig struct {
	// 基础连接配置
	Network     string `yaml:"network" default:"tcp"`         // 网络类型：tcp/unix
	Addr        string `yaml:"addr" default:"127.0.0.1:6379"` // 地址：host:port
	NodeAddress string `yaml:"node_address"`                  // 节点地址（集群模式用）
	ClientName  string `yaml:"client_name"`                   // 客户端名称
	Username    string `yaml:"username"`                      // 用户名（Redis6+ ACL）
	Password    string `yaml:"password"`                      // 密码
	DB          int    `yaml:"db" default:"0"`                // 数据库索引

	// 重试配置
	MaxRetries      int           `yaml:"max_retries" default:"3"`           // 最大重试次数
	MinRetryBackoff time.Duration `yaml:"min_retry_backoff" default:"8ms"`   // 最小重试间隔
	MaxRetryBackoff time.Duration `yaml:"max_retry_backoff" default:"512ms"` // 最大重试间隔

	// 超时配置
	DialTimeout        time.Duration `yaml:"dial_timeout" default:"5s"`            // 连接超时
	ReadTimeout        time.Duration `yaml:"read_timeout" default:"3s"`            // 读取超时
	WriteTimeout       time.Duration `yaml:"write_timeout" default:"3s"`           // 写入超时
	PoolTimeout        time.Duration `yaml:"pool_timeout"`                         // 连接池等待超时（默认=ReadTimeout+1s）
	DialerRetries      int           `yaml:"dialer_retries" default:"5"`           // 拨号重试次数
	DialerRetryTimeout time.Duration `yaml:"dialer_retry_timeout" default:"100ms"` // 拨号重试间隔

	// 连接池配置
	PoolFIFO              bool          `yaml:"pool_fifo" default:"false"`        // 连接池类型：FIFO/LIFO
	PoolSize              int           `yaml:"pool_size"`                        // 连接池基础大小（默认=10*CPU核心数）
	MaxConcurrentDials    int           `yaml:"max_concurrent_dials"`             // 最大并发拨号数
	MinIdleConns          int           `yaml:"min_idle_conns" default:"0"`       // 最小空闲连接数
	MaxIdleConns          int           `yaml:"max_idle_conns" default:"0"`       // 最大空闲连接数
	MaxActiveConns        int           `yaml:"max_active_conns" default:"0"`     // 最大活跃连接数（0=无限制）
	ConnMaxIdleTime       time.Duration `yaml:"conn_max_idle_time" default:"30m"` // 连接最大空闲时间
	ConnMaxLifetime       time.Duration `yaml:"conn_max_lifetime" default:"0"`    // 连接最大存活时间
	ConnMaxLifetimeJitter time.Duration `yaml:"conn_max_lifetime_jitter"`         // 存活时间抖动值

	// 高级配置
	Protocol              int    `yaml:"protocol" default:"3"`                 // RESP 协议版本：2/3
	ContextTimeoutEnabled bool   `yaml:"context_timeout_enabled"`              // 是否尊重上下文超时
	ReadBufferSize        int    `yaml:"read_buffer_size" default:"32768"`     // 读缓冲区大小
	WriteBufferSize       int    `yaml:"write_buffer_size" default:"32768"`    // 写缓冲区大小
	DisableIdentity       bool   `yaml:"disable_identity" default:"false"`     // 禁用 CLIENT SETINFO
	IdentitySuffix        string `yaml:"identity_suffix"`                      // 客户端名称后缀
	UnstableResp3         bool   `yaml:"unstable_resp3" default:"false"`       // 启用 RESP3 不稳定模式
	FailingTimeoutSeconds int    `yaml:"failing_timeout_seconds" default:"15"` // 集群节点失败超时

	TLS TLSConfig `yaml:"tls"`
}

// TLSConfig TLS 配置子结构体
type TLSConfig struct {
	Enabled    bool   `yaml:"enabled" default:"false"`  // 是否启用 TLS
	ServerName string `yaml:"server_name"`              // 服务端名称（用于证书验证）
	Insecure   bool   `yaml:"insecure" default:"false"` // 跳过证书验证（测试环境用）
}

func (c *RedisConfig) ToRedisOptions() *redis.Options {
	// 构建 TLS 配置
	var tlsCfg *tls.Config
	if c.TLS.Enabled {
		tlsCfg = &tls.Config{
			ServerName:         c.TLS.ServerName,
			InsecureSkipVerify: c.TLS.Insecure, // 生产环境务必设为 false
		}
	}

	// 处理 PoolSize 默认值（10 * CPU核心数）
	poolSize := c.PoolSize
	if poolSize <= 0 {
		poolSize = 10 * 1 // 简化版，生产环境替换为：runtime.GOMAXPROCS(0)
	}

	// 处理 PoolTimeout 默认值（ReadTimeout + 1s）
	poolTimeout := c.PoolTimeout
	if poolTimeout <= 0 {
		poolTimeout = c.ReadTimeout + time.Second
	}

	// 转换为原生 Options
	return &redis.Options{
		Network:               c.Network,
		Addr:                  c.Addr,
		NodeAddress:           c.NodeAddress,
		ClientName:            c.ClientName,
		Protocol:              c.Protocol,
		Username:              c.Username,
		Password:              c.Password,
		DB:                    c.DB,
		MaxRetries:            c.MaxRetries,
		MinRetryBackoff:       c.MinRetryBackoff,
		MaxRetryBackoff:       c.MaxRetryBackoff,
		DialTimeout:           c.DialTimeout,
		DialerRetries:         c.DialerRetries,
		DialerRetryTimeout:    c.DialerRetryTimeout,
		ReadTimeout:           c.ReadTimeout,
		WriteTimeout:          c.WriteTimeout,
		ContextTimeoutEnabled: c.ContextTimeoutEnabled,
		ReadBufferSize:        c.ReadBufferSize,
		WriteBufferSize:       c.WriteBufferSize,
		PoolFIFO:              c.PoolFIFO,
		PoolSize:              poolSize,
		MaxConcurrentDials:    c.MaxConcurrentDials,
		PoolTimeout:           poolTimeout,
		MinIdleConns:          c.MinIdleConns,
		MaxIdleConns:          c.MaxIdleConns,
		MaxActiveConns:        c.MaxActiveConns,
		ConnMaxIdleTime:       c.ConnMaxIdleTime,
		ConnMaxLifetime:       c.ConnMaxLifetime,
		ConnMaxLifetimeJitter: c.ConnMaxLifetimeJitter,
		TLSConfig:             tlsCfg,
		DisableIdentity:       c.DisableIdentity,
		IdentitySuffix:        c.IdentitySuffix,
		UnstableResp3:         c.UnstableResp3,
		FailingTimeoutSeconds: c.FailingTimeoutSeconds,
	}
}

func hasRedisStack(ctx context.Context) (bool, error) {
	//执行 MODULE LIST 命令
	modules, err := redisClient.Do(ctx, "MODULE", "LIST").Result()
	if err != nil {
		return false, fmt.Errorf("执行 MODULE LIST 失败：%w", err)
	}

	//后检查是否包含 "redis-stack-server"
	modulesStr := fmt.Sprint(modules)
	if strings.Contains(modulesStr, "redis-stack-server") {
		return true, nil
	}

	// 未匹配到关键词
	return false, fmt.Errorf("没有Redis Stack模块")
}
