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

package kafka

import (
	"fmt"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spf13/viper"
	"github.com/topfreegames/extensions/kafka/mocks"
	. "github.com/topfreegames/extensions/testing"
	"github.com/topfreegames/extensions/util"
)

var _ = Describe("Kafka Extension", func() {
	logger, hook := test.NewNullLogger()
	logger.Level = logrus.DebugLevel

	BeforeEach(func() {
		hook.Reset()
	})

	Describe("[Unit]", func() {
		var kafkaConsumerClientMock *mocks.ConsumerClientMock
		var consumer *Consumer

		startConsuming := func() {
			go func() {
				defer GinkgoRecover()
				consumer.ConsumeLoop()
			}()
			time.Sleep(5 * time.Millisecond)
		}

		publishEvent := func(ev kafka.Event) {
			consumer.Consumer.Events() <- ev
			//This time.sleep is necessary to allow go's goroutines to perform work
			//Please do not remove
			time.Sleep(5 * time.Millisecond)
		}

		BeforeEach(func() {
			kafkaConsumerClientMock = mocks.NewConsumerClientMock()
			config := viper.New()
			config.Set("extensions.kafkaconsumer.topics", []string{"com.games.test"})
			config.Set("extensions.kafkaconsumer.brokers", "localhost:9941")
			config.Set("extensions.kafkaconsumer.group", "testGroup")
			config.Set("extensions.kafkaconsumer.sessionTimeout", 6000)
			config.Set("extensions.kafkaconsumer.offsetResetStrategy", "latest")
			config.Set("extensions.kafkaconsumer.handleAllMessagesBeforeExiting", true)

			var err error
			consumer, err = NewConsumer(config, logger, kafkaConsumerClientMock)
			Expect(err).NotTo(HaveOccurred())
		})

		Describe("Creating new client", func() {
			It("should return connected client", func() {
				Expect(consumer.Brokers).NotTo(HaveLen(0))
				Expect(consumer.Topics).To(HaveLen(1))
				Expect(consumer.ConsumerGroup).NotTo(BeNil())
				Expect(consumer.msgChan).NotTo(BeClosed())
				Expect(consumer.Consumer).To(Equal(kafkaConsumerClientMock))
			})
		})

		Describe("Stop consuming", func() {
			It("should stop consuming", func() {
				consumer.run = true
				consumer.StopConsuming()
				Expect(consumer.run).To(BeFalse())
			})
		})

		Describe("Consume loop", func() {
			It("should fail if subscribing to topic fails", func() {
				kafkaConsumerClientMock.Error = fmt.Errorf("could not subscribe")
				err := consumer.ConsumeLoop()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("could not subscribe"))
			})

			It("should subscribe to topic", func() {
				startConsuming()
				defer consumer.StopConsuming()
				Eventually(kafkaConsumerClientMock.SubscribedTopics, 5).Should(HaveKey("com.games.test"))
			})

			It("should assign partition", func() {
				topic := consumer.Config.GetStringSlice("topics")[0]
				startConsuming()
				defer consumer.StopConsuming()
				part := kafka.TopicPartition{
					Topic:     &topic,
					Partition: 1,
				}

				event := kafka.AssignedPartitions{
					Partitions: []kafka.TopicPartition{part},
				}
				publishEvent(event)
				Eventually(kafkaConsumerClientMock.AssignedPartitions, 5).Should(ContainElement(part))
			})

			It("should log error if fails to assign partition", func() {
				topic := consumer.Config.GetStringSlice("topics")[0]
				startConsuming()
				defer consumer.StopConsuming()

				time.Sleep(5 * time.Millisecond)

				part := kafka.TopicPartition{
					Topic:     &topic,
					Partition: 1,
				}

				event := kafka.AssignedPartitions{
					Partitions: []kafka.TopicPartition{part},
				}

				kafkaConsumerClientMock.Error = fmt.Errorf("Failed to assign partition")
				publishEvent(event)
				Expect(hook.Entries).To(ContainLogMessage("error assigning partitions"))
			})

			It("should revoke partitions", func() {
				topic := consumer.Config.GetStringSlice("topics")[0]
				startConsuming()
				defer consumer.StopConsuming()
				part := kafka.TopicPartition{
					Topic:     &topic,
					Partition: 1,
				}
				kafkaConsumerClientMock.AssignedPartitions = []kafka.TopicPartition{part}
				Expect(kafkaConsumerClientMock.AssignedPartitions).NotTo(BeEmpty())

				event := kafka.RevokedPartitions{}
				publishEvent(event)
				Eventually(kafkaConsumerClientMock.AssignedPartitions, 5).Should(BeEmpty())
			})

			It("should stop loop if fails to revoke partitions", func() {
				topic := consumer.Config.GetStringSlice("topics")[0]
				startConsuming()
				defer consumer.StopConsuming()
				part := kafka.TopicPartition{
					Topic:     &topic,
					Partition: 1,
				}
				kafkaConsumerClientMock.AssignedPartitions = []kafka.TopicPartition{part}
				Expect(kafkaConsumerClientMock.AssignedPartitions).NotTo(BeEmpty())

				kafkaConsumerClientMock.Error = fmt.Errorf("Failed to unassign partition")
				event := kafka.RevokedPartitions{}
				publishEvent(event)
				Expect(hook.Entries).To(ContainLogMessage("error revoking partitions"))
			})

			It("should receive message", func() {
				topic := consumer.Config.GetStringSlice("topics")[0]
				startConsuming()
				defer consumer.StopConsuming()
				part := kafka.TopicPartition{
					Topic:     &topic,
					Partition: 1,
				}
				val := []byte("test")
				event := &kafka.Message{TopicPartition: part, Value: val}
				consumer.messagesReceived = 999

				publishEvent(event)
				Eventually(consumer.msgChan, 5).Should(Receive(&val))
				Expect(consumer.messagesReceived).To(BeEquivalentTo(1000))
			})

			It("should handle partition EOF", func() {
				startConsuming()
				defer consumer.StopConsuming()

				event := kafka.PartitionEOF{}
				publishEvent(event)

				Expect(hook.Entries).To(ContainLogMessage("Reached partition EOF."))
			})

			It("should handle offsets committed", func() {
				startConsuming()
				defer consumer.StopConsuming()

				event := kafka.OffsetsCommitted{}
				publishEvent(event)

				Expect(hook.Entries).To(ContainLogMessage("Offsets committed successfully."))
			})

			It("should handle error", func() {
				startConsuming()
				defer consumer.StopConsuming()

				event := kafka.Error{}
				publishEvent(event)

				Eventually(consumer.run, 5).Should(BeFalse())
				Expect(hook.Entries).To(ContainLogMessage("Error in Kafka connection."))
			})

			It("should handle unexpected message", func() {
				startConsuming()
				defer consumer.StopConsuming()

				event := &mocks.MockEvent{}
				publishEvent(event)

				Expect(hook.Entries).To(ContainLogMessage("Kafka event not recognized."))
			})
		})

		Describe("Configuration Defaults", func() {
			It("should configure defaults", func() {
				cnf := viper.New()
				cons, err := NewConsumer(cnf, logger, kafkaConsumerClientMock)
				Expect(err).NotTo(HaveOccurred())
				cnf = cons.Config
				cons.loadConfigurationDefaults()

				Expect(cons.Config.GetStringSlice("topics")).To(Equal([]string{"com.games.test"}))
				Expect(cons.Config.GetString("brokers")).To(Equal("localhost:9092"))
				Expect(cons.Config.GetString("group")).To(Equal("test"))
				Expect(cons.Config.GetInt("sessionTimeout")).To(Equal(6000))
				Expect(cons.Config.GetString("offsetResetStrategy")).To(Equal("latest"))
				Expect(cons.Config.GetBool("handleAllMessagesBeforeExiting")).To(BeTrue())
			})

			It("should read a config with prefix", func() {
				cnf := viper.New()
				cnf.Set("test.sessionTimeout", 123)
				cons, err := NewConsumerWithPrefix(cnf, logger, "test", kafkaConsumerClientMock)
				Expect(err).NotTo(HaveOccurred())
				cnf = cons.Config
				cons.loadConfigurationDefaults()

				Expect(cons.Config.GetStringSlice("topics")).To(Equal([]string{"com.games.test"}))
				Expect(cons.Config.GetString("brokers")).To(Equal("localhost:9092"))
				Expect(cons.Config.GetString("group")).To(Equal("test"))
				Expect(cons.Config.GetInt("sessionTimeout")).To(Equal(123))
				Expect(cons.Config.GetString("offsetResetStrategy")).To(Equal("latest"))
				Expect(cons.Config.GetBool("handleAllMessagesBeforeExiting")).To(BeTrue())
			})
		})

		Describe("Pending Messages Waiting Group", func() {
			It("should return the waiting group", func() {
				pmwg := consumer.PendingMessagesWaitGroup()
				Expect(pmwg).NotTo(BeNil())
			})
		})

		Describe("WaitUntilReady", func() {
			It("should receive kafka.AssignedPartitions and be ready", func() {
				go consumer.ConsumeLoop()
				defer consumer.StopConsuming()
				topic := consumer.Config.GetStringSlice("topics")[0]
				part := kafka.TopicPartition{
					Topic:     &topic,
					Partition: 1,
				}

				event := kafka.AssignedPartitions{
					Partitions: []kafka.TopicPartition{part},
				}
				publishEvent(event)
				Eventually(kafkaConsumerClientMock.AssignedPartitions, 5).Should(ContainElement(part))
				consumer.WaitUntilReady() // should not block
			})

			It("should block forever if no kafka.AssignedPartitions arrive", func() {
				go consumer.ConsumeLoop()
				defer consumer.StopConsuming()
				ready := make(chan bool, 1)
				go func() {
					consumer.WaitUntilReady()
					ready <- true
				}()
				ticker := time.NewTicker(10 * time.Millisecond)
				select {
				case <-ready:
					panic(fmt.Errorf("should not happen"))
				case <-ticker.C:
					return // synthetic timeout
				}
			})
		})

		Describe("Cleanup", func() {
			It("should stop running upon cleanup", func() {
				consumer.run = true
				err := consumer.Cleanup()
				Expect(err).NotTo(HaveOccurred())
				Expect(consumer.run).To(BeFalse())
			})

			It("should close connection to kafka upon cleanup", func() {
				err := consumer.Cleanup()
				Expect(err).NotTo(HaveOccurred())
				Expect(kafkaConsumerClientMock.Closed).To(BeTrue())
			})

			It("should return error when closing connection to kafka upon cleanup", func() {
				kafkaConsumerClientMock.Error = fmt.Errorf("Could not close connection")
				err := consumer.Cleanup()
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal(kafkaConsumerClientMock.Error.Error()))
			})
		})
	})

	XDescribe("[Integration]", func() {
		var config *viper.Viper
		var err error

		BeforeEach(func() {
			config, err = util.NewViperWithConfigFile("./config/test.yaml")
			Expect(err).NotTo(HaveOccurred())
		})

		Describe("Creating new client", func() {
			It("should return connected client", func() {
				client, err := NewConsumer(config, logger)
				Expect(err).NotTo(HaveOccurred())

				Expect(client.Brokers).NotTo(HaveLen(0))
				Expect(client.Topics).To(HaveLen(1))
				Expect(client.ConsumerGroup).NotTo(BeNil())
				Expect(client.msgChan).NotTo(BeClosed())
			})
		})

		Describe("ConsumeLoop", func() {
			It("should consume message and add it to msgChan", func() {
				client, err := NewConsumer(config, logger)
				Expect(err).NotTo(HaveOccurred())
				Expect(client).NotTo(BeNil())
				defer client.StopConsuming()
				go client.ConsumeLoop()

				Eventually(func() []*logrus.Entry {
					return hook.Entries
				}, 10*time.Second).Should(ContainLogMessage("reached EOF at com.games.teste[0]@0(Broker: No more messages)"))
				p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": client.Brokers})
				Expect(err).NotTo(HaveOccurred())
				err = p.Produce(
					&kafka.Message{
						TopicPartition: kafka.TopicPartition{
							Topic:     &client.Topics[0],
							Partition: kafka.PartitionAny,
						},
						Value: []byte("Hello Go!")},
					nil,
				)
				Expect(err).NotTo(HaveOccurred())
				Eventually(client.msgChan, 10*time.Second).Should(Receive(Equal([]byte("Hello Go!"))))
			})
		})
	})
})
