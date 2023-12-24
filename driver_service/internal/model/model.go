package model

type Trip struct {
	ID       string        `json:"id" bson:"_id,omitempty"`
	DriverID string        `json:"driver_id" bson:"driver_id"`
	OfferID  string        `json:"offer_id" bson:"offer_id"`
	From     LatLngLiteral `json:"from" bson:"from"`
	To       LatLngLiteral `json:"to" bson:"to"`
	Price    Money         `json:"price" bson:"price"`
	Status   string        `json:"status" bson:"status"`
}

type LatLngLiteral struct {
	Lat float32 `json:"lat" bson:"lat"`
	Lng float32 `json:"lng" bson:"lng"`
}

type Money struct {
	Amount   float64 `json:"amount" bson:"amount"`
	Currency string  `json:"currency" bson:"currency"`
}

type Driver struct {
	ID   string  `json:"id"`
	Name string  `json:"name"`
	Lat  float32 `json:"lat"`
	Lng  float32 `json:"lng"`
}

type TripMessage struct {
	TripId string `json:"trip_id"`
}

type KafkaMessageToDriver struct {
	ID              string `json:"id"`
	Source          string `json:"source"`
	Type            string `json:"type"`
	DataContentType string `json:"datacontenttype"`
	Time            string `json:"time"`
	Data            Trip   `json:"data"`
}

type SendTripData struct {
	ID       string `json:"trip_id" bson:"_id,omitempty"`
	DriverID string `json:"driver_id" bson:"driver_id"`
}

type KafkaMessageToTrip struct {
	ID              string       `json:"id"`
	Source          string       `json:"source"`
	Type            string       `json:"type"`
	DataContentType string       `json:"datacontenttype"`
	Time            string       `json:"time"`
	Data            SendTripData `json:"data"`
}
