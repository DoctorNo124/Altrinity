package services

import (
	"altrinity/api/repositories"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"github.com/redis/go-redis/v9"
)

type VolunteerService struct {
	Repo      *repositories.VolunteerRepository // Wraps Postgres and Redis connections
	RouteRepo *repositories.RouteRepo
}

// Position update threshold logic
const (
	MinDistanceMeters = 50.0            // Only persist if volunteer moved >50m
	MinUpdateInterval = 5 * time.Minute // Or if last update >5 minutes ago
)

func (s *VolunteerService) UpdatePosition(ctx context.Context, pos repositories.Position) error {
	// --- Store in Redis for live map ---
	key := fmt.Sprintf("position:%s", pos.ID)

	payload, _ := json.Marshal(pos)
	s.Repo.Redis.Set(ctx, key, string(payload), 10*time.Minute) // expire old ones
	s.Repo.Redis.Publish(ctx, "positions", string(payload))

	// --- Check last persisted position ---
	last, err := s.Repo.GetLastPosition(ctx, pos.ID)
	if err != nil && err != redis.Nil {
		return err
	}

	if shouldPersist(pos, last) {
		return s.Repo.UpsertPosition(ctx, pos)
	}
	return nil
}

// shouldPersist returns true if user moved significantly or time expired
func shouldPersist(curr, last repositories.Position) bool {
	if last.ID == "" {
		return true // first time
	}

	distance := haversine(curr.Lat, curr.Lng, last.Lat, last.Lng)
	timeDiff := time.Since(last.UpdatedAt)

	return distance > MinDistanceMeters || timeDiff > MinUpdateInterval
}

// haversine formula to compute distance between two coords
func haversine(lat1, lon1, lat2, lon2 float64) float64 {
	const R = 6371000.0 // meters
	dLat := (lat2 - lat1) * math.Pi / 180
	dLon := (lon2 - lon1) * math.Pi / 180
	lat1r := lat1 * math.Pi / 180
	lat2r := lat2 * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1r)*math.Cos(lat2r)*math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return R * c
}

func (s *VolunteerService) GetAllPositions(ctx context.Context) ([]repositories.Position, error) {
	return s.Repo.GetAllPositions(ctx)
}

func (s *VolunteerService) AppendPositionToRoute(ctx context.Context, userID string, lat, lng float64) error {
	return s.RouteRepo.AppendPositionToRedis(ctx, userID, lat, lng)
}

func (s *VolunteerService) CompleteActiveRoute(ctx context.Context, userID string) error {
	return s.RouteRepo.FinalizeRoute(ctx, userID)
}
