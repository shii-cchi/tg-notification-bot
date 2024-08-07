package model

import "time"

type Task struct {
	Task       string
	ChatID     int64
	TaskTimeMs int
	CreatedAt  time.Time
}
