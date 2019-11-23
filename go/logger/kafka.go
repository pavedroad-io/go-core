package logger

import (
	"crypto/tls"
	"fmt"
	"time"

	"github.com/Shopify/sarama"
)

type PartitionType int8

const (
	RandomPartition PartitionType = iota
	HashPartition
	RoundRobinPartition
)

type ProducerConfiguration struct {
	Brokers     []string
	Topic       string
	Partition   PartitionType
	KeyType     string
	Compression sarama.CompressionCodec
	Ack         sarama.RequiredAcks
	FlushFreq   time.Duration
	EnableTLS   bool
	TLSCfg      tls.Config
}

// Wrapper around sarama.NewAsyncProducer
// Partition and key types ignored for now

func NewAsyncProducer(config ProducerConfiguration) (sarama.AsyncProducer, error) {
	cfg := sarama.NewConfig()
	cfg.Producer.RequiredAcks = config.Ack
	cfg.Producer.Compression = config.Compression
	cfg.Producer.Flush.Frequency = config.FlushFreq * time.Millisecond
	cfg.Producer.Return.Successes = false
	cfg.Producer.Return.Errors = false

	switch config.Partition {
	case HashPartition:
		cfg.Producer.Partitioner = sarama.NewHashPartitioner
	case RandomPartition:
		cfg.Producer.Partitioner = sarama.NewRandomPartitioner
	case RoundRobinPartition:
		cfg.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	default:
		return nil, fmt.Errorf("Partition not supported: %v", config.Partition)
	}

	if config.EnableTLS {
		cfg.Net.TLS.Enable = true
		cfg.Net.TLS.Config = &config.TLSCfg
	}

	return sarama.NewAsyncProducer(config.Brokers, cfg)
}
