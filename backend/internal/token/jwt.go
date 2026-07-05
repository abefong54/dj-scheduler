// Package token signs and describes the JWTs issued after sign-in. The Claims
// shape here mirrors the one the auth middleware validates; both interoperate
// over the JSON wire format (organizer_id, email, role, school_id, standard
// claims). Role distinguishes an organizer session from a student session.
package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Role values carried in a JWT. Every token issued going forward sets a role;
// there is no back-compat path (no real users to preserve).
const (
	RoleOrganizer = "organizer"
	RoleStudent   = "student"
)

// Claims is the JWT payload. OrganizerID is set for organizer tokens; for a
// student token it is empty and the student's DJ id travels in the subject
// (RegisteredClaims.Subject). Role is always set; SchoolID is set for student
// tokens (a student = one DJ = one school), letting student middleware scope by
// the token instead of a membership lookup.
type Claims struct {
	OrganizerID string `json:"organizer_id"`
	Email       string `json:"email"`
	Role        string `json:"role"`
	SchoolID    string `json:"school_id"`
	jwt.RegisteredClaims
}

// Sign issues an HS256 JWT for the given organizer, valid for ttl. The role is
// stamped as organizer.
func Sign(secret, organizerID, email string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		OrganizerID: organizerID,
		Email:       email,
		Role:        RoleOrganizer,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}

// SignStudent issues an HS256 JWT for a student, valid for ttl. The student's DJ
// id is the subject; role is student and school_id scopes the session (a student
// = one DJ = one school), so student middleware reads the school from the token.
func SignStudent(secret, djID, schoolID string, ttl time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		Role:     RoleStudent,
		SchoolID: schoolID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   djID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
}
