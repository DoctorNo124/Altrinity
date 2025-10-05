package middleware

import (
	"errors"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

// VerifiedUser represents decoded JWT user info
type VerifiedUser struct {
	ID       string
	Username string
	Email    string
	FullName string
	Roles    []string
}

// VerifyJWT validates a JWT token using the cached JWKS.
// It returns (isValid, *VerifiedUser, error)
func VerifyJWT(tokenStr string, requiredRole string) (bool, *VerifiedUser, error) {
	if jwks == nil {
		return false, nil, errors.New("JWKS not initialized")
	}

	token, err := jwt.Parse(tokenStr, jwks.Keyfunc)
	if err != nil {
		return false, nil, fmt.Errorf("failed to parse token: %w", err)
	}
	if !token.Valid {
		return false, nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return false, nil, errors.New("invalid claims structure")
	}

	// Basic claims
	user := &VerifiedUser{
		ID:       stringOrEmpty(claims["sub"]),
		Username: stringOrEmpty(claims["preferred_username"]),
		Email:    stringOrEmpty(claims["email"]),
	}

	if full, ok := claims["name"].(string); ok {
		user.FullName = full
	} else {
		first := stringOrEmpty(claims["given_name"])
		last := stringOrEmpty(claims["family_name"])
		user.FullName = strings.TrimSpace(first + " " + last)
	}
	// Extract roles from Keycloak's realm_access
	if ra, ok := claims["realm_access"].(map[string]interface{}); ok {
		if roles, ok := ra["roles"].([]interface{}); ok {
			for _, r := range roles {
				user.Roles = append(user.Roles, r.(string))
			}
		}
	}

	// Optional role enforcement
	if requiredRole != "" && !user.HasRole(requiredRole) {
		return false, user, fmt.Errorf("user lacks required role: %s", requiredRole)
	}

	return true, user, nil
}

// Helper to check if user has role
func (u *VerifiedUser) HasRole(role string) bool {
	for _, r := range u.Roles {
		if r == role {
			return true
		}
	}
	return false
}

func stringOrEmpty(v interface{}) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
