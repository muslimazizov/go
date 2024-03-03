package redis

import (
	"time"

	"gogogo/config"

	"github.com/gomodule/redigo/redis"
)

type Cache struct {
	Pool *redis.Pool
}

func New(cfg *config.Config) Cache {
	pool := &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", cfg.RedisURL)
		},
	}

	return Cache{Pool: pool}
}
