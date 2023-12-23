package http_handler

import (
	"driver_service/internal/constants"
	"driver_service/internal/controller/kafka_handler"
	"driver_service/internal/model"
	"driver_service/internal/usecase"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/segmentio/kafka-go"
)

type HttpHandler struct {
	tripUsecase   usecase.TripUsecase
	kafkaProducer *kafka.Writer
}

func NewHttpHandler(tripUsecase usecase.TripUsecase, writer *kafka.Writer) *HttpHandler {
	return &HttpHandler{
		tripUsecase:   tripUsecase,
		kafkaProducer: writer,
	}
}

func (h *HttpHandler) GetTrips(w http.ResponseWriter, r *http.Request) {
	driverID := r.Header.Get("user_id")
	if driverID == "" {
		http.Error(w, "Driver ID not found in the request header", http.StatusUnauthorized)
		return
	}

	_, err := h.tripUsecase.GetTrips()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)

		return
	}

	trips := make([]model.Trip, 0)
	for i := 0; i < 7; i++ {
		res, _ := h.tripUsecase.GetTripsByStatus(constants.DRIVER_SEARCH)
		trips = append(trips, res...)
		time.Sleep(1 * time.Second)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(trips)

}

func (h *HttpHandler) GetTripByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tripID := vars["trip_id"]
	userID := r.Header.Get("user_id")

	if tripID == "" {
		http.Error(w, "trip_id is required", http.StatusBadRequest)
		return
	}

	isHave, err := h.tripUsecase.CheckHaveTripByUser(tripID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !isHave {
		http.Error(w, "this user haven't got a this trip", http.StatusBadRequest)
		return
	}

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
	_, errGet := h.tripUsecase.GetTripByID(tripID)

	if errGet != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("This trip doesn't exists"))
		return
	}

	isFree := h.tripUsecase.CheckFreeTrip(tripID)
	if !isFree {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("This trip already taken by other driver!"))
		return
	}

	err := h.tripUsecase.AcceptTrip(tripID, userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error accepting trip: " + err.Error()))
		return
	}

	message := model.KafkaMessageToTrip{
		ID:              tripID,
		Source:          "/driver",
		Type:            "trip.command.accept",
		DataContentType: "application/json",
		Time:            time.Now().Format(time.RFC822),
		Data: model.SendTripData{
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

	err = kafka_handler.SendMessage(h.kafkaProducer, messageBytes, "ACCEPTED")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Error sending accepted trip message to Kafka: " + err.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Trip accepted successfully"))
}

func (h *HttpHandler) StartTrip(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tripID := vars["trip_id"]
	userID := r.Header.Get("user_id")

	if tripID == "" {
		http.Error(w, "trip_id is required", http.StatusBadRequest)
		return
	}

	isHave, err := h.tripUsecase.CheckHaveTripByUser(tripID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !isHave {
		http.Error(w, "this user hasn't got this trip", http.StatusBadRequest)
		return
	}

	trip, err := h.tripUsecase.GetTripByID(tripID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.tripUsecase.StartTrip(tripID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	message := model.KafkaMessageToTrip{
		ID:              tripID,
		Source:          "/driver",
		Type:            "trip.command.start",
		DataContentType: "application/json",
		Time:            time.Now().Format(time.RFC822),
		Data: model.SendTripData{
			ID:       tripID,
			DriverID: userID,
		},
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		http.Error(w, "Error marshaling started trip message: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = kafka_handler.SendMessage(h.kafkaProducer, messageBytes, "STARTED")
	if err != nil {
		http.Error(w, "Error sending started trip message to Kafka: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Trip started successfully"))
}

func (h *HttpHandler) EndTrip(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tripID := vars["trip_id"]
	userID := r.Header.Get("user_id")

	if tripID == "" {
		http.Error(w, "trip_id is required", http.StatusBadRequest)
		return
	}

	isHave, err := h.tripUsecase.CheckHaveTripByUser(tripID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !isHave {
		http.Error(w, "this user hasn't got this trip", http.StatusBadRequest)
		return
	}

	trip, err := h.tripUsecase.GetTripByID(tripID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.tripUsecase.EndTrip(tripID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	message := model.KafkaMessageToTrip{
		ID:              tripID,
		Source:          "/driver",
		Type:            "trip.command.end",
		DataContentType: "application/json",
		Time:            time.Now().Format(time.RFC822),
		Data: model.SendTripData{
			ID:       tripID,
			DriverID: userID,
		},
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		http.Error(w, "Error marshaling started trip message: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = kafka_handler.SendMessage(h.kafkaProducer, messageBytes, "ENDED")
	if err != nil {
		http.Error(w, "Error sending started trip message to Kafka: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Trip ended successfully"))
}

func (h *HttpHandler) CancelTrip(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tripID := vars["trip_id"]
	userID := r.Header.Get("user_id")

	if tripID == "" {
		http.Error(w, "trip_id is required", http.StatusBadRequest)
		return
	}

	isHave, err := h.tripUsecase.CheckHaveTripByUser(tripID, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !isHave {
		http.Error(w, "this user hasn't got this trip", http.StatusBadRequest)
		return
	}

	trip, err := h.tripUsecase.GetTripByID(tripID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.tripUsecase.StartTrip(tripID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	message := model.KafkaMessageToTrip{
		ID:              tripID,
		Source:          "/driver",
		Type:            "trip.command.cancel",
		DataContentType: "application/json",
		Time:            time.Now().Format(time.RFC822),
		Data: model.SendTripData{
			ID:       tripID,
			DriverID: userID,
		},
	}

	messageBytes, err := json.Marshal(message)
	if err != nil {
		http.Error(w, "Error marshaling started trip message: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = kafka_handler.SendMessage(h.kafkaProducer, messageBytes, "CANCELED")
	if err != nil {
		http.Error(w, "Error sending started trip message to Kafka: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Trip canceled successfully"))
}

// to fix locate in other package
// fixed
