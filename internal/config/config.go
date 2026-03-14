package config

import (
	"context"
	"crypto/tls"
	"time"
)

var (
	EmbeddingConfig EmbeddingModelConfig
	RedisConfig     RedisClientConfig
)

func Init(ctx context.Context) error {
	return Load(ctx)
}

type EmbeddingModelConfig struct {
	Choose    string    `yaml:"choose"`
	BatchSize int       `yaml:"batch_size"`
	KeyPrefix string    `yaml:"key_prefix"`
	Dashscope DashScope `yaml:"dashscope"`
}

type DashScope struct {
	APIKey     string `yaml:"api_key"`
	BaseURL    string `yaml:"base_url"`
	Model      string `yaml:"model"`
	Dimensions int    `yaml:"dimensions"`
}

type RedisClientConfig struct {
	Network     string `yaml:"network" default:"tcp"`
	Addr        string `yaml:"addr" default:"127.0.0.1:6379"`
	NodeAddress string `yaml:"node_address"`
	ClientName  string `yaml:"client_name"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	DB          int    `yaml:"db" default:"0"`

	MaxRetries      int           `yaml:"max_retries" default:"3"`
	MinRetryBackoff time.Duration `yaml:"min_retry_backoff" default:"8ms"`
	MaxRetryBackoff time.Duration `yaml:"max_retry_backoff" default:"512ms"`

	DialTimeout        time.Duration `yaml:"dial_timeout" default:"5s"`
	ReadTimeout        time.Duration `yaml:"read_timeout" default:"3s"`
	WriteTimeout       time.Duration `yaml:"write_timeout" default:"3s"`
	PoolTimeout        time.Duration `yaml:"pool_timeout"`
	DialerRetries      int           `yaml:"dialer_retries" default:"5"`
	DialerRetryTimeout time.Duration `yaml:"dialer_retry_timeout" default:"100ms"`

	PoolFIFO              bool          `yaml:"pool_fifo" default:"false"`
	PoolSize              int           `yaml:"pool_size"`
	MaxConcurrentDials    int           `yaml:"max_concurrent_dials"`
	MinIdleConns          int           `yaml:"min_idle_conns" default:"0"`
	MaxIdleConns          int           `yaml:"max_idle_conns" default:"0"`
	MaxActiveConns        int           `yaml:"max_active_conns" default:"0"`
	ConnMaxIdleTime       time.Duration `yaml:"conn_max_idle_time" default:"30m"`
	ConnMaxLifetime       time.Duration `yaml:"conn_max_lifetime" default:"0"`
	ConnMaxLifetimeJitter time.Duration `yaml:"conn_max_lifetime_jitter"`

	Protocol              int    `yaml:"protocol" default:"3"`
	ContextTimeoutEnabled bool   `yaml:"context_timeout_enabled"`
	ReadBufferSize        int    `yaml:"read_buffer_size" default:"32768"`
	WriteBufferSize       int    `yaml:"write_buffer_size" default:"32768"`
	DisableIdentity       bool   `yaml:"disable_identity" default:"false"`
	IdentitySuffix        string `yaml:"identity_suffix"`
	UnstableResp3         bool   `yaml:"unstable_resp3" default:"false"`
	FailingTimeoutSeconds int    `yaml:"failing_timeout_seconds" default:"15"`

	TLS TLSConfig `yaml:"tls"`
}

type TLSConfig struct {
	Enabled    bool   `yaml:"enabled" default:"false"`
	ServerName string `yaml:"server_name"`
	Insecure   bool   `yaml:"insecure" default:"false"`
}

func (c *RedisClientConfig) ToRedisOptions() interface{} {
	var tlsCfg *tls.Config
	if c.TLS.Enabled {
		tlsCfg = &tls.Config{
			ServerName:         c.TLS.ServerName,
			InsecureSkipVerify: c.TLS.Insecure,
		}
	}

	poolSize := c.PoolSize
	if poolSize <= 0 {
		poolSize = 10
	}

	poolTimeout := c.PoolTimeout
	if poolTimeout <= 0 {
		poolTimeout = c.ReadTimeout + time.Second
	}

	return struct {
		Network               string
		Addr                  string
		NodeAddress           string
		ClientName            string
		Protocol              int
		Username              string
		Password              string
		DB                    int
		MaxRetries            int
		MinRetryBackoff       time.Duration
		MaxRetryBackoff       time.Duration
		DialTimeout           time.Duration
		DialerRetries         int
		DialerRetryTimeout    time.Duration
		ReadTimeout           time.Duration
		WriteTimeout          time.Duration
		ContextTimeoutEnabled bool
		ReadBufferSize        int
		WriteBufferSize       int
		PoolFIFO              bool
		PoolSize              int
		MaxConcurrentDials    int
		PoolTimeout           time.Duration
		MinIdleConns          int
		MaxIdleConns          int
		MaxActiveConns        int
		ConnMaxIdleTime       time.Duration
		ConnMaxLifetime       time.Duration
		ConnMaxLifetimeJitter time.Duration
		TLSConfig             *tls.Config
		DisableIdentity       bool
		IdentitySuffix        string
		UnstableResp3         bool
		FailingTimeoutSeconds int
	}{
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
