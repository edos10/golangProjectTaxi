package http_handler

import (
	"driver_service/internal/constants"
	"driver_service/internal/controller/kafka_handler"
	"driver_service/internal/location"
	"driver_service/internal/model"
	"driver_service/internal/usecase"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

type HttpHandler struct {
	tripUsecase   usecase.TripUsecase
	kafkaProducer *kafka.Writer
	logger        *zap.Logger
}

func NewHttpHandler(tripUsecase usecase.TripUsecase, writer *kafka.Writer, logger *zap.Logger) *HttpHandler {
	return &HttpHandler{
		tripUsecase:   tripUsecase,
		kafkaProducer: writer,
		logger:        logger,
	}
}

func (h *HttpHandler) GetTrips(w http.ResponseWriter, r *http.Request) {
	driverID := r.Header.Get("user_id")
	if driverID == "" {
		h.logger.Error("Driver ID not found in the request header", zap.String("error_type", "DriverIDNotFound"))
		http.Error(w, "Driver ID not found in the request header", http.StatusUnauthorized)
		return
	}

	_, err := h.tripUsecase.GetTrips()
	if err != nil {
		h.logger.Error("Error getting trips", zap.Error(err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	trips := make([]model.Trip, 0)

	for i := 0; i < 7; i++ {
		res, _ := h.tripUsecase.GetTripsByStatus(constants.DRIVER_SEARCH)

		trips = append(trips, res...)
		time.Sleep(1 * time.Second)
	}

	tripsAvailableForDriver := make([]model.Trip, 0)

	for i := 0; i < len(trips); i++ {
		// получим список всех водителей по параметрам поездки
		lat, lng, radius := trips[i].From.Lat, trips[i].From.Lng, constants.RADIUS
		drivers, err := location.GetLocationDrivers(lat, lng, radius)
		if err != nil {
			h.logger.Error("Error getting location drivers", zap.Error(err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// смотрим, можно ли отправить поездку данному водителю как ту, которую он может взять

		// здесь можно такие поездки, взятые водителем уже или законченные, просто отдавать в другом роуте
		// или как вариант модификации, отдавать просто все поездки, а на фронте бить их на 2 вида статусов: DRIVER_SEARCH(то есть предлагаемые водителю) и законченные
		// я пока сделаю так, чтобы были только предлагаемые

		if trips[i].Status != constants.DRIVER_SEARCH {
			continue
		}

		checkDriverAvailable := false
		for j := 0; j < len(drivers); j++ {
			// этому водителю подходит эта поездка по радиусу и она в статусе поиска водителя
			if drivers[j].ID == driverID {
				checkDriverAvailable = true
				break
			}
		}
		if checkDriverAvailable {
			tripsAvailableForDriver = append(tripsAvailableForDriver, trips[i])
		}

	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tripsAvailableForDriver); err != nil {
		h.logger.Error("Error encoding JSON response", zap.Error(err))
		http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
		return
	}
	h.logger.Info("GetTrips request processed successfully", zap.Int("trips_count", len(tripsAvailableForDriver)))
}

func (h *HttpHandler) GetTripByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tripID := vars["trip_id"]
	userID := r.Header.Get("user_id")

	if tripID == "" {
		h.logger.Error("trip_id is required", zap.String("user_id", userID))
		http.Error(w, "trip_id is required", http.StatusBadRequest)
		return
	}

	isHave, err := h.tripUsecase.CheckHaveTripByUser(tripID, userID)
	if err != nil {
		h.logger.Error("Error checking if the user has the trip", zap.Error(err), zap.String("user_id", userID), zap.String("trip_id", tripID))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !isHave {
		h.logger.Error("This user hasn't got this trip", zap.String("user_id", userID), zap.String("trip_id", tripID))
		http.Error(w, "this user haven't got a this trip", http.StatusBadRequest)
		return
	}

	trip, err := h.tripUsecase.GetTripByID(tripID)

	if err != nil {
		h.logger.Error("Error getting trip by ID", zap.Error(err), zap.String("user_id", userID), zap.String("trip_id", tripID))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(trip); err != nil {
		h.logger.Error("Error encoding JSON response", zap.Error(err), zap.String("user_id", userID), zap.String("trip_id", tripID))
		http.Error(w, "Error encoding JSON response", http.StatusInternalServerError)
		return
	}
	h.logger.Info("GetTripByID request processed successfully", zap.String("user_id", userID), zap.String("trip_id", tripID))
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
		h.logger.Error("trip_id is required", zap.String("user_id", userID))
		http.Error(w, "trip_id is required", http.StatusBadRequest)
		return
	}

	_, errGet := h.tripUsecase.GetTripByID(tripID)

	if errGet != nil {
		h.logger.Error("Error getting trip by ID", zap.Error(errGet), zap.String("user_id", userID), zap.String("trip_id", tripID))
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("This trip doesn't exists"))
		return
	}

	isHave, err := h.tripUsecase.CheckHaveTripByUser(tripID, userID)
	if err != nil {
		h.logger.Error("Error checking if the user has the trip", zap.Error(err), zap.String("user_id", userID), zap.String("trip_id", tripID))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !isHave {
		h.logger.Error("This user hasn't got this trip", zap.String("user_id", userID), zap.String("trip_id", tripID))
		http.Error(w, "this user hasn't got this trip", http.StatusBadRequest)
		return
	}

	err = h.tripUsecase.StartTrip(tripID)
	if err != nil {
		h.logger.Error("Error starting trip", zap.Error(err), zap.String("user_id", userID), zap.String("trip_id", tripID))
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
		h.logger.Error("Error marshaling started trip message", zap.Error(err), zap.String("user_id", userID), zap.String("trip_id", tripID))
		http.Error(w, "Error marshaling started trip message: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = kafka_handler.SendMessage(h.kafkaProducer, messageBytes, "STARTED")
	if err != nil {
		h.logger.Error("Error sending started trip message to Kafka", zap.Error(err), zap.String("user_id", userID), zap.String("trip_id", tripID))
		http.Error(w, "Error sending started trip message to Kafka: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("Trip started successfully", zap.String("user_id", userID), zap.String("trip_id", tripID))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Trip started successfully"))
}

func (h *HttpHandler) EndTrip(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tripID := vars["trip_id"]
	userID := r.Header.Get("user_id")

	if tripID == "" {
		h.logger.Error("trip_id is required", zap.String("user_id", userID))
		http.Error(w, "trip_id is required", http.StatusBadRequest)
		return
	}

	trip, errGet := h.tripUsecase.GetTripByID(tripID)

	if errGet != nil {
		h.logger.Error("Error getting trip by ID", zap.Error(errGet), zap.String("user_id", userID), zap.String("trip_id", tripID))
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("This trip doesn't exists"))
		return
	}

	isHave, err := h.tripUsecase.CheckHaveTripByUser(tripID, userID)
	if err != nil {
		h.logger.Error("Error checking if the user has the trip", zap.Error(err), zap.String("user_id", userID), zap.String("trip_id", tripID))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !isHave {
		h.logger.Error("This user hasn't got this trip", zap.String("user_id", userID), zap.String("trip_id", tripID))
		http.Error(w, "this user hasn't got this trip", http.StatusBadRequest)
		return
	}

	if trip.Status != constants.STARTED {
		h.logger.Error("Cannot end trip that is not in STARTED state", zap.String("user_id", userID), zap.String("trip_id", tripID))
		http.Error(w, "Cannot end trip that is not in STARTED state", http.StatusBadRequest)
		return
	}

	err = h.tripUsecase.EndTrip(tripID)
	if err != nil {
		h.logger.Error("Error ending trip", zap.Error(err), zap.String("user_id", userID), zap.String("trip_id", tripID))
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
		h.logger.Error("Error marshaling ended trip message", zap.Error(err), zap.String("user_id", userID), zap.String("trip_id", tripID))
		http.Error(w, "Error marshaling ended trip message: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = kafka_handler.SendMessage(h.kafkaProducer, messageBytes, "ENDED")
	if err != nil {
		h.logger.Error("Error sending ended trip message to Kafka", zap.Error(err), zap.String("user_id", userID), zap.String("trip_id", tripID))
		http.Error(w, "Error sending ended trip message to Kafka: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("Trip ended successfully", zap.String("user_id", userID), zap.String("trip_id", tripID))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Trip ended successfully"))
}

func (h *HttpHandler) CancelTrip(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tripID := vars["trip_id"]
	userID := r.Header.Get("user_id")

	if tripID == "" {
		h.logger.Error("trip_id is required", zap.String("user_id", userID))
		http.Error(w, "trip_id is required", http.StatusBadRequest)
		return
	}

	trip, errGet := h.tripUsecase.GetTripByID(tripID)

	if errGet != nil {
		h.logger.Error("Error getting trip by ID", zap.Error(errGet), zap.String("user_id", userID), zap.String("trip_id", tripID))
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("This trip doesn't exists"))
		return
	}

	isHave, err := h.tripUsecase.CheckHaveTripByUser(tripID, userID)
	if err != nil {
		h.logger.Error("Error checking if the user has the trip", zap.Error(err), zap.String("user_id", userID), zap.String("trip_id", tripID))
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !isHave {
		h.logger.Error("This user hasn't got this trip", zap.String("user_id", userID), zap.String("trip_id", tripID))
		http.Error(w, "this user hasn't got this trip", http.StatusBadRequest)
		return
	}

	if trip.Status != constants.STARTED {
		h.logger.Error("Cannot cancel trip that is not in STARTED state", zap.String("user_id", userID), zap.String("trip_id", tripID))
		http.Error(w, "Cannot cancel trip that is not in STARTED state", http.StatusBadRequest)
		return
	}

	err = h.tripUsecase.CancelTrip(tripID)
	if err != nil {
		h.logger.Error("Error canceling trip", zap.Error(err), zap.String("user_id", userID), zap.String("trip_id", tripID))
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
		h.logger.Error("Error marshaling cancel trip message", zap.Error(err), zap.String("user_id", userID), zap.String("trip_id", tripID))
		http.Error(w, "Error marshaling started trip message: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = kafka_handler.SendMessage(h.kafkaProducer, messageBytes, "CANCELED")
	if err != nil {
		h.logger.Error("Error sending cancel trip message to Kafka", zap.Error(err), zap.String("user_id", userID), zap.String("trip_id", tripID))
		http.Error(w, "Error sending started trip message to Kafka: "+err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("Trip canceled successfully", zap.String("user_id", userID), zap.String("trip_id", tripID))
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Trip canceled successfully"))
}

// to fix locate in other package
// fixed
