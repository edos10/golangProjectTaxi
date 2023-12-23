package usecase

import (
	"driver_service/internal/constants"
	"driver_service/internal/model"
	"driver_service/internal/repository"
)

type TripUsecase interface {
	GetTrips() ([]model.Trip, error)
	GetTripByID(tripId string) (*model.Trip, error)
	StartTrip(tripId string) error
	AcceptTrip(tripId string, userId string) error
	EndTrip(tripId string) error
	CancelTrip(tripId string) error
	GetTripsByStatus(status string) ([]model.Trip, error)
	CheckFreeTrip(tripID string) bool
	CheckHaveTripByUser(tripID string, userID string) (bool, error)
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

func (u *tripUsecase) StartTrip(tripId string) error {
	return u.tripRepository.ChangeTripStatusById(tripId, constants.STARTED)
}

func (u *tripUsecase) EndTrip(tripId string) error {
	return u.tripRepository.ChangeTripStatusById(tripId, constants.ENDED)
}

func (u *tripUsecase) CancelTrip(tripId string) error {
	return u.tripRepository.ChangeTripStatusById(tripId, constants.CANCELED)
}

func (u *tripUsecase) AcceptTrip(tripId string, userId string) error {
	return u.tripRepository.AcceptTripById(tripId, userId)
}

func (u *tripUsecase) GetTripsByStatus(status string) ([]model.Trip, error) {
	return u.tripRepository.GetTripsByStatus(status)
}

func (u *tripUsecase) CheckFreeTrip(tripID string) bool {
	return u.tripRepository.CheckFreeTrip(tripID)
}

func (u *tripUsecase) CheckHaveTripByUser(tripID string, userID string) (bool, error) {
	return u.tripRepository.CheckHaveTripByUser(tripID, userID)
}
