package interfaces

import "github.com/Shopify/sarama"

type SyncProducer interface {
	sarama.SyncProducer
}
