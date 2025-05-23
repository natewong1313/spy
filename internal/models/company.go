package models

import "time"

type Company struct {
	Name           string
	PlatformType   string // "greenhouse"
	PlatformURL    string
	CreatedAt      time.Time
	GreenhouseName string
}
