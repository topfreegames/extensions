package interfaces

import "github.com/IBM/sarama"

type SyncProducer interface {
	sarama.SyncProducer
}
