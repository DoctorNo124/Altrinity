package models

import (
	"time"

	"github.com/google/uuid"
)

type Route struct {
	ID        uuid.UUID `db:"id" json:"id"`
	UserID    uuid.UUID `db:"user_id" json:"user_id"`
	Points    string    `db:"points" json:"points"` // JSON string (jsonb in Postgres)
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type RoutePoint struct {
	Lat       float64 `json:"lat"`
	Lng       float64 `json:"lng"`
	Timestamp int64   `json:"timestamp"`
}
