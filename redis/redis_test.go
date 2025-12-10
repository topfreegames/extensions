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

package redis

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	"github.com/topfreegames/extensions/v9/redis/mocks"
	"go.uber.org/mock/gomock"
)

var _ = Describe("Redis Extension", func() {
	var config *viper.Viper
	var mockClient *mocks.MockRedisClient
	var mockCtrl *gomock.Controller

	BeforeEach(func() {
		config = viper.New()
		config.SetConfigFile("../config/test.yaml")
		mockCtrl = gomock.NewController(GinkgoT())
		mockClient = mocks.NewMockRedisClient(mockCtrl)
		Expect(config.ReadInConfig()).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		mockCtrl.Finish()
	})

	Describe("[Unit]", func() {
		Describe("Connect", func() {
			It("Should use config to load connection details", func() {
				mockClient.EXPECT().Ping(gomock.Any())
				client, err := NewClient("extensions.redis", config, mockClient)
				Expect(err).NotTo(HaveOccurred())
				Expect(client.Options).NotTo(BeNil())
			})
		})

		Describe("IsConnected", func() {
			It("should verify that db is connected", func() {
				mockClient.EXPECT().Ping(gomock.Any())
				client, err := NewClient("extensions.redis", config, mockClient)
				Expect(err).NotTo(HaveOccurred())
				mockClient.EXPECT().Ping(gomock.Any())
				Expect(client.IsConnected()).To(BeTrue())
			})

			It("should not be connected if something other than 'PONG' returned", func() {
				mockClient.EXPECT().Ping(gomock.Any())
				client, err := NewClient("extensions.redis", config, mockClient)
				Expect(err).NotTo(HaveOccurred())
				cmd := redis.NewStatusCmd(context.Background())
				cmd.SetVal("OK")
				mockClient.EXPECT().Ping(gomock.Any()).Return(cmd)
				Expect(client.IsConnected()).To(BeFalse())
			})

			It("should not be connected if redis error", func() {
				mockClient.EXPECT().Ping(gomock.Any())
				client, err := NewClient("extensions.redis", config, mockClient)
				Expect(err).NotTo(HaveOccurred())
				cmd := redis.NewStatusCmd(context.Background())
				cmd.SetErr(fmt.Errorf("redis error"))
				mockClient.EXPECT().Ping(gomock.Any()).Return(cmd)
				Expect(client.IsConnected()).To(BeFalse())
			})
		})

		Describe("Close", func() {
			It("should close if no errors", func() {
				mockClient.EXPECT().Ping(gomock.Any())
				client, err := NewClient("extensions.redis", config, mockClient)
				Expect(err).NotTo(HaveOccurred())
				mockClient.EXPECT().Close()
				err = client.Close()
				Expect(err).NotTo(HaveOccurred())
			})

			It("should not close if errors", func() {
				mockClient.EXPECT().Ping(gomock.Any())
				client, err := NewClient("extensions.redis", config, mockClient)
				Expect(err).NotTo(HaveOccurred())
				mockClient.EXPECT().Close().Return(fmt.Errorf("redis error"))
				err = client.Close()
				Expect(err).To(HaveOccurred())
			})
		})

		Describe("WaitForConnection", func() {
			It("should wait for connection", func() {
				mockClient.EXPECT().Ping(gomock.Any())
				client, err := NewClient("extensions.redis", config, mockClient)
				Expect(err).NotTo(HaveOccurred())

				mockClient.EXPECT().Ping(gomock.Any())
				err = client.WaitForConnection(1)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Describe("Cleanup", func() {
			It("should close connection", func() {
				mockClient.EXPECT().Ping(gomock.Any())
				client, err := NewClient("extensions.redis", config, mockClient)
				Expect(err).NotTo(HaveOccurred())
				mockClient.EXPECT().Close()
				err = client.Cleanup()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		// EnterCriticalSection requires a concrete *redis.Client for redislock
		// This functionality is tested in the Integration tests below
		XDescribe("EnterCriticalSection", func() {
			It("should lock in redis", func() {
				mockClient.EXPECT().Ping(gomock.Any())
				client, err := NewClient("extensions.redis", config, mockClient)
				Expect(err).NotTo(HaveOccurred())
				t := 10 * time.Millisecond
				mockClient.EXPECT().SetNX(gomock.Any(), "mock", gomock.Any(), 10*time.Millisecond).Return(redis.NewBoolCmd(context.Background()))
				client.EnterCriticalSection(mockClient, "mock", t, t, t)
				mockClient.EXPECT().Close()
				err = client.Cleanup()
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	XDescribe("[Integration]", func() {
		Describe("Creating new client", func() {
			It("should return connected client", func() {
				client, err := NewClient("extensions.redis", config)
				Expect(err).NotTo(HaveOccurred())
				defer client.Close()
				Expect(client).NotTo(BeNil())
				client.WaitForConnection(10)
				Expect(client.IsConnected()).To(BeTrue())
			})
		})

		Describe("Test locking", func() {
			It("should get a lock only once", func() {
				client, err := NewClient("extensions.redis", config)
				Expect(err).NotTo(HaveOccurred())
				defer client.Close()
				Expect(client).NotTo(BeNil())
				client.WaitForConnection(10)
				Expect(client.IsConnected()).To(BeTrue())
				t := 100 * time.Millisecond
				lock, err := client.EnterCriticalSection(client.Client, "lock", t, t, t)
				Expect(err).NotTo(HaveOccurred())
				Expect(lock).NotTo(BeNil())
				lock2, err := client.EnterCriticalSection(client.Client, "lock", t, 0, 0)
				Expect(lock2).To(BeNil())
				Expect(err).To(HaveOccurred())
			})

			It("should get a lock, unlock and get it again", func() {
				client, err := NewClient("extensions.redis", config)
				Expect(err).NotTo(HaveOccurred())
				defer client.Close()
				Expect(client).NotTo(BeNil())
				client.WaitForConnection(10)
				Expect(client.IsConnected()).To(BeTrue())
				t := 100 * time.Millisecond
				lock, err := client.EnterCriticalSection(client.Client, "lock2", t, t, t)
				Expect(err).NotTo(HaveOccurred())
				Expect(lock).NotTo(BeNil())
				err = client.LeaveCriticalSection(lock)
				Expect(err).NotTo(HaveOccurred())
				lock2, err := client.EnterCriticalSection(client.Client, "lock2", t, 0, 0)
				Expect(err).NotTo(HaveOccurred())
				Expect(lock2).NotTo(BeNil())
			})
		})
	})
})
