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

package pg

import (
	"errors"

	"go.uber.org/mock/gomock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/spf13/viper"
	"github.com/topfreegames/extensions/v9/pg/mocks"
)

var _ = Describe("PG Extension", func() {
	var config *viper.Viper
	var mockCtrl *gomock.Controller
	var mockDb *mocks.MockDB
	var mockTxWrapper *mocks.MockTxWrapper

	BeforeEach(func() {
		config = viper.New()
		config.SetConfigFile("../config/test.yaml")
		Expect(config.ReadInConfig()).NotTo(HaveOccurred())
	})

	Describe("[Unit]", func() {
		BeforeEach(func() {
			mockCtrl = gomock.NewController(GinkgoT())
			mockDb = mocks.NewMockDB(mockCtrl)
			mockTxWrapper = mocks.NewMockTxWrapper(mockCtrl)
		})

		AfterEach(func() {
			mockCtrl.Finish()
		})

		Describe("Connect", func() {
			It("Should use config to load connection details", func() {
				client, err := NewClient("extensions.pg", config, mockDb, mockTxWrapper)
				Expect(err).NotTo(HaveOccurred())
				Expect(client.Options).NotTo(BeNil())
			})
		})

		Describe("IsConnected", func() {
			It("should verify that db is connected", func() {
				mockDb.EXPECT().Context()
				mockDb.EXPECT().Exec("select 1").Return(NewTestResult(nil, 1), nil)
				client, err := NewClient("extensions.pg", config, mockDb, mockTxWrapper)
				Expect(err).NotTo(HaveOccurred())
				Expect(client.IsConnected()).To(BeTrue())
			})

			It("should not be connected if error", func() {
				client, err := NewClient("extensions.pg", config, mockDb, mockTxWrapper)
				Expect(err).NotTo(HaveOccurred())
				mockDb.EXPECT().Context()
				mockDb.EXPECT().Exec("select 1").Return(NewTestResult(nil, 1), errors.New("could not connect"))
				Expect(client.IsConnected()).To(BeFalse())
			})

			It("should not be connected if zero rows returned", func() {
				client, err := NewClient("extensions.pg", config, mockDb, mockTxWrapper)
				Expect(err).NotTo(HaveOccurred())
				mockDb.EXPECT().Context()
				mockDb.EXPECT().Exec("select 1").Return(NewTestResult(nil, 0), nil)
				Expect(client.IsConnected()).To(BeFalse())
			})
		})

		Describe("Close", func() {
			It("should close if no errors", func() {
				mockDb.EXPECT().Close()
				client, err := NewClient("extensions.pg", config, mockDb, mockTxWrapper)
				Expect(err).NotTo(HaveOccurred())
				err = client.Close()
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return error", func() {
				client, err := NewClient("extensions.pg", config, mockDb, mockTxWrapper)
				Expect(err).NotTo(HaveOccurred())

				mockDb.EXPECT().Close().Return(errors.New("could not close"))
				err = client.Close()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not close"))
			})
		})

		Describe("WaitForConnection", func() {
			It("should wait for connection", func() {
				mockDb.EXPECT().Context()
				mockDb.EXPECT().Exec("select 1").Return(NewTestResult(nil, 1), nil)
				client, err := NewClient("extensions.pg", config, mockDb, mockTxWrapper)
				Expect(err).NotTo(HaveOccurred())

				err = client.WaitForConnection(1)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should error waiting for connection", func() {
				client, err := NewClient("extensions.pg", config, mockDb, mockTxWrapper)
				Expect(err).NotTo(HaveOccurred())

				mockDb.EXPECT().Context().AnyTimes()
				mockDb.EXPECT().Exec("select 1").Return(NewTestResult(nil, 0), nil).AnyTimes()
				err = client.WaitForConnection(3)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("timed out waiting for PostgreSQL to connect"))
			})
		})

		Describe("Cleanup", func() {
			It("should close connection", func() {
				client, err := NewClient("extensions.pg", config, mockDb, mockTxWrapper)
				Expect(err).NotTo(HaveOccurred())

				mockDb.EXPECT().Close()
				err = client.Cleanup()
				Expect(err).NotTo(HaveOccurred())
			})

			It("should return error if error when closing connection", func() {
				client, err := NewClient("extensions.pg", config, mockDb, mockTxWrapper)
				Expect(err).NotTo(HaveOccurred())

				mockDb.EXPECT().Close().Return(errors.New("failed to close connection"))
				err = client.Cleanup()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("failed to close connection"))
			})
		})
	})

	Describe("[Integration]", func() {
		XDescribe("Creating new client", func() {
			It("should return connected client", func() {
				client, err := NewClient("extensions.pg", config, nil, nil)
				Expect(err).NotTo(HaveOccurred())
				defer client.Close()
				Expect(client).NotTo(BeNil())
				client.WaitForConnection(10)
				Expect(client.IsConnected()).To(BeTrue())
			})
		})
	})
})
