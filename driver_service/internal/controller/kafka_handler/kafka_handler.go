package kafka_handler

import (
	"context"
	"driver_service/internal/constants"
	"driver_service/internal/model"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
	"go.mongodb.org/mongo-driver/mongo"
)

type KafkaProducer struct {
	writer *kafka.Writer
}

type KafkaConsumer struct {
	reader   *kafka.Reader
	database *mongo.Database
}

var async = flag.Bool("a", false, "use async")

func NewKafkaProducer(brokers []string, topic string) *KafkaProducer {
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

	return &KafkaProducer{
		writer: writer,
	}
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
		Topic:          "DRIVER_SEARCH",
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
		var trip model.KafkaMessage
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

func (k *KafkaProducer) SendMessage(body []byte) error {
	ctx := context.Background()
	errSend := k.writer.WriteMessages(ctx, kafka.Message{Key: []byte(k.writer.Topic), Value: body})
	return errSend
}
