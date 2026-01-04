package models

import "time"

type (
	Heartbeat struct {
		ID          uint `gorm:"primaryKey"`
		Application string
		Timestamp   int64

		CreatedAt time.Time
		UpdatedAt time.Time
	}
)
