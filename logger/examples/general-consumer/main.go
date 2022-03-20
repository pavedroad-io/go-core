package main

import (
	"fmt"
	"os"
	"os/signal"

	"github.com/Shopify/sarama"
)

var (
	brokers        = []string{"localhost:9092"}
	config         = sarama.NewConfig()
)

func main() {

	// init master consumer
	master, err := sarama.NewConsumer(brokers, config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get consumer: %s\n", err.Error())
		os.Exit(1)
	}
	defer master.Close()
	// get topics
	topics, err := master.Topics()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get topics: %s\n", err.Error())
		os.Exit(1)
	}
	// init exit
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// consume all topics
	for _, topic := range topics {
		// get partitions
		partitions, err := master.Partitions(topic)
		if err != nil {
			fmt.Fprintf(os.Stderr,
				"Failed to get partitions for topic %s: %s\n",
				topic, err.Error())
			os.Exit(1)
		}
		// consume all partitions
		for _, partition := range partitions {
			// init partition consumer
			consumer, err := master.ConsumePartition(topic, partition,
				sarama.OffsetOldest)
			if err != nil {
				fmt.Fprintf(os.Stderr,
					"Failed to get consumer for topic %s partition %d: %s\n",
					topic, partition, err.Error())
				os.Exit(1)
			}
			defer consumer.Close()
			// consume log messages
			go func() {
				for {
					select {
					case msg := <-consumer.Messages():
						fmt.Printf("T:%s P:%d K:%s V:%s\n\n",
							msg.Topic, msg.Partition, msg.Key, msg.Value)
					case err := <-consumer.Errors():
						fmt.Fprintf(os.Stderr, "Error T:%s P:%d %s\n",
							topic, partition, err)
					case <-interrupt:
						fmt.Printf("\nExiting consumer\n")
						return
					}
				}
			}()
		}
	}
}
