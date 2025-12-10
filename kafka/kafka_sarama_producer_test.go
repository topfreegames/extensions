/*
 * Copyright (c) 2018 TFG Co
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

package kafka

import (
	"github.com/IBM/sarama"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"go.uber.org/mock/gomock"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spf13/viper"
	"github.com/topfreegames/extensions/v9/kafka/mocks"
	"github.com/topfreegames/extensions/v9/util"
)

var _ = XDescribe("SyncProducer Extension", func() {
	var config *viper.Viper
	var mockProducer *mocks.MockSyncProducer
	var mockCtrl *gomock.Controller
	logger, hook := test.NewNullLogger()
	logger.Level = logrus.DebugLevel

	BeforeEach(func() {
		var err error
		config, err = util.NewViperWithConfigFile("../config/test.yaml")
		Expect(err).NotTo(HaveOccurred())
		mockCtrl = gomock.NewController(GinkgoT())
		mockProducer = mocks.NewMockSyncProducer(mockCtrl)
	})

	AfterEach(func() {
		mockCtrl.Finish()
		hook.Reset()
	})

	Describe("[Unit]", func() {
		It("should produce a message", func() {
			topic := "test-topic"
			message := "test-message"
			c := sarama.NewConfig()
			c.Producer.Return.Errors = true
			c.Producer.Return.Successes = true
			producer, err := NewSyncProducer(config, logger, c, mockProducer)
			Expect(err).NotTo(HaveOccurred())
			mockProducer.EXPECT().SendMessage(gomock.Eq(&sarama.ProducerMessage{
				Topic: topic,
				Value: sarama.ByteEncoder(message),
			}))
			_, _, err = producer.Produce(topic, []byte(message))
			Expect(err).NotTo(HaveOccurred())
		})

		It("should be configured with default brokers if no config is passed", func() {
			config := viper.New()
			c := sarama.NewConfig()
			c.Producer.Return.Errors = true
			c.Producer.Return.Successes = true
			producer, err := NewSyncProducer(config, logger, c, mockProducer)
			Expect(err).NotTo(HaveOccurred())
			Expect(producer.Brokers).To(Equal("localhost:9092"))
		})

		It("should be configured with brokers if config is passed", func() {
			c := sarama.NewConfig()
			c.Producer.Return.Errors = true
			c.Producer.Return.Successes = true
			producer, err := NewSyncProducer(config, logger, c, mockProducer)
			Expect(err).NotTo(HaveOccurred())
			Expect(producer.Brokers).To(Equal("localhost:9941"))
		})
	})

	Describe("Configuration Defaults", func() {
		It("should configure defaults", func() {
			cnf := viper.New()
			c := sarama.NewConfig()
			cons, err := NewSyncProducer(cnf, logger, c, mockProducer)
			Expect(err).NotTo(HaveOccurred())

			Expect(cons.config.GetString("brokers")).To(Equal("localhost:9092"))
		})

		It("should read a config with prefix", func() {
			cnf := viper.New()
			cnf.Set("brokers", "localhost:1234")
			c := sarama.NewConfig()
			cons, err := NewSyncProducerWithPrefix(cnf, logger, c, "test", mockProducer)
			Expect(err).NotTo(HaveOccurred())

			Expect(cons.config.GetString("brokers")).To(Equal("localhost:1234"))

		})
	})

	XDescribe("[Integration]", func() {
		Describe("Create a new producer", func() {
			It("should return a connected client", func() {
				c := sarama.NewConfig()
				c.Producer.Return.Errors = true
				c.Producer.Return.Successes = true
				producer, err := NewSyncProducer(config, logger, c)
				Expect(err).NotTo(HaveOccurred())
				Expect(producer.Producer).NotTo(BeNil())
			})
		})
	})
})
