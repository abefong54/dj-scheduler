// Package middleware holds Gin middleware for the HTTP interface layer.
package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// OrganizerIDKey is the gin.Context key under which the authenticated
// organizer's ID is stored. Handlers read it via c.Get(middleware.OrganizerIDKey).
const OrganizerIDKey = "organizer_id"

// Claims is the JWT payload issued to authenticated organizers.
type Claims struct {
	OrganizerID string `json:"organizer_id"`
	Email       string `json:"email"`
	jwt.RegisteredClaims
}

// Auth returns middleware that validates a Bearer JWT signed with secret.
// On success it stores the organizer ID in the context and calls the next
// handler; otherwise it aborts with 401 and never reaches the handler.
func Auth(secret string) gin.HandlerFunc {
	key := []byte(secret)
	return func(c *gin.Context) {
		const prefix = "Bearer "
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, prefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		tokenStr := strings.TrimSpace(authHeader[len(prefix):])

		var claims Claims
		_, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return key, nil
		})
		if err != nil {
			if errors.Is(err, jwt.ErrTokenExpired) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token expired"})
				return
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		c.Set(OrganizerIDKey, claims.OrganizerID)
		c.Next()
	}
}
