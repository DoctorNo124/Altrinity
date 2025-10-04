package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	keycloakURL  string
	realm        string
	clientID     string
	clientSecret string
)

type introspectionResponse struct {
	Active   bool     `json:"active"`
	Username string   `json:"username,omitempty"`
	Scope    string   `json:"scope,omitempty"`
	Roles    []string `json:"roles,omitempty"` // you may need to customize claim extraction
}

// AuthMiddleware always calls introspection endpoint
func AuthMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing or invalid Authorization header"})
			c.Abort()
			return
		}

		rawToken := strings.TrimPrefix(authHeader, "Bearer ")

		if !introspectToken(rawToken, requiredRole) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or unauthorized token"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// Call Keycloak introspection endpoint
func introspectToken(token, requiredRole string) bool {
	introspectURL := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token/introspect", keycloakURL, realm)

	data := fmt.Sprintf("token=%s&client_id=%s&client_secret=%s", token, clientID, clientSecret)
	req, _ := http.NewRequest("POST", introspectURL, bytes.NewBufferString(data))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("introspection error:", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("introspection failed with status:", resp.Status)
		return false
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		fmt.Println("decode error:", err)
		return false
	}

	// Must be active
	if active, ok := body["active"].(bool); !ok || !active {
		return false
	}

	// Optional: enforce role
	if requiredRole != "" {
		if ra, ok := body["realm_access"].(map[string]interface{}); ok {
			if roles, ok := ra["roles"].([]interface{}); ok {
				for _, r := range roles {
					if r.(string) == requiredRole {
						return true
					}
				}
				return false
			}
		}
	}

	return true
}
