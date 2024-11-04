package interfaces

import (
	"context"

	"github.com/redis/go-redis/v9"
)

type UniversalClient interface {
	redis.Cmdable
	AddHook(redis.Hook)
	Watch(ctx context.Context, fn func(*redis.Tx) error, keys ...string) error
	Do(ctx context.Context, args ...interface{}) *redis.Cmd
	Process(ctx context.Context, cmd redis.Cmder) error
	Subscribe(ctx context.Context, channels ...string) *redis.PubSub
	PSubscribe(ctx context.Context, channels ...string) *redis.PubSub
	SSubscribe(ctx context.Context, channels ...string) *redis.PubSub
	Close() error
	PoolStats() *redis.PoolStats
}
