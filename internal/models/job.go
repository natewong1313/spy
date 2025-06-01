package models

import (
	"time"
)

type Job struct {
	Company   string
	Title     string
	Locations []string
	URL       string
	UpdatedAt time.Time
	CreatedAt time.Time
}
