package config

import (
	"context"
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

func NewRedisBL(addr, passwd string, db int) (*RedisBL, error) {
	rdb := redis.NewClient(&redis.Options{Addr: addr, Password: passwd, DB: db})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisBL{
		client: rdb,
		prefix: "bl:ip:",
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
