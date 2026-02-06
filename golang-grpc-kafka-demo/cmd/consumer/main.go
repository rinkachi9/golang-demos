package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"

	"github.com/segmentio/kafka-go"
)

func main() {
	// Configuration
	topic := "sensor-data"
	groupID := "sensor-consumer-group"
	brokerAddress := "localhost:9092"

	// Create a new reader with the configured brokers and topic
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  []string{brokerAddress},
		GroupID:  groupID,
		Topic:    topic,
		MinBytes: 10e3, // 10KB
		MaxBytes: 10e6, // 10MB
	})
	defer r.Close()

	// Context for graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	log.Println("Starting Kafka Consumer...")

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			m, err := r.FetchMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					// Context canceled, exit loop
					return
				}
				log.Printf("Error fetching message: %v", err)
				continue
			}

			// Process the message (Simulate complex calculation)
			processMessage(m)

			// Update the reader's offset commits
			if err := r.CommitMessages(ctx, m); err != nil {
				log.Printf("failed to commit messages: %v", err)
			}
		}
	}()

	<-ctx.Done()
	log.Println("Shutting down consumer...")
	wg.Wait()
}

func processMessage(m kafka.Message) {
	fmt.Printf("Message at topic/partition/offset %v/%v/%v: %s = %s\n", m.Topic, m.Partition, m.Offset, string(m.Key), string(m.Value))
}
