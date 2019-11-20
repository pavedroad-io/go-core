package main

import (
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	"github.com/pavedroad-io/go-core/logger"
)

// Create loggers for zap and logrus
// TODO should KafkaProducerCfg be an anonymous member?
// Note that Kafka producer flush frequency is set to one half second
// Thus the one second sleeps below will cause the queue to be flushed

func main() {
	config := logger.Configuration{
		LogLevel:        logger.Info,
		EnableKafka:     true,
		KafkaJSONFormat: true,
		KafkaProducerCfg: logger.ProducerConfiguration{
			Brokers:       []string{"localhost:9092"},
			Topic:         "logs",
			PartitionType: "random",
			KeyType:       "hash",
			Compression:   sarama.CompressionSnappy,
			Ack:           sarama.WaitForLocal,
			FlushFreq:     500, // milliseconds
			EnableTLS:     false,
		},
		EnableConsole:     true,
		ConsoleJSONFormat: false,
		EnableFile:        false,
		FileJSONFormat:    true,
		FileLocation:      "pavedroad.log",
	}

	// try a zap logger

	newlog, err := logger.NewLogger(config, logger.InstanceZapLogger)
	if err != nil {
		fmt.Printf("Could not instantiate zap logger %s", err.Error())
	} else {
		newlog.Infof("Zap using Infof")
		time.Sleep(time.Second)

		jsonlog := logger.WithFields(logger.Fields{"key1": "value1"})
		jsonlog.Infof("Zap using Infof with fields")
		time.Sleep(time.Second)
	}

	// try a logrus logger

	newlog, err = logger.NewLogger(config, logger.InstanceLogrusLogger)
	if err != nil {
		fmt.Printf("Could not instantiate logrus logger %s", err.Error())
	} else {
		newlog.Debugf("Logrus using Debugf (should not log)")
		newlog.Infof("Logrus using Infof")
		newlog.Warnf("Logrus using Warnf")
		newlog.Errorf("Logrus using Errorf")
		newlog.Print("Logrus using Print")
		newlog.Printf("Logrus using Printf")
		newlog.Println("Logrus using Println")
		time.Sleep(time.Second)

		jsonlog := logger.WithFields(logger.Fields{"key1": "value1"})
		jsonlog.Debugf("Logrus using Debugf with fields (should not log)")
		jsonlog.Infof("Logrus using Infof with fields")
		jsonlog.Warnf("Logrus using Warnf with fields")
		jsonlog.Errorf("Logrus using Errorf with fields")
		jsonlog.Print("Logrus using Print with fields")
		jsonlog.Printf("Logrus using Printf with fields")
		jsonlog.Println("Logrus using Println with fields")
		time.Sleep(time.Second)
	}
}
