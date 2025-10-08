package services

import (
	"altrinity/api/models"
	"altrinity/api/repositories"
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type RouteService interface {
	SaveRoute(ctx context.Context, userID string, points []models.RoutePoint) error
	GetUserRoutes(ctx context.Context, userID string) ([]models.Route, error)
}

type routeService struct {
	repo repositories.RouteRepo
}

func NewRouteService(repo repositories.RouteRepo) RouteService {
	return &routeService{repo: repo}
}

func (s *routeService) SaveRoute(ctx context.Context, userID string, points []models.RoutePoint) error {
	if len(points) < 2 {
		return errors.New("not enough points")
	}

	processed := []map[string]interface{}{}
	for i := 0; i < len(points)-1; i++ {
		dur := points[i+1].Timestamp - points[i].Timestamp
		processed = append(processed, map[string]interface{}{
			"lat":      points[i].Lat,
			"lng":      points[i].Lng,
			"duration": dur,
		})
	}

	data, err := json.Marshal(processed)
	if err != nil {
		return err
	}

	uuidParsed, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	route := &models.Route{
		ID:        uuid.New(),
		UserID:    uuidParsed,
		Points:    string(data),
		CreatedAt: time.Now(),
	}

	return s.repo.Create(ctx, route)
}

func (s *routeService) GetUserRoutes(ctx context.Context, userID string) ([]models.Route, error) {
	uuidParsed, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	return s.repo.GetByUser(ctx, uuidParsed)
}
