package main

import (
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	"github.com/pavedroad-io/core/go/logger"
)

// Create loggers for zap and logrus
// TODO should KafkaProducerCfg be an anonymous member?
// Note that Kafka producer flush frequency is set to one half second
// Thus the one second sleeps below will cause the queue to be flushed

func main() {
	config := logger.Configuration{
		LogLevel:          logger.Info,
		EnableCloudEvents: true,
		EnableKafka:       true,
		KafkaFormat:       logger.TypeCEFormat,
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
		EnableConsole:        true,
		ConsoleFormat:        logger.TypeTextFormat,
		ConsoleLevelTruncate: true,
		EnableFile:           false,
		FileFormat:           logger.TypeJSONFormat,
		FileLevelTruncate:    false,
		FileLocation:         "pavedroad.log",
	}

	// try a zap logger

	log, err := logger.NewLogger(config, logger.InstanceZapLogger)
	if err != nil {
		fmt.Printf("Could not instantiate zap logger %s", err.Error())
	} else {
		log.Debugf("Zap using Debugf (should not appear)")
		log.Infof("Zap using Infof")
		log.Warnf("Zap using Warnf")
		log.Errorf("Zap using Errorf")
		time.Sleep(time.Second)
	}

	// try a logrus logger

	log, err = logger.NewLogger(config, logger.InstanceLogrusLogger)
	if err != nil {
		fmt.Printf("Could not instantiate logrus logger %s", err.Error())
	} else {
		log.Debugf("Logrus using Debugf (should not appear)")
		log.Infof("Logrus using Infof")
		log.Warnf("Logrus using Warnf")
		log.Errorf("Logrus using Errorf")
		log.Print("Logrus using Print")
		log.Printf("Logrus using Printf")
		log.Println("Logrus using Println")
		time.Sleep(time.Second)
	}
}
