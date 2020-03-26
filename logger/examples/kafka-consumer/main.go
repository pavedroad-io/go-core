package main

import (
	"fmt"
	// stdlog "log"
	"os"
	"os/signal"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
)

var (
	brokers = []string{"localhost:9092"}
	group   = "mygroup"
	topics  = []string{"logs"}
	config  = cluster.NewConfig()
)

// Use sarama-cluster to consume logs

func main() {

	// init config
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	// init consumer
	consumer, err := cluster.NewConsumer(brokers, group, topics, config)
	if err != nil {
		panic(err)
	}
	// uncomment to enable connection debugging
	// sarama.Logger = stdlog.New(os.Stdout, "[sarama] ", stdlog.LstdFlags)

	// init exit
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	defer consumer.Close()

	// consume log messages
	for {
		select {
		case msg := <-consumer.Messages():
			fmt.Printf("P:%d K:%s V:%s\n\n", msg.Partition, msg.Key, msg.Value)
			consumer.MarkOffset(msg, "kafka-test")
		case err := <-consumer.Errors():
			fmt.Fprintf(os.Stderr, "Message error: %s\n", err)
		case <-interrupt:
			fmt.Printf("\nExiting consumer\n")
			return
		}
	}
}
