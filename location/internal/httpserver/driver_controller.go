package httpserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/google/uuid"
	"github.com/location/internal/services"
)

type driverController struct {
	service *service.LocationService
	writer *HttpWriter
}

func (controller *driverController) GetDrivers(w http.ResponseWriter, r *http.Request, params GetDriversParams) {
	modelDrivers, err := controller.service.GetDriversInRadius(r.Context(), params.Lat, params.Lng, params.Radius)
	if err != nil {
		err = fmt.Errorf("Internal Server Error: %w", err)
		controller.writer.writeError(w, http.StatusInternalServerError, err)
		return
	}

	if len(*modelDrivers) == 0 {
		controller.writer.writeResponse(w, http.StatusNotFound, "Not found")
	}
	
	var drivers []Driver
	for _, modelDriver := range *modelDrivers {
		driver := driverToDto(modelDriver)
		drivers = append(drivers, driver)
	}

	controller.writer.writeJSONResponse(w, http.StatusOK, drivers)
}

func (controller *driverController) UpdateDriverLocation(w http.ResponseWriter, r *http.Request, driverId string) {
	var location LatLngLiteral

	if err := json.NewDecoder(r.Body).Decode(&location); err != nil {
		err = fmt.Errorf("Invalid Body: %w", err)
		controller.writer.writeError(w, http.StatusBadRequest, err)
		return
	}

	guid, err := uuid.Parse(driverId)
	if err != nil {
		err = fmt.Errorf("Invalid Driver Id: %w", err)
		controller.writer.writeError(w, http.StatusBadRequest, err)
		return
	}
	
	existsCheck, err := controller.service.UpdateDriverLocation(r.Context(), guid, locationFromDto(location))
	if existsCheck {
		controller.writer.writeError(w, http.StatusNotFound, err)
		return
	}
	if err != nil {
		err = fmt.Errorf("Internal Server Error: %w", err)
		controller.writer.writeError(w, http.StatusInternalServerError, err)
		return
	}
}

func NewServerInterface(service *service.LocationService, writer *HttpWriter) ServerInterface {
	
	return &driverController{
		service: service,
		writer: writer,
	}
}
