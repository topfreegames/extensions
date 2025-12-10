package redis

/*
 * Copyright (c) 2019 TFG Co <backend@tfgco.com>
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

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tracing Redis", func() {
	var client *redis.Client
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
		client = redis.NewClient(&redis.Options{})
	})

	Describe("[Unit]", func() {
		It("del command", func() {
			testHook := &testRedisHook{}
			client.AddHook(testHook)
			client.Del(ctx, "123")
			Expect(testHook.lastCmd).Should(Equal("del"))
		})
		It("ZRevRange command", func() {
			testHook := &testRedisHook{}
			client.AddHook(testHook)
			client.ZRevRange(ctx, "123", 1, 2)
			Expect(testHook.lastCmd).Should(Equal("zrevrange"))
		})
		It("set command", func() {
			testHook := &testRedisHook{}
			client.AddHook(testHook)
			client.Set(ctx, "AAA", "BBB", 0)
			Expect(testHook.lastCmd).Should(Equal("set"))
		})
		It("exists with prefix", func() {
			testHook := &testRedisHook{}
			client.AddHook(testHook)
			client.Set(ctx, "topid:AAA", "BBB", 0)
			Expect(testHook.lastCmd).Should(Equal("set"))
		})
	})
})

type testRedisHook struct {
	lastCmd string
}

func (h *testRedisHook) DialHook(next redis.DialHook) redis.DialHook {
	return next
}

func (h *testRedisHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		h.lastCmd = cmd.Name()
		return next(ctx, cmd)
	}
}

func (h *testRedisHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return next
}

func TestTracingRedis(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tracing Redis")
}
