package models

import "time"

type (
	Heartbeat struct {
		ID uint `gorm:"primaryKey"`

		Application string
		Timestamp   int64
		Metadata    string

		CreatedAt time.Time
		UpdatedAt time.Time
	}
)
