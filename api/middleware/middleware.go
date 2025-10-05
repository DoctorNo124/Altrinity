package middleware

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
)

var jwks *keyfunc.JWKS

// Initialize JWKS cache from Keycloak
func InitJWKS() {
	jwksURL := os.Getenv("KEYCLOAK_URL") + "/realms/" + os.Getenv("KEYCLOAK_REALM") + "/protocol/openid-connect/certs"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	jwksOpts := keyfunc.Options{
		Ctx: ctx,
	}
	var err error
	jwks, err = keyfunc.Get(jwksURL, jwksOpts)
	if err != nil {
		log.Fatalf("Failed to get JWKS from Keycloak: %v", err)
	}
	log.Println("✅ JWKS initialized from Keycloak")
}

// AuthMiddleware verifies JWTs and (optionally) enforces a role
func AuthMiddleware(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing Authorization header"})
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, jwks.Keyfunc)
		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid claims"})
			return
		}

		sub, _ := claims["sub"].(string)
		realmAccess, _ := claims["realm_access"].(map[string]interface{})
		rolesRaw, _ := realmAccess["roles"].([]interface{})
		var roles []string
		for _, r := range rolesRaw {
			roles = append(roles, r.(string))
		}

		// Role check
		if requiredRole != "" {
			has := false
			for _, r := range roles {
				if r == requiredRole {
					has = true
					break
				}
			}
			if !has {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
				return
			}
		}

		c.Set("user_id", sub)
		c.Set("roles", roles)
		preferredUsername, _ := claims["preferred_username"].(string)
		email, _ := claims["email"].(string)
		c.Set("username", preferredUsername)
		c.Set("email", email)
		c.Next()
	}
}
