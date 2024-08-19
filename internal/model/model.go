package model

import (
	"time"
)

type Task struct {
	ID         int64
	Task       string
	TaskTime   string
	ChatID     int64
	TaskTimeMs int
	CreatedAt  time.Time
}

type TaskInfo struct {
	TaskID       int64
	TaskWithTime string
}
