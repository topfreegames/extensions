package instances

import (
	"context"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/topfreegames/extensions/v9/redis/interfaces/experimental"
)

type RedisClientInstance struct {
	instance *redis.Client
}

func NewRedisClientInstance(instance *redis.Client) experimental.RedisInstance {
	return &RedisClientInstance{
		instance: instance,
	}
}

func (r RedisClientInstance) WithContext(ctx context.Context) experimental.RedisInstance {
	newInstance := r.instance.WithContext(ctx)
	return NewRedisClientInstance(newInstance)
}

func (r RedisClientInstance) InstanceName() string {
	return strconv.Itoa(r.instance.Options().DB)
}

func (r RedisClientInstance) BLPop(timeout time.Duration, keys ...string) *redis.StringSliceCmd {
	return r.instance.BLPop(timeout, keys...)
}

func (r RedisClientInstance) Close() error {
	return r.instance.Close()
}

func (r RedisClientInstance) Context() context.Context {
	return r.instance.Context()
}

func (r RedisClientInstance) Del(keys ...string) *redis.IntCmd {
	return r.instance.Del(keys...)
}

func (r RedisClientInstance) Eval(script string, keys []string, args ...interface{}) *redis.Cmd {
	return r.instance.Eval(script, keys, args...)
}

func (r RedisClientInstance) EvalSha(sha1 string, keys []string, args ...interface{}) *redis.Cmd {
	return r.instance.EvalSha(sha1, keys, args...)
}

func (r RedisClientInstance) Exists(keys ...string) *redis.IntCmd {
	return r.instance.Exists(keys...)
}

func (r RedisClientInstance) Get(key string) *redis.StringCmd {
	return r.instance.Get(key)
}

func (r RedisClientInstance) HDel(key string, fields ...string) *redis.IntCmd {
	return r.instance.HDel(key, fields...)
}

func (r RedisClientInstance) HGet(key, field string) *redis.StringCmd {
	return r.instance.HGet(key, field)
}

func (r RedisClientInstance) HGetAll(s string) *redis.StringStringMapCmd {
	return r.instance.HGetAll(s)
}

func (r RedisClientInstance) HMGet(s string, s2 ...string) *redis.SliceCmd {
	return r.instance.HMGet(s, s2...)
}

func (r RedisClientInstance) HMSet(s string, m map[string]interface{}) *redis.StatusCmd {
	return r.instance.HMSet(s, m)
}

func (r RedisClientInstance) HSet(key, field string, value interface{}) *redis.BoolCmd {
	return r.instance.HSet(key, field, value)
}

func (r RedisClientInstance) MGet(keys ...string) *redis.SliceCmd {
	return r.instance.MGet(keys...)
}

func (r RedisClientInstance) LRange(key string, start, stop int64) *redis.StringSliceCmd {
	return r.instance.LRange(key, start, stop)
}

func (r RedisClientInstance) Ping() *redis.StatusCmd {
	return r.instance.Ping()
}

func (r RedisClientInstance) RPopLPush(source string, destination string) *redis.StringCmd {
	return r.instance.RPopLPush(source, destination)
}

func (r RedisClientInstance) RPush(key string, values ...interface{}) *redis.IntCmd {
	return r.instance.RPush(key, values...)
}

func (r RedisClientInstance) SAdd(key string, members ...interface{}) *redis.IntCmd {
	return r.instance.SAdd(key, members...)
}

func (r RedisClientInstance) SCard(key string) *redis.IntCmd {
	return r.instance.SCard(key)
}

func (r RedisClientInstance) SIsMember(key string, member interface{}) *redis.BoolCmd {
	return r.instance.SIsMember(key, member)
}

func (r RedisClientInstance) SMembers(key string) *redis.StringSliceCmd {
	return r.instance.SMembers(key)
}

func (r RedisClientInstance) SPopN(key string, count int64) *redis.StringSliceCmd {
	return r.instance.SPopN(key, count)
}

func (r RedisClientInstance) SRem(key string, members ...interface{}) *redis.IntCmd {
	return r.instance.SRem(key, members...)
}

func (r RedisClientInstance) ScriptExists(scripts ...string) *redis.BoolSliceCmd {
	return r.instance.ScriptExists(scripts...)
}

func (r RedisClientInstance) ScriptLoad(script string) *redis.StringCmd {
	return r.instance.ScriptLoad(script)
}

func (r RedisClientInstance) Set(key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return r.instance.Set(key, value, expiration)
}

func (r RedisClientInstance) SetNX(key string, value interface{}, expiration time.Duration) *redis.BoolCmd {
	return r.instance.SetNX(key, value, expiration)
}

func (r RedisClientInstance) TTL(key string) *redis.DurationCmd {
	return r.instance.TTL(key)
}

func (r RedisClientInstance) TxPipeline() redis.Pipeliner {
	return r.instance.TxPipeline()
}

func (r RedisClientInstance) ZAdd(key string, members ...redis.Z) *redis.IntCmd {
	return r.instance.ZAdd(key, members...)
}

func (r RedisClientInstance) ZCard(key string) *redis.IntCmd {
	return r.instance.ZCard(key)
}

func (r RedisClientInstance) ZRangeByScore(key string, opt redis.ZRangeBy) *redis.StringSliceCmd {
	return r.instance.ZRangeByScore(key, opt)
}

func (r RedisClientInstance) ZRangeByScoreWithScores(key string, opt redis.ZRangeBy) *redis.ZSliceCmd {
	return r.instance.ZRangeByScoreWithScores(key, opt)
}

func (r RedisClientInstance) ZRangeWithScores(key string, start, stop int64) *redis.ZSliceCmd {
	return r.instance.ZRangeWithScores(key, start, stop)
}

func (r RedisClientInstance) ZRank(key, member string) *redis.IntCmd {
	return r.instance.ZRank(key, member)
}

func (r RedisClientInstance) ZRem(key string, members ...interface{}) *redis.IntCmd {
	return r.instance.ZRem(key, members...)
}

func (r RedisClientInstance) ZRevRangeByScore(key string, opt redis.ZRangeBy) *redis.StringSliceCmd {
	return r.instance.ZRevRangeByScore(key, opt)
}

func (r RedisClientInstance) ZRevRangeByScoreWithScores(key string, opt redis.ZRangeBy) *redis.ZSliceCmd {
	return r.instance.ZRevRangeByScoreWithScores(key, opt)
}

func (r RedisClientInstance) ZRevRangeWithScores(key string, start, stop int64) *redis.ZSliceCmd {
	return r.instance.ZRevRangeWithScores(key, start, stop)
}

func (r RedisClientInstance) ZRevRank(key, member string) *redis.IntCmd {
	return r.instance.ZRevRank(key, member)
}

func (r RedisClientInstance) ZScore(key, member string) *redis.FloatCmd {
	return r.instance.ZScore(key, member)
}

func (r RedisClientInstance) WrapProcess(middleware func(old func(cmd redis.Cmder) error) func(cmd redis.Cmder) error) {
	r.instance.WrapProcess(middleware)
}

func (r RedisClientInstance) WrapProcessPipeline(pipe func(old func(cmds []redis.Cmder) error) func(cmds []redis.Cmder) error) {
	r.instance.WrapProcessPipeline(pipe)
}
