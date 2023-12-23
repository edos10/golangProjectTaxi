package repository

import (
	"context"
	"driver_service/internal/constants"
	"driver_service/internal/model"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type TripRepository interface {
	GetTrips() ([]model.Trip, error)
	GetTripByID(tripID string) (*model.Trip, error)
	ChangeTripStatusById(tripId string, state string) error
	AcceptTripById(tripID string, userID string) error
	GetTripsByStatus(status string) ([]model.Trip, error)
	CheckFreeTrip(tripID string) bool
	CheckHaveTripByUser(tripID string, userID string) (bool, error)
}

type tripRepository struct {
	client     *mongo.Client
	Database   *mongo.Database
	collection *mongo.Collection
}

func NewTripRepository() *tripRepository {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}

	database := client.Database("taxi")
	collection := database.Collection("trips")

	return &tripRepository{
		client:     client,
		Database:   database,
		collection: collection,
	}
}

func (r *tripRepository) GetTrips() ([]model.Trip, error) {
	var trips []model.Trip
	cursor, err := r.collection.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var trip model.Trip
		if err := cursor.Decode(&trip); err != nil {
			return nil, err
		}
		trips = append(trips, trip)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return trips, nil
}

func (r *tripRepository) GetTripByID(tripID string) (*model.Trip, error) {
	var trip model.Trip
	err := r.collection.FindOne(context.Background(), bson.M{"_id": tripID}).Decode(&trip)
	if err != nil {
		return nil, err
	}

	return &trip, nil
}

func (r *tripRepository) ChangeTripStatusById(tripId string, state string) error {
	filter := bson.M{"_id": tripId}
	update := bson.M{"$set": bson.M{"status": state}}

	_, err := r.collection.UpdateOne(context.Background(), filter, update)
	return err
}

func (r *tripRepository) GetTripsByStatus(status string) ([]model.Trip, error) {
	var trips []model.Trip

	filter := bson.M{"status": status}

	cursor, err := r.collection.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	for cursor.Next(context.Background()) {
		var trip model.Trip
		if err := cursor.Decode(&trip); err != nil {
			return nil, err
		}
		trips = append(trips, trip)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return trips, nil

}

func (r *tripRepository) CheckFreeTrip(tripID string) bool {
	trip, _ := r.GetTripByID(tripID)
	return trip.Status == constants.DRIVER_SEARCH
}

func (r *tripRepository) AcceptTripById(tripId string, userId string) error {
	filter := bson.M{"_id": tripId}
	update := bson.M{"$set": bson.M{"status": constants.DRIVER_FOUND, "driver_id": userId}}

	_, err := r.collection.UpdateOne(context.Background(), filter, update)
	return err
}

func (r *tripRepository) CheckHaveTripByUser(tripID string, userID string) (bool, error) {
	filter := bson.M{"_id": tripID, "driver_id": userID}

	count, err := r.collection.CountDocuments(context.Background(), filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
