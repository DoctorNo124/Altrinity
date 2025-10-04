package controllers

import (
	"net/http"

	"altrinity/api/services"

	"github.com/gin-gonic/gin"
)

type AdminController struct {
	Service *services.AdminService
}

func (a *AdminController) ListUsers(c *gin.Context) {
	users, err := a.Service.ListUsers()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return
	}
	c.JSON(http.StatusOK, users)
}

func (a *AdminController) UpdateUserRole(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Role string `json:"role"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if err := a.Service.UpdateUserRole(id, req.Role); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update role"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "role updated"})
}
