package models

import "time"

type (
	Session struct {
		ID             uint `gorm:"primaryKey"`
		Application    string
		StartTimestamp int64
		EndTimestamp   int64

		CreatedAt time.Time
		UpdatedAt time.Time
	}
)
