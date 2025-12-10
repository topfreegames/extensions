/*
 * Copyright (c) 2017 TFG Co
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
	"github.com/confluentinc/confluent-kafka-go/kafka"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spf13/viper"
	"github.com/topfreegames/extensions/v9/kafka/mocks"
	"github.com/topfreegames/extensions/v9/util"
)

var _ = Describe("Producer Extension", func() {
	var config *viper.Viper
	var mockProducer *mocks.ProducerClientMock
	logger, hook := test.NewNullLogger()
	logger.Level = logrus.DebugLevel

	BeforeEach(func() {
		var err error
		config, err = util.NewViperWithConfigFile("../config/test.yaml")
		Expect(err).NotTo(HaveOccurred())
		mockProducer = mocks.NewProducerClientMock()
		mockProducer.StartConsumingMessagesInProduceChannel()
		hook.Reset()
	})

	Describe("[Unit]", func() {
		Describe("Handling Message Sent", func() {
			It("should send message", func() {
				Producer, err := NewProducer(config, logger, mockProducer)
				Expect(err).NotTo(HaveOccurred())
				Producer.SendAsync([]byte("test message"), "test-topic")
				Eventually(func() int {
					return Producer.Producer.(*mocks.ProducerClientMock).SentMessages
				}).Should(Equal(1))
			})

			It("should log kafka responses", func() {
				Producer, err := NewProducer(config, logger, mockProducer)
				Expect(err).NotTo(HaveOccurred())
				testTopic := "ttopic"
				Producer.Producer.Events() <- &kafka.Message{
					TopicPartition: kafka.TopicPartition{
						Topic:     &testTopic,
						Partition: 0,
						Offset:    0,
						Error:     nil,
					},
				}
				Eventually(func() string {
					return hook.LastEntry().Message
				}).Should(ContainSubstring("delivered feedback to topic"))
			})
		})
	})

	Describe("Configuration Defaults", func() {
		It("should configure defaults", func() {
			cnf := viper.New()
			cons, err := NewProducer(cnf, logger, mockProducer)
			Expect(err).NotTo(HaveOccurred())

			Expect(cons.Config.GetString("brokers")).To(Equal("localhost:9092"))
		})

		It("should read a config with prefix", func() {
			cnf := viper.New()
			cnf.Set("brokers", "localhost:1234")
			cons, err := NewProducerWithPrefix(cnf, logger, "test", mockProducer)
			Expect(err).NotTo(HaveOccurred())

			Expect(cons.Config.GetString("brokers")).To(Equal("localhost:1234"))

		})
	})

	XDescribe("[Integration]", func() {
		Describe("Creating new producer", func() {
			It("should return connected client", func() {
				kafkaProducer, err := NewProducer(config, logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(kafkaProducer).NotTo(BeNil())
				Expect(kafkaProducer.Producer).NotTo(BeNil())
				Expect(kafkaProducer.Config).NotTo(BeNil())
				Expect(kafkaProducer.Logger).NotTo(BeNil())
			})
		})
	})

})
