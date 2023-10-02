package models

import (
	"github.com/google/uuid"
)

type Job struct {
	UUID   uuid.UUID `json:"uuid"`
	Status string    `json:"status"`
}

type Jobs []struct {
	Job Job
}
