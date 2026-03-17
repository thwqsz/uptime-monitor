package models

import "time"

type Target struct {
	ID           int64
	UserID       int64
	URL          string
	Timeout      int
	IntervalTime int
	CreatedAt    time.Time
}
