package main

import (
	"fmt"
	// stdlog "log"
	"os"
	"os/signal"

	"github.com/Shopify/sarama"
)

var (
	brokers        = []string{"localhost:9092", "192.168.64.2:9092"}
	topic   string = "logs"
	config         = sarama.NewConfig()
)

func main() {

	// init consumer
	master, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get consumer: %s\n", err.Error())
		os.Exit(1)
	}
	// init partition 0 consumer
	consumer, err := master.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get partition consumer: %s\n",
			err.Error())
		os.Exit(1)
	}
	// uncomment to enable connection debugging
	// sarama.Logger = stdlog.New(os.Stdout, "[sarama] ", stdlog.LstdFlags)

	// init exit
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	defer consumer.Close()
	defer master.Close()

	// consume log messages
	for {
		select {
		case msg := <-consumer.Messages():
			fmt.Printf("P:%d K:%s V:%s\n\n", msg.Partition, msg.Key, msg.Value)
		case err := <-consumer.Errors():
			fmt.Fprintf(os.Stderr, "Message error: %s\n", err)
		case <-interrupt:
			fmt.Printf("\nExiting consumer\n")
			return
		}
	}
}
