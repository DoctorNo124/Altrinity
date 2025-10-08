package controllers

import (
	"altrinity/api/models"
	"altrinity/api/services"
	"net/http"

	"github.com/gin-gonic/gin"
)

type RouteController struct {
	service services.RouteService
}

func NewRouteController(service services.RouteService) *RouteController {
	return &RouteController{service: service}
}

func (ctrl *RouteController) RegisterRoutes(router *gin.RouterGroup) {
	routes := router.Group("/routes")

	// Apply your AuthMiddleware when you register routes in main.go
	{
		routes.POST("", ctrl.SaveRoute)            // authenticated user saves own route
		routes.GET("/:userID", ctrl.GetUserRoutes) // admin or supervisor fetches a user's route
	}
}

// The POST body now only needs the array of route points
type SaveRouteRequest struct {
	Route []models.RoutePoint `json:"route"`
}

func (ctrl *RouteController) SaveRoute(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user context"})
		return
	}

	var body SaveRouteRequest
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON"})
		return
	}

	if err := ctrl.service.SaveRoute(c, userID.(string), body.Route); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "route saved"})
}

func (ctrl *RouteController) GetUserRoutes(c *gin.Context) {
	userID := c.Param("userID")

	routes, err := ctrl.service.GetUserRoutes(c, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, routes)
}
