package instances

import (
	"context"
	"time"

	"github.com/go-redis/redis"
	"github.com/topfreegames/extensions/v9/redis/interfaces/experimental"
)

type RedisClusterClientInstance struct {
	instance *redis.ClusterClient
}

func NewRedisClusterClientInstance(instance *redis.ClusterClient) experimental.RedisInstance {
	return &RedisClusterClientInstance{
		instance: instance,
	}
}

func (r RedisClusterClientInstance) WithContext(ctx context.Context) experimental.RedisInstance {
	newInstance := r.instance.WithContext(ctx)
	return NewRedisClusterClientInstance(newInstance)
}

func (r RedisClusterClientInstance) InstanceName() string {
	return "cluster"
}

func (r RedisClusterClientInstance) BLPop(timeout time.Duration, keys ...string) *redis.StringSliceCmd {
	return r.instance.BLPop(timeout, keys...)
}

func (r RedisClusterClientInstance) Close() error {
	return r.instance.Close()
}

func (r RedisClusterClientInstance) Context() context.Context {
	return r.instance.Context()
}

func (r RedisClusterClientInstance) Del(keys ...string) *redis.IntCmd {
	return r.instance.Del(keys...)
}

func (r RedisClusterClientInstance) Eval(script string, keys []string, args ...interface{}) *redis.Cmd {
	return r.instance.Eval(script, keys, args...)
}

func (r RedisClusterClientInstance) EvalSha(sha1 string, keys []string, args ...interface{}) *redis.Cmd {
	return r.instance.EvalSha(sha1, keys, args...)
}

func (r RedisClusterClientInstance) Exists(keys ...string) *redis.IntCmd {
	return r.instance.Exists(keys...)
}

func (r RedisClusterClientInstance) Get(key string) *redis.StringCmd {
	return r.instance.Get(key)
}

func (r RedisClusterClientInstance) HDel(key string, fields ...string) *redis.IntCmd {
	return r.instance.HDel(key, fields...)
}

func (r RedisClusterClientInstance) HGet(key, field string) *redis.StringCmd {
	return r.instance.HGet(key, field)
}

func (r RedisClusterClientInstance) HGetAll(s string) *redis.StringStringMapCmd {
	return r.instance.HGetAll(s)
}

func (r RedisClusterClientInstance) HMGet(s string, s2 ...string) *redis.SliceCmd {
	return r.instance.HMGet(s, s2...)
}

func (r RedisClusterClientInstance) HMSet(s string, m map[string]interface{}) *redis.StatusCmd {
	return r.instance.HMSet(s, m)
}

func (r RedisClusterClientInstance) HSet(key, field string, value interface{}) *redis.BoolCmd {
	return r.instance.HSet(key, field, value)
}

func (r RedisClusterClientInstance) MGet(keys ...string) *redis.SliceCmd {
	return r.instance.MGet(keys...)
}

func (r RedisClusterClientInstance) LRange(key string, start, stop int64) *redis.StringSliceCmd {
	return r.instance.LRange(key, start, stop)
}

func (r RedisClusterClientInstance) Ping() *redis.StatusCmd {
	return r.instance.Ping()
}

func (r RedisClusterClientInstance) RPopLPush(source string, destination string) *redis.StringCmd {
	return r.instance.RPopLPush(source, destination)
}

func (r RedisClusterClientInstance) RPush(key string, values ...interface{}) *redis.IntCmd {
	return r.instance.RPush(key, values...)
}

func (r RedisClusterClientInstance) SAdd(key string, members ...interface{}) *redis.IntCmd {
	return r.instance.SAdd(key, members...)
}

func (r RedisClusterClientInstance) SCard(key string) *redis.IntCmd {
	return r.instance.SCard(key)
}

func (r RedisClusterClientInstance) SIsMember(key string, member interface{}) *redis.BoolCmd {
	return r.instance.SIsMember(key, member)
}

func (r RedisClusterClientInstance) SMembers(key string) *redis.StringSliceCmd {
	return r.instance.SMembers(key)
}

func (r RedisClusterClientInstance) SPopN(key string, count int64) *redis.StringSliceCmd {
	return r.instance.SPopN(key, count)
}

func (r RedisClusterClientInstance) SRem(key string, members ...interface{}) *redis.IntCmd {
	return r.instance.SRem(key, members...)
}

func (r RedisClusterClientInstance) ScriptExists(scripts ...string) *redis.BoolSliceCmd {
	return r.instance.ScriptExists(scripts...)
}

func (r RedisClusterClientInstance) ScriptLoad(script string) *redis.StringCmd {
	return r.instance.ScriptLoad(script)
}

func (r RedisClusterClientInstance) Set(key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return r.instance.Set(key, value, expiration)
}

func (r RedisClusterClientInstance) SetNX(key string, value interface{}, expiration time.Duration) *redis.BoolCmd {
	return r.instance.SetNX(key, value, expiration)
}

func (r RedisClusterClientInstance) TTL(key string) *redis.DurationCmd {
	return r.instance.TTL(key)
}

func (r RedisClusterClientInstance) TxPipeline() redis.Pipeliner {
	return r.instance.TxPipeline()
}

func (r RedisClusterClientInstance) ZAdd(key string, members ...redis.Z) *redis.IntCmd {
	return r.instance.ZAdd(key, members...)
}

func (r RedisClusterClientInstance) ZCard(key string) *redis.IntCmd {
	return r.instance.ZCard(key)
}

func (r RedisClusterClientInstance) ZRangeByScore(key string, opt redis.ZRangeBy) *redis.StringSliceCmd {
	return r.instance.ZRangeByScore(key, opt)
}

func (r RedisClusterClientInstance) ZRangeByScoreWithScores(key string, opt redis.ZRangeBy) *redis.ZSliceCmd {
	return r.instance.ZRangeByScoreWithScores(key, opt)
}

func (r RedisClusterClientInstance) ZRangeWithScores(key string, start, stop int64) *redis.ZSliceCmd {
	return r.instance.ZRangeWithScores(key, start, stop)
}

func (r RedisClusterClientInstance) ZRank(key, member string) *redis.IntCmd {
	return r.instance.ZRank(key, member)
}

func (r RedisClusterClientInstance) ZRem(key string, members ...interface{}) *redis.IntCmd {
	return r.instance.ZRem(key, members...)
}

func (r RedisClusterClientInstance) ZRevRangeByScore(key string, opt redis.ZRangeBy) *redis.StringSliceCmd {
	return r.instance.ZRevRangeByScore(key, opt)
}

func (r RedisClusterClientInstance) ZRevRangeByScoreWithScores(key string, opt redis.ZRangeBy) *redis.ZSliceCmd {
	return r.instance.ZRevRangeByScoreWithScores(key, opt)
}

func (r RedisClusterClientInstance) ZRevRangeWithScores(key string, start, stop int64) *redis.ZSliceCmd {
	return r.instance.ZRevRangeWithScores(key, start, stop)
}

func (r RedisClusterClientInstance) ZRevRank(key, member string) *redis.IntCmd {
	return r.instance.ZRevRank(key, member)
}

func (r RedisClusterClientInstance) ZScore(key, member string) *redis.FloatCmd {
	return r.instance.ZScore(key, member)
}

func (r RedisClusterClientInstance) WrapProcess(middleware func(old func(cmd redis.Cmder) error) func(cmd redis.Cmder) error) {
	r.instance.WrapProcess(middleware)
}

func (r RedisClusterClientInstance) WrapProcessPipeline(pipe func(old func(cmds []redis.Cmder) error) func(cmds []redis.Cmder) error) {
	r.instance.WrapProcessPipeline(pipe)
}
