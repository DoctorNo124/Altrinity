package controllers

import (
	"altrinity/api/middleware"
	"altrinity/api/repositories"
	"altrinity/api/services"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

// VolunteerController handles volunteer map updates and admin streams.
type VolunteerController struct {
	Service *services.VolunteerService
}

type UserIdWrapper struct {
	UserId string `json:"userId"`
}

// Volunteer sends location updates periodically (mobile side).
func (vc *VolunteerController) UpdatePosition(c *gin.Context) {
	var pos repositories.Position
	if err := c.ShouldBindJSON(&pos); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid position data"})
		return
	}

	token := c.GetHeader("Authorization")
	tokenStr := strings.TrimPrefix(token, "Bearer ")

	ok, user, err := middleware.VerifyJWT(tokenStr, "volunteer")
	if !ok || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or unauthorized token"})
		return
	}

	// Always trust identity from token
	pos.ID = user.ID
	pos.FullName = user.FullName

	// Persist position to PostGIS (optional)
	if err := vc.Service.UpdatePosition(context.Background(), pos); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := vc.Service.AppendPositionToRoute(context.Background(), pos.ID, pos.Lat, pos.Lng); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Publish update to Redis so all admins see it
	data, _ := json.Marshal(pos)
	if err := vc.Service.Repo.Redis.Publish(context.Background(), "positions", string(data)).Err(); err != nil {
		fmt.Println("redis publish error:", err)
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// Admin subscribes to Redis "positions" channel via WebSocket.
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func (vc *VolunteerController) StreamPositions(c *gin.Context) {
	tokenStr := c.Query("token")
	if tokenStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	ok, _, err := middleware.VerifyJWT(tokenStr, "admin")
	if !ok || err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or unauthorized token"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("websocket upgrade failed:", err)
		return
	}
	defer conn.Close()

	sub := vc.Service.Repo.Redis.Subscribe(context.Background(), "positions")
	defer sub.Close()

	for msg := range sub.Channel() {
		// Each message payload already includes volunteer ID + full name
		conn.WriteMessage(websocket.TextMessage, []byte(msg.Payload))
	}
}

// REST endpoint for debugging / fallback (optional).
func (vc *VolunteerController) GetPositions(c *gin.Context) {
	positions, err := vc.Service.GetAllPositions(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch positions"})
		return
	}
	if positions == nil {
		positions = []repositories.Position{}
	}
	c.JSON(http.StatusOK, positions)
}

func (vc *VolunteerController) GetRoutePositions(c *gin.Context) {
	userID := c.Query("user_id") // optional filter for a specific volunteer
	points, err := vc.Service.RouteRepo.GetAllPositions(context.Background(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch positions"})
		return
	}
	c.JSON(http.StatusOK, points)
}

func (vc *VolunteerController) CompleteRoute(c *gin.Context) {
	var userIdWrapper UserIdWrapper
	if err := c.ShouldBindJSON(&userIdWrapper); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid position data"})
		return
	}
	err := vc.Service.CompleteActiveRoute(context.Background(), userIdWrapper.UserId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to complete route"})
		return
	}
	c.JSON(http.StatusOK, "nice")
}
