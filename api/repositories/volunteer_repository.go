package repositories

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/jmoiron/sqlx"
)

type VolunteerRepository struct {
	DB    *sqlx.DB
	Redis *redis.Client
}

type Position struct {
	ID        string    `db:"volunteer_id" json:"id"`
	FullName  string    `db:"full_name" json:"fullName"`
	Lat       float64   `db:"lat" json:"lat"`
	Lng       float64   `db:"lng" json:"lng"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
}

// Upsert latest position into PostGIS
func (r *VolunteerRepository) UpsertPosition(ctx context.Context, pos Position) error {
	query := `
	INSERT INTO volunteer_positions (volunteer_id, full_name, position, updated_at)
	VALUES ($1, $2, ST_SetSRID(ST_MakePoint($3, $4), 4326)::geography, NOW())
	ON CONFLICT (volunteer_id) DO UPDATE
	SET full_name = EXCLUDED.full_name,
	    position = EXCLUDED.position,
	    updated_at = NOW();`
	_, err := r.DB.ExecContext(ctx, query, pos.ID, pos.FullName, pos.Lat, pos.Lng)
	return err
}

// Get last persisted position for comparison
func (r *VolunteerRepository) GetLastPosition(ctx context.Context, userID string) (Position, error) {
	var p Position
	query := `SELECT volunteer_id, full_name, ST_Y(position::geometry) AS lat,
		       ST_X(position::geometry) AS lng, updated_at FROM volunteer_positions WHERE volunteer_id = $1`
	err := r.DB.GetContext(ctx, &p, query, userID)
	return p, err
}

func (r *VolunteerRepository) GetAllPositions(ctx context.Context) ([]Position, error) {
	var positions []Position
	err := r.DB.SelectContext(ctx, &positions, `
		SELECT volunteer_id,
		       ST_Y(position::geometry) AS lat,
		       ST_X(position::geometry) AS lng
		FROM volunteer_positions`)
	return positions, err
}
