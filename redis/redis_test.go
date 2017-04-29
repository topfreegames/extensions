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
	"time"

	"github.com/go-redis/redis"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
	"github.com/topfreegames/extensions/redis/mocks"
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
				mockClient.EXPECT().Ping()
				client, err := NewClient("extensions.redis", config, mockClient)
				Expect(err).NotTo(HaveOccurred())
				Expect(client.Options).NotTo(BeNil())
			})
		})

		Describe("IsConnected", func() {
			It("should verify that db is connected", func() {
				mockClient.EXPECT().Ping()
				client, err := NewClient("extensions.redis", config, mockClient)
				Expect(err).NotTo(HaveOccurred())
				mockClient.EXPECT().Ping()
				Expect(client.IsConnected()).To(BeTrue())
			})

			It("should not be connected if something other than 'PONG' returned", func() {
				mockClient.EXPECT().Ping()
				client, err := NewClient("extensions.redis", config, mockClient)
				Expect(err).NotTo(HaveOccurred())
				mockClient.EXPECT().Ping().Return(&redis.StatusCmd{})
				Expect(client.IsConnected()).To(BeFalse())
			})
		})

		Describe("Close", func() {
			It("should close if no errors", func() {
				mockClient.EXPECT().Ping()
				client, err := NewClient("extensions.redis", config, mockClient)
				Expect(err).NotTo(HaveOccurred())
				mockClient.EXPECT().Close()
				err = client.Close()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Describe("WaitForConnection", func() {
			It("should wait for connection", func() {
				mockClient.EXPECT().Ping()
				client, err := NewClient("extensions.redis", config, mockClient)
				Expect(err).NotTo(HaveOccurred())

				mockClient.EXPECT().Ping()
				err = client.WaitForConnection(1)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Describe("Cleanup", func() {
			It("should close connection", func() {
				mockClient.EXPECT().Ping()
				client, err := NewClient("extensions.redis", config, mockClient)
				Expect(err).NotTo(HaveOccurred())
				mockClient.EXPECT().Close()
				err = client.Cleanup()
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Describe("EnterCriticalSection", func() {
			It("should lock in redis", func() {
				mockClient.EXPECT().Ping()
				client, err := NewClient("extensions.redis", config, mockClient)
				Expect(err).NotTo(HaveOccurred())
				t := 10 * time.Millisecond
				mockClient.EXPECT().SetNX("mock", gomock.Any(), 10*time.Millisecond).Return(&redis.BoolCmd{})
				client.EnterCriticalSection(mockClient, "mock", t, t, t)
				mockClient.EXPECT().Close()
				err = client.Cleanup()
				Expect(err).NotTo(HaveOccurred())
			})
		})

	})

	Describe("[Integration]", func() {
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
				Expect(lock.IsLocked()).To(Equal(true))
				lock2, err := client.EnterCriticalSection(client.Client, "lock", t, 0, 0)
				Expect(lock2).To(BeNil())
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
				Expect(lock.IsLocked()).To(Equal(true))
				err = client.LeaveCriticalSection(lock)
				Expect(err).NotTo(HaveOccurred())
				Expect(lock.IsLocked()).To(Equal(false))
				lock2, err := client.EnterCriticalSection(client.Client, "lock2", t, 0, 0)
				Expect(lock2).NotTo(BeNil())
				Expect(lock2.IsLocked()).To(Equal(true))
			})

		})
	})
})
