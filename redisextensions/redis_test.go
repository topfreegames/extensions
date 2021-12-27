package redisextensions_test

/*
 * Copyright (c) 2021 TFG Co <backend@tfgco.com>
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
	"time"

	"github.com/go-redis/redis/v8"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
	"github.com/topfreegames/extensions/v9/redisextensions"
)

var _ = Describe("Redis", func() {
	var client *redis.Client
	ctx := context.TODO()

	BeforeEach(func() {
		var err error
		config := viper.New()
		config.SetConfigFile("../config/test.yaml")
		Expect(config.ReadInConfig()).NotTo(HaveOccurred())
		client, err = redisextensions.NewClient(ctx, "extensions.redis", config)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("[Integration]", func() {
		It("set", func() {
			key := "extensions:testkey"
			expectedValue := "value"
			_, err := client.Set(ctx, key, expectedValue, time.Duration(0)).Result()
			Expect(err).NotTo(HaveOccurred())
			value, err := client.Get(ctx, key).Result()
			Expect(value).To(Equal(expectedValue))
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

func TestRedisExtensions(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Redis v8 extensions")
}
