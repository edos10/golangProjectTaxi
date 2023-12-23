package http_handler

import (
	"driver_service/internal/constants"
	"driver_service/internal/controller/kafka_handler"
	"driver_service/internal/model"
	"driver_service/internal/usecase"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type HttpHandler struct {
	tripUsecase   usecase.TripUsecase
	kafkaProducer *kafka_handler.KafkaProducer
	kafkaConsumer *kafka_handler.KafkaConsumer
}

func NewHttpHandler(tripUsecase usecase.TripUsecase) *HttpHandler {
	return &HttpHandler{
		tripUsecase: tripUsecase,
	}
}

func (h *HttpHandler) GetTrips(w http.ResponseWriter, r *http.Request) {
	driverID := r.Header.Get("user_id")
	if driverID == "" {
		http.Error(w, "Driver ID not found in the request header", http.StatusUnauthorized)
		return
	}
	fmt.Println("before first")

	_, err := h.tripUsecase.GetTrips()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}
	fmt.Println("after first")

	trips := make([]model.Trip, 0)
	for i := 0; i < 7; i++ {
		fmt.Println("before cycle")
		res, _ := h.tripUsecase.GetTripsByStatus(constants.DRIVER_SEARCH)
		trips = append(trips, res...)
		fmt.Println(i)
		time.Sleep(1 * time.Second)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trips)

}

func (h *HttpHandler) GetTripByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tripID := vars["trip_id"]
	if tripID == "" {
		http.Error(w, "trip_id is required", http.StatusBadRequest)
		return
	}
	fmt.Println(tripID)
	trip, err := h.tripUsecase.GetTripByID(tripID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trip)
}

func (h *HttpHandler) AcceptTrip(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	tripID := vars["trip_id"]
	userID := r.Header.Get("user_id")

	fmt.Println(tripID, userID)
	_, errGet := h.tripUsecase.GetTripByID(tripID)

	if errGet != nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	err := h.tripUsecase.AcceptTrip(tripID, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error accepting trip: " + err.Error()))
		return
	}

	message := model.KafkaMessage{
		ID:              tripID,
		Source:          "/driver",
		Type:            "trip.command.accept",
		DataContentType: "application/json",
		Time:            time.Now().Format(time.RFC822),
		Data: model.Trip{
			ID:       tripID,
			DriverID: userID,
		},
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error marshaling accepted trip message: " + err.Error()))
		return
	}

	err = h.kafkaProducer.SendMessage(messageBytes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error sending accepted trip message to Kafka: " + err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Trip accepted successfully"))
}

func (h *HttpHandler) StartTrip(w http.ResponseWriter, r *http.Request) {

}

func (h *HttpHandler) EndTrip(w http.ResponseWriter, r *http.Request) {

}

func (h *HttpHandler) CancelTrip(w http.ResponseWriter, r *http.Request) {

}

func getLocationDrivers(driverID string) ([]model.Driver, error) {
	return []model.Driver{
		{ID: "1", Name: "Driver 1", Latitude: 40.7128, Longitude: -74.0060},
		{ID: "2", Name: "Driver 2", Latitude: 34.0522, Longitude: -118.2437},
	}, nil
}
