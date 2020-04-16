package main

import (
	"fmt"
	"os/user"
	"time"

	"github.com/pavedroad-io/go-core/logger"
)

// Create loggers for zap and logrus
// Note that Kafka producer flush frequency is set to one half second
// Thus the one second sleeps below will cause the queue to be flushed

func main() {
	user, _ := user.Current()
	config := logger.LoggerConfiguration{
		LogPackage:        logger.ZapType,
		LogLevel:          logger.InfoType,
		EnableTimeStamps:  true,
		EnableColorLevels: true,
		EnableCloudEvents: true,
		CloudEventsCfg: logger.CloudEventsConfiguration{
			SetID: logger.CEHMAC,
		},
		EnableKafka: true,
		KafkaFormat: logger.CEFormat,
		KafkaProducerCfg: logger.ProducerConfiguration{
			Brokers:       []string{"localhost:9092"},
			Topic:         "logs",
			Partition:     logger.RandomPartition,
			Key:           logger.FixedKey,
			KeyName:       user.Username,
			Compression:   logger.CompressionSnappy,
			AckWait:       logger.WaitForLocal,
			ProdFlushFreq: 500, // milliseconds
			EnableTLS:     false,
			EnableDebug:   false,
		},
		EnableConsole: true,
		ConsoleFormat: logger.TextFormat,
		EnableFile:    false,
		FileFormat:    logger.JSONFormat,
		FileLocation:  "pavedroad.log",
	}

	// try a zap logger

	log, err := logger.NewLogger(config)
	if err != nil {
		fmt.Printf("Could not instantiate zap logger: %s\n", err.Error())
	} else {
		log.Debugf("Zap using %s", "Debugf (should not appear)")
		log.Infof("Zap using %s", "Infof")
		log.Warnf("Zap using %s", "Warnf")
		log.Errorf("Zap using %s", "Errorf")
		log.Printf("Zap using %s", "Printf")
		log.Print("Zap using", "Print")
		log.Println("Zap using", "Println")
		time.Sleep(time.Second)
	}

	// try a logrus logger
	// switch to UUID ID and level key

	config.LogPackage = logger.LogrusType
	config.CloudEventsCfg.SetID = logger.CEUUID
	config.KafkaProducerCfg.Key = logger.LevelKey
	log, err = logger.NewLogger(config)
	if err != nil {
		fmt.Printf("Could not instantiate logrus logger: %s", err.Error())
	} else {
		log.Debugf("Logrus using %s", "Debugf (should not appear)")
		log.Infof("Logrus using %s", "Infof")
		log.Warnf("Logrus using %s", "Warnf")
		log.Errorf("Logrus using %s", "Errorf")
		log.Printf("Logrus using %s", "Printf")
		log.Print("Logrus using", "Print")
		log.Println("Logrus using", "Println")
		time.Sleep(time.Second)
	}

	// try setting key to message subject field value

	config.KafkaProducerCfg.Key = logger.ExtractedKey
	config.KafkaProducerCfg.KeyName = "subject"
	log, err = logger.NewLogger(config)
	if err != nil {
		fmt.Printf("Could not instantiate logrus logger: %s\n", err.Error())
	} else {
		log.Infof("Logrus using Infof and subject key (level)")
		time.Sleep(time.Second)
	}

	// try setting key to current time in seconds

	config.KafkaProducerCfg.Key = logger.TimeSecondKey
	log, err = logger.NewLogger(config)
	if err != nil {
		fmt.Printf("Could not instantiate logrus logger: %s\n", err.Error())
	} else {
		log.Infof("Logrus using Infof and seconds key")
		time.Sleep(time.Second)
	}

	// try setting key to current time in nanoseconds

	config.KafkaProducerCfg.Key = logger.TimeNanoSecondKey
	log, err = logger.NewLogger(config)
	if err != nil {
		fmt.Printf("Could not instantiate logrus logger: %s\n", err.Error())
	} else {
		log.Infof("Logrus using Infof and nanoseconds key")
		time.Sleep(time.Second)
	}
}
