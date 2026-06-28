// Package token signs and describes the organizer JWTs issued after sign-in.
// The Claims shape here mirrors the one the auth middleware validates; both
// interoperate over the JSON wire format (organizer_id, email, standard claims).
package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims is the JWT payload issued to authenticated organizers.
type Claims struct {
	OrganizerID string `json:"organizer_id"`
	Email       string `json:"email"`
	jwt.RegisteredClaims
}

// Sign issues an HS256 JWT for the given organizer, valid for ttl.
func Sign(secret, organizerID, email string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		OrganizerID: organizerID,
		Email:       email,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}
