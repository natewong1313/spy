package models

import "time"

type Company struct {
	Name           string
	PlatformType   string // "greenhouse" | "ashby"
	PlatformURL    string
	CreatedAt      time.Time
	GreenhouseName string
	AshbyName      string
}
