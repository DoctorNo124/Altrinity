package repositories

import (
	"altrinity/api/models"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type RouteRepo struct {
	DB    *sqlx.DB
	Redis *redis.Client
}

func (r *RouteRepo) Create(ctx context.Context, route *models.Route) error {
	query := `
		INSERT INTO routes (id, user_id, points, created_at)
		VALUES ($1, $2, $3::jsonb, $4)
	`
	_, err := r.DB.ExecContext(ctx, query,
		route.ID,
		route.UserID,
		route.Points,
		time.Now(),
	)
	return err
}

func (r *RouteRepo) GetByUser(ctx context.Context, userID uuid.UUID) ([]models.Route, error) {
	query := `
		SELECT id, user_id, points, created_at
		FROM routes
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	var routes []models.Route
	err := r.DB.SelectContext(ctx, &routes, query, userID)
	return routes, err
}

func (r *RouteRepo) GetLatestByUser(ctx context.Context, userID uuid.UUID) (*models.Route, error) {
	query := `
		SELECT id, user_id, points, created_at
		FROM routes
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	var route models.Route
	err := r.DB.GetContext(ctx, &route, query, userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &route, nil
}

func (r *RouteRepo) GetById(ctx context.Context, id uuid.UUID) (*models.Route, error) {
	query := `
		SELECT id, user_id, points, created_at
		FROM routes
		WHERE id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	var route models.Route
	err := r.DB.GetContext(ctx, &route, query, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &route, nil
}

func (r *RouteRepo) AppendPositionToRedis(ctx context.Context, userID string, lat, lng float64) error {
	key := fmt.Sprintf("route:%s", userID)
	lastRaw, _ := r.Redis.LIndex(ctx, key, -1).Result()

	var last Position
	var weight float64
	now := time.Now().Unix()

	if lastRaw != "" {
		_ = json.Unmarshal([]byte(lastRaw), &last)
		delta := float64(now-last.Timestamp) / 60.0 // seconds → minutes
		weight = math.Min(math.Max(delta, 0.2), 20.0)
	} else {
		weight = 0.2
	}

	pos := Position{Lat: lat, Lng: lng, Timestamp: now, Weight: weight}
	data, _ := json.Marshal(pos)

	if err := r.Redis.RPush(ctx, key, data).Err(); err != nil {
		return err
	}
	return r.Redis.Publish(ctx, "positions", data).Err()
}

func (r *RouteRepo) FinalizeRoute(ctx context.Context, userID string) error {
	key := fmt.Sprintf("route:%s", userID)
	points, err := r.Redis.LRange(ctx, key, 0, -1).Result()
	if err != nil || len(points) == 0 {
		return err
	}

	var coords []string
	for _, raw := range points {
		var p Position
		_ = json.Unmarshal([]byte(raw), &p)
		coords = append(coords, fmt.Sprintf("%f %f", p.Lng, p.Lat))
	}
	linestring := fmt.Sprintf("LINESTRING(%s)", strings.Join(coords, ","))

	_, err = r.DB.ExecContext(ctx,
		`INSERT INTO routes (volunteer_id, geom, active, end_time)
		 VALUES ($1, ST_GeomFromText($2,4326), false, now())`,
		userID, linestring)
	if err == nil {
		r.Redis.Del(ctx, key)
	}
	return err
}

// Return all routes with dwell heat points in GeoJSON format
func (r *RouteRepo) GetAllRoutesGeoJSON(ctx context.Context) (map[string]interface{}, error) {
	query := `
	WITH routes_geo AS (
	  SELECT
	    r.id,
	    r.volunteer_id,
	    ST_AsGeoJSON(r.geom) AS geometry
	  FROM routes r
	),
	dwells AS (
	  SELECT
	    route_id,
	    ST_X(geom) AS lng,
	    ST_Y(geom) AS lat,
	    dwell_seconds
	  FROM dwell_points
	)
	SELECT jsonb_build_object(
	  'type', 'FeatureCollection',
	  'features', jsonb_agg(
	    jsonb_build_object(
	      'type', 'Feature',
	      'geometry', routes_geo.geometry::jsonb,
	      'properties', jsonb_build_object(
	        'volunteer_id', routes_geo.volunteer_id,
	        'route_id', routes_geo.id,
	        'dwells', (
	          SELECT jsonb_agg(jsonb_build_object(
	            'lat', lat,
	            'lng', lng,
	            'dwell_seconds', dwell_seconds
	          )) FROM dwells WHERE dwells.route_id = routes_geo.id
	        )
	      )
	    )
	  )
	)
	FROM routes_geo;
	`

	var resultJSON []byte
	err := r.DB.QueryRowContext(ctx, query).Scan(&resultJSON)
	if err != nil {
		return nil, err
	}

	var geo map[string]interface{}
	err = json.Unmarshal(resultJSON, &geo)
	return geo, err
}

func (r *RouteRepo) GetAllPositions(ctx context.Context, userID string) ([]Position, error) {
	key := fmt.Sprintf("route:%s", userID)
	values, err := r.Redis.LRange(ctx, key, 0, -1).Result()
	if err != nil {
		return nil, err
	}
	var points []Position
	for _, v := range values {
		var p Position
		if err := json.Unmarshal([]byte(v), &p); err == nil {
			points = append(points, p)
		}
	}
	return points, nil
}
