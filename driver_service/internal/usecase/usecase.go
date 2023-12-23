package usecase

import (
	"driver_service/internal/model"
	"driver_service/internal/repository"
)

type TripUsecase interface {
	GetTrips() ([]model.Trip, error)
	GetTripByID(tripId string) (*model.Trip, error)
	StartTrip(tripId string, state string) error
	AcceptTrip(tripId string, state string) error
	GetTripsByStatus(status string) ([]model.Trip, error)
}

type tripUsecase struct {
	tripRepository repository.TripRepository
}

func NewTripUsecase(tripRepository repository.TripRepository) TripUsecase {
	return &tripUsecase{
		tripRepository: tripRepository,
	}
}

func (uc *tripUsecase) GetTrips() ([]model.Trip, error) {
	return uc.tripRepository.GetTrips()
}

func (u *tripUsecase) GetTripByID(tripID string) (*model.Trip, error) {
	return u.tripRepository.GetTripByID(tripID)
}

func (u *tripUsecase) StartTrip(tripId string, state string) error {
	return u.tripRepository.ChangeTripStatusById(tripId, "active")
}

func (u *tripUsecase) EndTrip(tripId string, state string) error {
	return u.tripRepository.ChangeTripStatusById(tripId, "ended")
}

func (u *tripUsecase) CancelTrip(tripId string, state string) error {
	return u.tripRepository.ChangeTripStatusById(tripId, "canceled")
}

func (u *tripUsecase) AcceptTrip(tripId string, state string) error {
	return u.tripRepository.ChangeTripStatusById(tripId, "accepted")
}

func (u *tripUsecase) GetTripsByStatus(status string) ([]model.Trip, error) {
	return u.tripRepository.GetTripsByStatus(status)
}
