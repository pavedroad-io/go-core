package logger

import (
	"crypto/tls"
	"time"

	"github.com/Shopify/sarama"
)

type ProducerConfiguration struct {
	Brokers       []string
	Topic         string
	PartitionType string
	KeyType       string
	Compression   sarama.CompressionCodec
	Ack           sarama.RequiredAcks
	FlushFreq     time.Duration
	EnableTLS     bool
	TLSCfg        tls.Config
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

	if config.EnableTLS {
		cfg.Net.TLS.Enable = true
		cfg.Net.TLS.Config = &config.TLSCfg
	}

	return sarama.NewAsyncProducer(config.Brokers, cfg)
}
