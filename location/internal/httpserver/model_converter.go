package httpserver

import (
	"github.com/location/internal/model"
)

func driverToDto(driver model.Driver) Driver {

	guid := driver.ID.String()
	return Driver{
		Id: &guid,
		LatLngLiteral: locationToDto(driver.Location),
	}
}

func locationToDto(location model.Location) LatLngLiteral {
	
	return LatLngLiteral {
		Lat: location.Lat,
		Lng: location.Lng,
	}
}

func locationFromDto(location LatLngLiteral) model.Location {
	
	return model.Location {
		Lat: location.Lat,
		Lng: location.Lng,
	}
}

