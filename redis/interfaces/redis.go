/*
 * Copyright (c) 2016 TFG Co <backend@tfgco.com>
 * Author: TFG Co <backend@tfgco.com>
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of
 * this software and associated documentation files (the "Software"), to deal in
 * the Software without restriction, including without limitation the rights to
 * use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
 * the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
 * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 */

package interfaces

import (
	"context"
	"time"

	"github.com/go-redis/redis"
)

// TraceWrapper represents the contract for a redis trace wrapper
type TraceWrapper interface {
	WithContext(ctx context.Context, c RedisClient) RedisClient
}

//RedisClient represents the contract for a redis client
type RedisClient interface {
	BLPop(timeout time.Duration, keys ...string) *redis.StringSliceCmd
	Close() error
	Context() context.Context
	Scan(cursor uint64, match string, count int64) *redis.ScanCmd
	Del(keys ...string) *redis.IntCmd
	Eval(script string, keys []string, args ...interface{}) *redis.Cmd
	EvalSha(sha1 string, keys []string, args ...interface{}) *redis.Cmd
	Exists(keys ...string) *redis.IntCmd
	Get(key string) *redis.StringCmd
	HDel(key string, fields ...string) *redis.IntCmd
	HGet(key, field string) *redis.StringCmd
	HGetAll(string) *redis.StringStringMapCmd
	HMGet(string, ...string) *redis.SliceCmd
	HMSet(string, map[string]interface{}) *redis.StatusCmd
	HSet(key, field string, value interface{}) *redis.BoolCmd
	HScan(key string, cursor uint64, match string, count int64) *redis.ScanCmd
	MGet(keys ...string) *redis.SliceCmd
	LRange(key string, start, stop int64) *redis.StringSliceCmd
	Ping() *redis.StatusCmd
	RPopLPush(source string, destination string) *redis.StringCmd
	RPush(key string, values ...interface{}) *redis.IntCmd
	SAdd(key string, members ...interface{}) *redis.IntCmd
	SCard(key string) *redis.IntCmd
	SIsMember(key string, member interface{}) *redis.BoolCmd
	SMembers(key string) *redis.StringSliceCmd
	SPopN(key string, count int64) *redis.StringSliceCmd
	SRem(key string, members ...interface{}) *redis.IntCmd
	ScriptExists(scripts ...string) *redis.BoolSliceCmd
	ScriptLoad(script string) *redis.StringCmd
	Set(key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	SetNX(key string, value interface{}, expiration time.Duration) *redis.BoolCmd
	TTL(key string) *redis.DurationCmd
	TxPipeline() redis.Pipeliner
	WithContext(context.Context) *redis.Client
	ZAdd(key string, members ...redis.Z) *redis.IntCmd
	ZCard(key string) *redis.IntCmd
	ZRangeByScore(key string, opt redis.ZRangeBy) *redis.StringSliceCmd
	ZRangeByScoreWithScores(key string, opt redis.ZRangeBy) *redis.ZSliceCmd
	ZRangeWithScores(key string, start, stop int64) *redis.ZSliceCmd
	ZRank(key, member string) *redis.IntCmd
	ZRem(key string, members ...interface{}) *redis.IntCmd
	ZRevRangeByScore(key string, opt redis.ZRangeBy) *redis.StringSliceCmd
	ZRevRangeByScoreWithScores(key string, opt redis.ZRangeBy) *redis.ZSliceCmd
	ZRevRangeWithScores(key string, start, stop int64) *redis.ZSliceCmd
	ZRevRank(key, member string) *redis.IntCmd
	ZScore(key, member string) *redis.FloatCmd
}
