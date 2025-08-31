package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/segmentio/kafka-go"
)

func main() {
	topic := "orders"
	partition := 0

	conn, err := kafka.DialLeader(context.Background(), "tcp", "localhost:9092", topic, partition)
	if err != nil {
		log.Fatalf("failed to dial leader: %v", err)
	}

	jsonData, err := os.ReadFile("./model.json")
	if err != nil {
		log.Fatalf("failed to read model file: %v", err)
	}

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_, err = conn.WriteMessages(
		kafka.Message{Value: jsonData},
	)
	if err != nil {
		log.Fatalf("failed to write messages: %v", err)
	}

	if err := conn.Close(); err != nil {
		log.Fatalf("failed to close connection: %v", err)
	}
	log.Println("Message sent successfully!")
}
