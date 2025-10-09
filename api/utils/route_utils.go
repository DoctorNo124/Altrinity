package utils

import (
	"encoding/json"
	"log"

	"altrinity/api/models"
)

// ParseRoutePoints safely parses a route's JSON string into []RoutePoint.
// Returns nil if parsing fails or string is empty.
func ParseRoutePoints(pointsJSON string) []models.RoutePoint {
	if pointsJSON == "" {
		return nil
	}

	var points []models.RoutePoint
	if err := json.Unmarshal([]byte(pointsJSON), &points); err != nil {
		log.Printf("⚠️ Failed to parse route points: %v", err)
		return nil
	}

	return points
}
