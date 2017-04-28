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
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
	"github.com/topfreegames/extensions/redis/mocks"
)

var _ = Describe("Redis Extension", func() {
	var config *viper.Viper

	BeforeEach(func() {
		config = viper.New()
		config.SetConfigFile("../config/test.yaml")
		Expect(config.ReadInConfig()).NotTo(HaveOccurred())
	})

	Describe("[Unit]", func() {
		var mockClient *mocks.RedisMock
		BeforeEach(func() {
			mockClient = mocks.NewRedisMock("PONG")
		})

		Describe("Connect", func() {
			It("Should use config to load connection details", func() {
				client, err := NewClient("extensions.redis", config, mockClient)
				Expect(err).NotTo(HaveOccurred())
				Expect(client.Options).NotTo(BeNil())
			})
		})

		Describe("IsConnected", func() {
			It("should verify that db is connected", func() {
				client, err := NewClient("extensions.redis", config, mockClient)
				Expect(err).NotTo(HaveOccurred())
				Expect(client.IsConnected()).To(BeTrue())
				Expect(mockClient.PingCount).NotTo(Equal(0))
			})

			It("should not be connected if error", func() {
				connErr := fmt.Errorf("Could not connect")
				client, err := NewClient("extensions.redis", config, mockClient)
				Expect(err).NotTo(HaveOccurred())
				mockClient.Error = connErr
				Expect(client.IsConnected()).To(BeFalse())
				Expect(mockClient.PingCount).NotTo(Equal(0))
			})

			It("should not be connected if something other than 'PONG' returned", func() {
				client, err := NewClient("extensions.redis", config, mockClient)
				Expect(err).NotTo(HaveOccurred())
				mockClient.PingReponse = "WHATEVER"
				Expect(client.IsConnected()).To(BeFalse())
				Expect(mockClient.PingCount).NotTo(Equal(0))
			})
		})

		Describe("Close", func() {
			It("should close if no errors", func() {
				client, err := NewClient("extensions.redis", config, mockClient)
				Expect(err).NotTo(HaveOccurred())
				err = client.Close()
				Expect(err).NotTo(HaveOccurred())
				Expect(mockClient.Closed).To(BeTrue())
			})

			It("should return error", func() {
				connErr := fmt.Errorf("Could not close")
				client, err := NewClient("extensions.redis", config, mockClient)
				Expect(err).NotTo(HaveOccurred())

				mockClient.Error = connErr

				err = client.Close()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("Could not close"))
			})
		})

		Describe("WaitForConnection", func() {
			It("should wait for connection", func() {
				client, err := NewClient("extensions.redis", config, mockClient)
				Expect(err).NotTo(HaveOccurred())

				err = client.WaitForConnection(1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should error waiting for connection", func() {
				pErr := fmt.Errorf("Connection failed")
				client, err := NewClient("extensions.redis", config, mockClient)
				Expect(err).NotTo(HaveOccurred())
				mockClient.Error = pErr

				err = client.WaitForConnection(10)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("timed out waiting for Redis to connect"))
			})
		})

		Describe("Cleanup", func() {
			It("should close connection", func() {
				client, err := NewClient("extensions.redis", config, mockClient)
				Expect(err).NotTo(HaveOccurred())
				err = client.Cleanup()
				Expect(err).NotTo(HaveOccurred())
				Expect(mockClient.Closed).To(BeTrue())
			})

			It("should return error if error when closing connection", func() {
				pErr := fmt.Errorf("failed to close connection")
				client, err := NewClient("extensions.redis", config, mockClient)
				Expect(err).NotTo(HaveOccurred())
				mockClient.Error = pErr
				err = client.Cleanup()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to close connection"))
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
	})
})
