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
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/IBM/sarama"
	"github.com/spf13/viper"
)

// SyncProducer is a kafka producer using sarama lib
type SyncProducer struct {
	config      *viper.Viper
	logger      *log.Logger
	kafkaConfig *sarama.Config
	Brokers     string
	Producer    sarama.SyncProducer
}

// NewSyncProducer creates a new kafka sync producer
func NewSyncProducer(
	config *viper.Viper,
	logger *log.Logger,
	kafkaConfig *sarama.Config,
	clientOrNil ...sarama.SyncProducer,
) (*SyncProducer, error) {
	return NewSyncProducerWithPrefix(config, logger, kafkaConfig, "extensions.kafkaproducer", clientOrNil...)
}

// NewSyncProducerWithPrefix creates a new kafka sync producer
func NewSyncProducerWithPrefix(
	config *viper.Viper,
	logger *log.Logger,
	kafkaConfig *sarama.Config,
	prefix string,
	clientOrNil ...sarama.SyncProducer,
) (*SyncProducer, error) {
	if prefix != "" {
		prefix += "."
	}
	s := &SyncProducer{
		config:      config,
		logger:      logger,
		kafkaConfig: kafkaConfig,
	}
	var producer sarama.SyncProducer
	if len(clientOrNil) == 1 {
		producer = clientOrNil[0]
	}
	err := s.configure(producer, prefix)
	return s, err
}

func (s *SyncProducer) loadConfigurationDefaults(prefix string) {
	s.config.SetDefault(prefix+"brokers", "localhost:9092")
}

func (s *SyncProducer) configure(producer sarama.SyncProducer, prefix string) error {
	s.loadConfigurationDefaults(prefix)
	s.Brokers = s.config.GetString(prefix + "brokers")
	l := s.logger.WithFields(log.Fields{
		"brokers": s.Brokers,
	})
	l.Debug("configuring kafka producer")
	if producer == nil {
		p, err := sarama.NewSyncProducer(strings.Split(s.Brokers, ","), s.kafkaConfig)
		s.Producer = p
		if err != nil {
			l.WithError(err).Error("error configuring kafka producer client")
			return err
		}
	} else {
		s.Producer = producer
	}
	l.Info("kafka producer initialized")
	return nil
}

// Produce produces a message
func (s *SyncProducer) Produce(topic string, message []byte) (int32, int64, error) {
	s.logger.WithFields(log.Fields{
		"topic":   topic,
		"message": string(message),
	}).Debug("kafka sync producer extension sending message")
	m := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(message),
	}
	return s.Producer.SendMessage(m)
}
