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
	"testing"

	"github.com/go-redis/redis"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Tracing Redis", func() {
	var client *redis.Client

	BeforeEach(func() {
		client = redis.NewClient(&redis.Options{})
	})

	Describe("[Unit]", func() {
		It("del command", func() {
			client.WrapProcess(func(old func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
				return func(cmd redis.Cmder) error {
					Expect(parseLong(cmd)).Should(Equal("del 123: 0"))
					Expect(parseShort(cmd)).Should(Equal("del"))
					return nil
				}
			})
			client.Del("123")
		})
		It("ZRevRange command", func() {
			client.WrapProcess(func(old func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
				return func(cmd redis.Cmder) error {
					Expect(parseLong(cmd)).Should(Equal("zrevrange 123 1 2: []"))
					Expect(parseShort(cmd)).Should(Equal("zrevrange"))
					return nil
				}
			})
			client.ZRevRange("123", 1, 2)
		})
		It("set command", func() {
			client.WrapProcess(func(old func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
				return func(cmd redis.Cmder) error {
					Expect(parseLong(cmd)).Should(Equal("set AAA BBB: "))
					Expect(parseShort(cmd)).Should(Equal("set"))
					return nil
				}
			})
			client.Set("AAA", "BBB", 0)
		})
		It("exists with prefix", func() {
			client.WrapProcess(func(old func(cmd redis.Cmder) error) func(cmd redis.Cmder) error {
				return func(cmd redis.Cmder) error {
					Expect(parseLong(cmd)).Should(Equal("set topid:AAA BBB: "))
					Expect(parseShort(cmd)).Should(Equal("set"))
					return nil
				}
			})
			client.Set("topid:AAA", "BBB", 0)
		})
	})
})

func TestTracingRedis(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Tracing Redis")
}
