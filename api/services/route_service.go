package services

import (
	"altrinity/api/models"
	"altrinity/api/repositories"
	"altrinity/api/utils"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type RouteService interface {
	SaveRoute(ctx context.Context, userID string, points []models.RoutePoint) error
	GetUserRoutes(ctx context.Context, userID string) ([]models.Route, error)
	GetLatestUserRoute(ctx context.Context, userID string) (*models.Route, error)
	GetById(ctx context.Context, id string) (*models.Route, error)
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

	data, err := json.Marshal(points)
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

// ✅ Returns all routes for a user, with parsed Points.
func (s *routeService) GetUserRoutes(ctx context.Context, userID string) ([]models.Route, error) {
	uuidParsed, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	routes, err := s.repo.GetByUser(ctx, uuidParsed)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch routes: %w", err)
	}

	for i := range routes {
		routes[i].PointsParsed = utils.ParseRoutePoints(routes[i].Points)
	}

	return routes, nil
}

// ✅ Returns the most recent route for a user, or nil if none.
func (s *routeService) GetLatestUserRoute(ctx context.Context, userID string) (*models.Route, error) {
	uuidParsed, err := uuid.Parse(userID)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	route, err := s.repo.GetLatestByUser(ctx, uuidParsed)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch latest route: %w", err)
	}
	if route == nil {
		return nil, nil
	}

	route.PointsParsed = utils.ParseRoutePoints(route.Points)
	return route, nil
}

// ✅ Returns the most recent route for a user, or nil if none.
func (s *routeService) GetById(ctx context.Context, id string) (*models.Route, error) {
	uuidParsed, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid route ID: %w", err)
	}

	route, err := s.repo.GetById(ctx, uuidParsed)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch route by Id: %w", err)
	}
	if route == nil {
		return nil, nil
	}

	route.PointsParsed = utils.ParseRoutePoints(route.Points)
	return route, nil
}
