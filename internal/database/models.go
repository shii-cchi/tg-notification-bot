// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.26.0

package database

import (
	"time"
)

type Task struct {
	ID        int64
	Task      string
	TaskTime  string
	ChatID    int64
	Status    string
	CreatedAt time.Time
}
