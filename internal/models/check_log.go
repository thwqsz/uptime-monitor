package models

import "time"

type CheckLog struct {
	ID             int64
	TargetID       int64
	Status         string
	StatusCode     int
	ErrorMsg       *string
	ResponseTimeMs int
	CheckedAt      time.Time
}
