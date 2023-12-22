package model

import(
	"github.com/google/uuid"
)

type Driver struct {
	ID uuid.UUID
	Location Location
}