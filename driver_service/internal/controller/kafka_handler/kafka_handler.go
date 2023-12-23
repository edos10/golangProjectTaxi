package kafka_handler

import (
	"context"
	"driver_service/internal/constants"
	"driver_service/internal/model"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/mongo"
)

type KafkaProducer struct {
	// indexes writers and topics for send in producer
	// 0 - driver_search
	// 1 - driver found
	// 2 - on position
	// 3 - started
	// 4 - ended
	// 5 - canceled
	writer *kafka.Writer
}

type KafkaConsumer struct {
	reader   *kafka.Reader
	database *mongo.Database
}

var async = flag.Bool("a", false, "use async")

func NewKafkaProducer(brokers []string, topic string) *kafka.Writer {
	logger := log.Default()
	//[]string{"127.0.0.1:29092"},

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:     brokers,
		Topic:       topic,
		Async:       *async,
		Logger:      kafka.LoggerFunc(logger.Printf),
		ErrorLogger: kafka.LoggerFunc(logger.Printf),
		BatchSize:   2000,

		// CompressionCodec: &compress.Lz4Codec,

		// Balancer: &SimpleBalancer{},
	})

	return writer
}

func NewKafkaConsumer(brokers []string, topic string, db *mongo.Database) *KafkaConsumer {
	return &KafkaConsumer{
		reader: kafka.NewReader(kafka.ReaderConfig{
			Brokers: brokers,
			Topic:   topic,
		}),
		database: db,
	}
}

func (k *KafkaConsumer) Consume() {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        []string{"127.0.0.1:29092"},
		Topic:          constants.DRIVER_SEARCH,
		GroupID:        "my-group",
		SessionTimeout: time.Second * 6,
	})
	defer reader.Close()

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			fmt.Println("Error fetching message:", err)
			continue
		}
		fmt.Println(msg.Value)
		var trip model.KafkaMessageToDriver
		err = json.Unmarshal(msg.Value, &trip)
		if err != nil {
			fmt.Println("Error unmarshalling trip:", err)
			continue
		}
		fmt.Println(trip)
		k.writeToMongo(trip.Data)

		err = k.reader.CommitMessages(context.Background(), msg)
		if err != nil {
			fmt.Println("Error committing message:", err)
		}

		time.Sleep(300 * time.Millisecond)
	}
}

func (k *KafkaConsumer) Close() {
	k.reader.Close()
}

func (k *KafkaConsumer) writeToMongo(trip model.Trip) error {
	collection := k.database.Collection("trips")
	trip.Status = constants.DRIVER_SEARCH
	fmt.Println(trip.ID)
	_, err := collection.InsertOne(context.Background(), trip)
	fmt.Println(err)
	return err
}

func SendMessage(writer *kafka.Writer, body []byte, key string) error {
	ctx := context.Background()
	if writer == nil {
		return errors.New("kafka writer is nil")
	}

	errSend := writer.WriteMessages(ctx, kafka.Message{Key: []byte(key), Value: body})
	return errSend
}
