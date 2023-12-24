package location

import (
	"bytes"
	"driver_service/internal/model"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func GetLocationDrivers(lat, lng, radius float32) ([]model.Driver, error) {
	url := fmt.Sprintf("http://%s/drivers", "localhost:8081")
	reqBody := map[string]float32{"lat": lat, "lng": lng, "radius": radius}
	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Failed to get drivers. Status code: %d", resp.StatusCode)
	}

	var drivers []model.Driver
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &drivers)
	if err != nil {
		return nil, err
	}

	return drivers, nil
}
