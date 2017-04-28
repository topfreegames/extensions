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
	"sync"

	"github.com/Sirupsen/logrus"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	raven "github.com/getsentry/raven-go"
	"github.com/spf13/viper"
	"github.com/topfreegames/extensions/kafka/interfaces"
	"github.com/topfreegames/extensions/util"
)

// Consumer for getting push requests
type Consumer struct {
	Brokers                        string
	Config                         *viper.Viper
	Consumer                       interfaces.KafkaConsumerClient
	ConsumerGroup                  string
	ChannelSize                    int
	Logger                         *logrus.Logger
	messagesReceived               int64
	msgChan                        chan []byte
	OffsetResetStrategy            string
	run                            bool
	SessionTimeout                 int
	Topics                         []string
	pendingMessagesWG              *sync.WaitGroup
	HandleAllMessagesBeforeExiting bool
}

// NewConsumer for creating a new Consumer instance
func NewConsumer(
	config *viper.Viper,
	logger *logrus.Logger,
	clientOrNil ...interfaces.KafkaConsumerClient,
) (*Consumer, error) {
	q := &Consumer{
		Config:            config,
		Logger:            logger,
		messagesReceived:  0,
		pendingMessagesWG: nil,
	}
	var client interfaces.KafkaConsumerClient
	if len(clientOrNil) == 1 {
		client = clientOrNil[0]
	}
	err := q.configure(client)
	if err != nil {
		return nil, err
	}
	return q, nil
}

func (q *Consumer) loadConfigurationDefaults() {
	q.Config.SetDefault("extensions.kafkaconsumer.topics", []string{"com.games.test"})
	q.Config.SetDefault("extensions.kafkaconsumer.brokers", "localhost:9092")
	q.Config.SetDefault("extensions.kafkaconsumer.channelSize", 100)
	q.Config.SetDefault("extensions.kafkaconsumer.group", "test")
	q.Config.SetDefault("extensions.kafkaconsumer.sessionTimeout", 6000)
	q.Config.SetDefault("extensions.kafkaconsumer.offsetResetStrategy", "latest")
	q.Config.SetDefault("extensions.kafkaconsumer.handleAllMessagesBeforeExiting", true)
}

func (q *Consumer) configure(client interfaces.KafkaConsumerClient) error {
	q.OffsetResetStrategy = q.Config.GetString("extensions.kafkaconsumer.offsetResetStrategy")
	q.Brokers = q.Config.GetString("extensions.kafkaconsumer.brokers")
	q.ConsumerGroup = q.Config.GetString("extensions.kafkaconsumer.group")
	q.SessionTimeout = q.Config.GetInt("extensions.kafkaconsumer.sessionTimeout")
	q.Topics = q.Config.GetStringSlice("extensions.kafkaconsumer.topics")
	q.ChannelSize = q.Config.GetInt("extensions.kafkaconsumer.channelSize")
	q.HandleAllMessagesBeforeExiting = q.Config.GetBool("extensions.kafkaconsumer.handleAllMessagesBeforeExiting")

	q.msgChan = make(chan []byte, q.ChannelSize)

	if q.HandleAllMessagesBeforeExiting {
		var wg sync.WaitGroup
		q.pendingMessagesWG = &wg
	}

	err := q.configureConsumer(client)
	if err != nil {
		return err
	}
	return nil
}

func (q *Consumer) configureConsumer(client interfaces.KafkaConsumerClient) error {
	l := q.Logger.WithFields(logrus.Fields{
		"method":                          "configureConsumer",
		"bootstrap.servers":               q.Brokers,
		"group.id":                        q.ConsumerGroup,
		"session.timeout.ms":              q.SessionTimeout,
		"go.events.channel.enable":        true,
		"go.application.rebalance.enable": true,
		"enable.auto.commit":              true,
		"default.topic.config": kafka.ConfigMap{
			"auto.offset.reset":  q.OffsetResetStrategy,
			"auto.commit.enable": true,
		},
		"topics": q.Topics,
	})
	l.Debug("configuring kafka queue extension")

	if client == nil {
		c, err := kafka.NewConsumer(&kafka.ConfigMap{
			"bootstrap.servers":               q.Brokers,
			"group.id":                        q.ConsumerGroup,
			"session.timeout.ms":              q.SessionTimeout,
			"go.events.channel.enable":        true,
			"go.application.rebalance.enable": true,
			"enable.auto.commit":              true,
			"default.topic.config": kafka.ConfigMap{
				"auto.offset.reset":  q.OffsetResetStrategy,
				"auto.commit.enable": true,
			},
		})
		if err != nil {
			l.WithError(err).Error("error configuring kafka queue")
			return err
		}
		q.Consumer = c
	} else {
		q.Consumer = client
	}
	l.Info("kafka queue configured")
	return nil
}

// PendingMessagesWaitGroup returns the waitGroup that is incremented every time a push is consumed
func (q *Consumer) PendingMessagesWaitGroup() *sync.WaitGroup {
	return q.pendingMessagesWG
}

// StopConsuming stops consuming messages from the queue
func (q *Consumer) StopConsuming() {
	q.run = false
}

// MessagesChannel returns the channel that will receive all messages got from kafka
func (q *Consumer) MessagesChannel() *chan []byte {
	return &q.msgChan
}

// ConsumeLoop consume messages from the queue and put in messages to send channel
func (q *Consumer) ConsumeLoop() error {
	q.run = true
	l := q.Logger.WithFields(logrus.Fields{
		"method": "ConsumeLoop",
		"topics": q.Topics,
	})

	err := q.Consumer.SubscribeTopics(q.Topics, nil)
	if err != nil {
		l.WithError(err).Error("error subscribing to topics")
		return err
	}

	l.Info("successfully subscribed to topics")

	for q.run == true {
		select {
		case ev := <-q.Consumer.Events():
			switch e := ev.(type) {
			case kafka.AssignedPartitions:
				err = q.assignPartitions(e.Partitions)
				if err != nil {
					l.WithError(err).Error("error assigning partitions")
				}
			case kafka.RevokedPartitions:
				err = q.unassignPartitions()
				if err != nil {
					l.WithError(err).Error("error revoking partitions")
				}
			case *kafka.Message:
				q.receiveMessage(e.TopicPartition, e.Value)
			case kafka.PartitionEOF:
				q.handlePartitionEOF(ev)
			case kafka.OffsetsCommitted:
				q.handleOffsetsCommitted(ev)
			case kafka.Error:
				q.handleError(ev)
				q.StopConsuming()
				return e
			default:
				q.handleUnrecognized(e)
			}
		}
	}

	return nil
}

func (q *Consumer) assignPartitions(partitions []kafka.TopicPartition) error {
	l := q.Logger.WithFields(logrus.Fields{
		"method":     "assignPartitions",
		"partitions": fmt.Sprintf("%v", partitions),
	})

	l.Debug("Assigning partitions...")
	err := q.Consumer.Assign(partitions)
	if err != nil {
		l.WithError(err).Error("Failed to assign partitions.")
		return err
	}
	l.Info("Partitions assigned.")
	return nil
}

func (q *Consumer) unassignPartitions() error {
	l := q.Logger.WithFields(logrus.Fields{
		"method": "unassignPartitions",
	})

	l.Debug("Unassigning partitions...")
	err := q.Consumer.Unassign()
	if err != nil {
		l.WithError(err).Error("Failed to unassign partitions.")
		return err
	}
	l.Info("Partitions unassigned.")
	return nil
}

func (q *Consumer) receiveMessage(topicPartition kafka.TopicPartition, value []byte) {
	l := q.Logger.WithFields(logrus.Fields{
		"method": "receiveMessage",
	})

	l.Debug("Processing received message...")

	q.messagesReceived++
	if q.messagesReceived%1000 == 0 {
		l.Infof("messages from kafka: %d", q.messagesReceived)
	}
	l.Debugf("message on %s:\n%s\n", topicPartition, string(value))
	if q.pendingMessagesWG != nil {
		q.pendingMessagesWG.Add(1)
	}
	q.msgChan <- value

	l.Debug("Received message processed.")
}

func (q *Consumer) handlePartitionEOF(ev kafka.Event) {
	l := q.Logger.WithFields(logrus.Fields{
		"method":    "handlePartitionEOF",
		"partition": fmt.Sprintf("%v", ev),
	})

	l.Debug("Reached partition EOF.")
}

func (q *Consumer) handleOffsetsCommitted(ev kafka.Event) {
	l := q.Logger.WithFields(logrus.Fields{
		"method":    "handleOffsetsCommitted",
		"partition": fmt.Sprintf("%v", ev),
	})

	l.Debug("Offsets committed successfully.")
}

func (q *Consumer) handleError(ev kafka.Event) {
	l := q.Logger.WithFields(logrus.Fields{
		"method": "handleError",
	})
	err := ev.(error)
	raven.CaptureError(err, map[string]string{
		"version":   util.Version,
		"extension": "kafka-consumer",
	})
	l.WithError(err).Error("Error in Kafka connection.")
}

func (q *Consumer) handleUnrecognized(ev kafka.Event) {
	l := q.Logger.WithFields(logrus.Fields{
		"method": "handleUnrecognized",
		"event":  fmt.Sprintf("%v", ev),
	})
	l.Warn("Kafka event not recognized.")
}

//Cleanup closes kafka consumer connection
func (q *Consumer) Cleanup() error {
	if q.run {
		q.StopConsuming()
	}
	if q.Consumer != nil {
		err := q.Consumer.Close()
		if err != nil {
			return err
		}
	}

	return nil
}
