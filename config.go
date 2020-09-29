package workers

import (
	"time"

	"github.com/garyburd/redigo/redis"
)

type config struct {
	Pools map[string]*pool
}

type pool struct {
	Namespace string
	Pool      *redis.Pool
}

// RedisOptions redis conn options
type RedisOptions struct {
	// 关键字，用于筛选Redis客户端，作为 config.Pools 中的 key
	KeyWord   string
	Namespace string // 辅助字段
	// Redis 服务端地址，需带端口号
	Addr     string
	Password string
	Database int
	MaxIdle  int
	// 默认单位: 秒
	IdleTimeout int
}

// Config global config for worker client
var Config *config

// DefaultRedisClient -
const DefaultRedisClient = ""

func init() {
	Config = &config{Pools: make(map[string]*pool)}
}

// Configure config worker client, support multi redis server
func Configure(options ...RedisOptions) {
	for _, option := range options {
		configure(option)
	}
}

func configure(options RedisOptions) {
	namespace := ""
	if options.Namespace != "" {
		namespace = options.Namespace + ":"
	}
	if options.Addr == "" {
		panic("please tell me redis address")
	}
	if options.MaxIdle == 0 {
		options.MaxIdle = 10
	}
	if options.IdleTimeout == 0 {
		options.IdleTimeout = 15
	}
	Config.Pools[options.KeyWord] = &pool{
		namespace,
		&redis.Pool{
			MaxIdle:     options.MaxIdle,
			IdleTimeout: time.Duration(options.IdleTimeout) * time.Second,
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial("tcp", options.Addr)
				if err != nil {
					return nil, err
				}
				if options.Password != "" {
					if _, err := c.Do("AUTH", options.Password); err != nil {
						c.Close()
						return nil, err
					}
				}
				if _, err := c.Do("SELECT", options.Database); err != nil {
					c.Close()
					return nil, err
				}
				return c, err
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				_, err := c.Do("PING")
				return err
			},
		},
	}
}
