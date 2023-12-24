package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

var async = flag.Bool("a", false, "use async")

type Trip struct {
	ID       string        `json:"id" bson:"_id,omitempty"`
	DriverID string        `json:"driver_id" bson:"driver_id"`
	OfferID  string        `json:"offer_id" bson:"offer_id"`
	From     LatLngLiteral `json:"from" bson:"from"`
	To       LatLngLiteral `json:"to" bson:"to"`
	Price    Money         `json:"price" bson:"price"`
	Status   string        `json:"status" bson:"status"`
}

type LatLngLiteral struct {
	Lat float64 `json:"lat" bson:"lat"`
	Lng float64 `json:"lng" bson:"lng"`
}

type Money struct {
	Amount   float64 `json:"amount" bson:"amount"`
	Currency string  `json:"currency" bson:"currency"`
}

type Driver struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type TripMessage struct {
	TripId string `json:"trip_id"`
}

type KafkaMessage struct {
	ID              string `json:"id"`
	Source          string `json:"source"`
	Type            string `json:"type"`
	DataContentType string `json:"datacontenttype"`
	Time            string `json:"time"`
	Data            Trip   `json:"data"`
}

func main() {
	flag.Parse()

	ctx := context.Background()

	logger := log.Default()

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:     []string{"127.0.0.1:29092", "127.0.0.1:39092", "127.0.0.1:49092"},
		Topic:       "DRIVER_SEARCH",
		Async:       *async,
		Logger:      kafka.LoggerFunc(logger.Printf),
		ErrorLogger: kafka.LoggerFunc(logger.Printf),
		BatchSize:   2000,
	})

	defer writer.Close()

	tripMessage := KafkaMessage{
		ID:              "Rzozcr4cr4kcmirj4cmi4c",
		Source:          "/trip",
		Type:            "trip.event.created",
		DataContentType: "application/json",
		Time:            time.Now().String(),
		Data: Trip{
			ID:      "abc7",
			OfferID: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6InN0cmluZyIsImZyb20iOnsibGF0IjowLCJsbmciOjB9LCJ0byI6eyJsYXQiOjAsImxuZyI6MH0sImNsaWVudF9pZCI6InN0cmluZyIsInByaWNlIjp7ImFtb3VudCI6OTkuOTUsImN1cnJlbmN5IjoiUlVCIn19.fg0Bv2ONjT4r8OgFqJ2tpv67ar7pUih2LhDRCRhWW3c",
			Price: Money{
				Currency: "RUB",
				Amount:   100,
			},
			Status: "DRIVER_SEARCH",
			From: LatLngLiteral{
				Lat: 0,
				Lng: 0,
			},
			To: LatLngLiteral{
				Lat: 0,
				Lng: 0,
			},
		},
	}

	messageBytes, err := json.Marshal(tripMessage)
	if err != nil {
		log.Fatal(err)
	}
	for {
		err = writer.WriteMessages(ctx, kafka.Message{Key: []byte("CREATED"), Value: messageBytes})
		if err != nil {
			log.Fatal(err)
		}
	}
}
