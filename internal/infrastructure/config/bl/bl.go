package config

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

type Blacklist interface {
	Add(ip string, ttl time.Duration) error
	Exists(ip string) (bool, error)
	Close() error
}

type RedisBL struct {
	client *redis.Client
	prefix string
}

func NewRedisBL() (*RedisBL, error) {
	cfg, err := LoadConfig("redis.yml")
	if err != nil {
		return nil, err
	}

	if len(cfg.Redis.SentinelAddrs) == 0 {
		return nil, fmt.Errorf("no sentinel addresses provided")
	}

	rdb := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:      cfg.Redis.MasterName,
		SentinelAddrs:   cfg.Redis.SentinelAddrs,
		Password:        cfg.Redis.Password,
		DB:              cfg.Redis.DB,
		DialTimeout:     3 * time.Second,
		ReadTimeout:     2 * time.Second,
		WriteTimeout:    2 * time.Second,
		MaxRetries:      3,
		MinRetryBackoff: 100 * time.Millisecond,
		MaxRetryBackoff: 2 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisBL{
		client: rdb,
	}, nil
}

func (b *RedisBL) Add(ip string, ttl time.Duration) error {
	ctx := context.Background()
	key := b.prefix + ip

	// 1 - есть ttl, 0 - навсегда
	return b.client.Set(ctx, key, "1", ttl).Err()
}

func (b *RedisBL) Exists(ip string) (bool, error) {
	ctx := context.Background()
	key := b.prefix + ip

	res, err := b.client.Exists(ctx, key).Result()
	if err != nil {
		return false, nil
	}
	return res == 1, nil
}

func (b *RedisBL) Close() error {
	return b.client.Close()
}
