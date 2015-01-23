package main

import (
	"time"
)

type Event struct {
	UID         string    `json:"uid"`
	Summary     string    `json:"summary"`
	Description string    `json:"description"`
	Speaker     string    `json:"speaker"`
	Location    string    `json:"location"`
	Category    string    `json:"category"`
	URL         string    `json:"url"`
	Start       time.Time `json:"start"`
	Duration    float64   `json:"duration"`
}
