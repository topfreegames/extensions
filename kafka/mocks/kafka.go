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

package mocks

import "github.com/confluentinc/confluent-kafka-go/kafka"

// ProducerClientMock  should be used for tests that need to send messages to Kafka
type ProducerClientMock struct {
	EventsChan   chan kafka.Event
	ProduceChan  chan *kafka.Message
	SentMessages int
}

// MockEvent implements kafka.Event
type MockEvent struct {
	Message *kafka.Message
}

//String returns string
func (m *MockEvent) String() string {
	return string(m.Message.Value)
}

// NewProducerClientMock creates a new instance
func NewProducerClientMock() *ProducerClientMock {
	k := &ProducerClientMock{
		EventsChan:   make(chan kafka.Event),
		ProduceChan:  make(chan *kafka.Message),
		SentMessages: 0,
	}
	return k
}

// StartConsumingMessagesInProduceChannel starts to consume messages in produce channel and incrementing sentMessages
func (k *ProducerClientMock) StartConsumingMessagesInProduceChannel() {
	go func() {
		for msg := range k.ProduceChan {
			k.SentMessages++
			k.EventsChan <- msg
		}
	}()
}

// Events returns the mock events channel
func (k *ProducerClientMock) Events() chan kafka.Event {
	return k.EventsChan
}

// ProduceChannel returns the mock produce channel
func (k *ProducerClientMock) ProduceChannel() chan *kafka.Message {
	return k.ProduceChan
}

// ConsumerClientMock  should be used for tests that need to send messages to Kafka
type ConsumerClientMock struct {
	SubscribedTopics   map[string]interface{}
	EventsChan         chan kafka.Event
	AssignedPartitions []kafka.TopicPartition
	Closed             bool
	Error              error
}

// NewConsumerClientMock creates a new instance
func NewConsumerClientMock(errorOrNil ...error) *ConsumerClientMock {
	var err error
	if len(errorOrNil) == 1 {
		err = errorOrNil[0]
	}
	k := &ConsumerClientMock{
		SubscribedTopics:   map[string]interface{}{},
		EventsChan:         make(chan kafka.Event),
		AssignedPartitions: []kafka.TopicPartition{},
		Closed:             false,
		Error:              err,
	}
	return k
}

//SubscribeTopics mock
func (k *ConsumerClientMock) SubscribeTopics(topics []string, callback kafka.RebalanceCb) error {
	if k.Error != nil {
		return k.Error
	}
	for _, topic := range topics {
		k.SubscribedTopics[topic] = callback
	}
	return nil
}

//Events mock
func (k *ConsumerClientMock) Events() chan kafka.Event {
	return k.EventsChan
}

//Assign mock
func (k *ConsumerClientMock) Assign(partitions []kafka.TopicPartition) error {
	if k.Error != nil {
		return k.Error
	}
	k.AssignedPartitions = partitions
	return nil
}

//Unassign mock
func (k *ConsumerClientMock) Unassign() error {
	if k.Error != nil {
		return k.Error
	}
	k.AssignedPartitions = []kafka.TopicPartition{}
	return nil
}

//Close mock
func (k *ConsumerClientMock) Close() error {
	if k.Error != nil {
		return k.Error
	}
	k.Closed = true
	return nil
}
