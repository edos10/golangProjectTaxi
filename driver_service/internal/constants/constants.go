package constants

import (
	"os"
	"strconv"
)

var DRIVER_SEARCH = "DRIVER_SEARCH"
var DRIVER_FOUND = "DRIVER_FOUND"
var ON_POSITION = "ON_POSITION"
var STARTED = "STARTED"
var ENDED = "ENDED"
var CANCELED = "CANCELED"

var PORT_TOPIC_DRIVER = "29092"

var RADIUS_INVALID, _ = strconv.ParseFloat(os.Getenv("RADIUS"), 32)
var RADIUS = float32(RADIUS_INVALID)
