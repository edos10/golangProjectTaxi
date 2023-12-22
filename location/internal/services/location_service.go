package service

import (
	"context"
	"math"
	"errors"
	"github.com/google/uuid"
	"github.com/location/internal/persistency"
	"github.com/location/internal/model"
)

type LocationService struct {
	driverRepo persistency.DriverRepo
}

func (service *LocationService) GetDriversInRadius(ctx context.Context, lat float32, lng float32, radius float32) (*[]model.Driver, error) {
	drivers, err := service.driverRepo.GetDrivers(ctx);
	if err != nil {
		return nil, err
	}
	
	var nearDrivers []model.Driver
	for _, driver := range *drivers {
		distance := distance(lat, lng, driver.Location.Lat, driver.Location.Lng)
		if distance <= radius {
			nearDrivers = append(nearDrivers, driver)
		}
	}

	return &nearDrivers, nil
}

func (service *LocationService) UpdateDriverLocation(ctx context.Context, driverId uuid.UUID, location model.Location) (bool, error) {
	exists, err := service.driverRepo.CheckDriverExistence(ctx, driverId)
	if err != nil {
		return false, err
	}

	if !exists {
		err = errors.New("Driver with such Id doesn't exist")
		return true, err
	}

	err =  service.driverRepo.UpdateDriverLocation(ctx, driverId, location)
	if err != nil {
		return false, err
	}

	return false, nil
}

func NewLocationService(repo persistency.DriverRepo) *LocationService {

	return &LocationService {
		driverRepo: repo,
	}
}

func distance(lat1 float32, lng1 float32, lat2 float32, lng2 float32) float32 {
	// Convert latitude and longitude from degrees to radians
	lat1Rad := float64(lat1) * (math.Pi / 180)
	lon1Rad := float64(lng1) * (math.Pi / 180)
	lat2Rad := float64(lat2) * (math.Pi / 180)
	lon2Rad := float64(lng2) * (math.Pi / 180)

	// Haversine formula
	dlat := lat2Rad - lat1Rad
	dlon := lon2Rad - lon1Rad
	a := math.Pow(math.Sin(dlat/2), 2) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Pow(math.Sin(dlon/2), 2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	// Radius of the Earth in kilometers
	earthRadius := 6371.0

	// Calculate the distance
	distance := earthRadius * c
	
	return float32(distance)
}