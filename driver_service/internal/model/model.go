package model

type Trip struct {
	ID       string        `json:"id" bson:"_id,omitempty"`
	DriverID string        `json:"driver_id" bson:"driver_id"`
	From     LatLngLiteral `json:"from" bson:"from"`
	To       LatLngLiteral `json:"to" bson:"to"`
	Price    Money         `json:"price" bson:"price"`
	Status   string        `json:"status" bson:"status"`
}

type LatLngLiteral struct {
	Lat float64 `json:"lat" bson:"lat"`
	Lng float64 `json:"lng" bson:"lng"`
}

type Money struct {
	Amount   float64 `json:"amount" bson:"amount"`
	Currency string  `json:"currency" bson:"currency"`
}

type Driver struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type TripMessage struct {
	TripId string `json:"trip_id"`
}

type KafkaMessage struct {
	ID              string `json:"id"`
	Source          string `json:"source"`
	Type            string `json:"type"`
	DataContentType string `json:"datacontenttype"`
	Time            string `json:"time"`
	Data            Trip   `json:"data"`
}
