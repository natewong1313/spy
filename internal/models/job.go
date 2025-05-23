package models

import (
	"time"

	"github.com/google/uuid"
)

type Job struct {
	ID        uuid.UUID
	Company   string
	Title     string
	Locations []string
	URL       string
	UpdatedAt time.Time
	CreatedAt time.Time
}
