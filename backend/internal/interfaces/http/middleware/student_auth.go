package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// StudentIDKey is the gin.Context key under which the authenticated student's DJ
// id is stored. Student handlers read it via c.GetString(middleware.StudentIDKey).
// The student's school lands in SchoolIDKey (read from the token, not a
// membership lookup — a student = one DJ = one school).
const StudentIDKey = "student_id"

// studentClaims is the subset of the student JWT the middleware validates. It
// mirrors token.Claims over the wire (role, school_id, subject=dj_id). Kept local
// so the middleware package doesn't depend on the token package.
type studentClaims struct {
	Role     string `json:"role"`
	SchoolID string `json:"school_id"`
	jwt.RegisteredClaims
}

// StudentAuth returns middleware that validates a Bearer student JWT signed with
// secret, requires role=student, and stashes the student's DJ id (SubjectKey) and
// school_id in the context. On any failure it aborts with 401 and never reaches
// the handler.
//
// SCAFFOLD: the JWT is parsed and the role/scope are enforced, so context keys
// are set correctly for downstream handlers. Track A owns any refinement (e.g.
// confirming the DJ still exists / is still a student).
func StudentAuth(secret string) gin.HandlerFunc {
	key := []byte(secret)
	return func(c *gin.Context) {
		const prefix = "Bearer "
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, prefix) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		tokenStr := strings.TrimSpace(authHeader[len(prefix):])

		var claims studentClaims
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

		// Require a student session with a resolved school — fail closed otherwise.
		if claims.Role != "student" || claims.Subject == "" || claims.SchoolID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}

		c.Set(StudentIDKey, claims.Subject)
		c.Set(SchoolIDKey, claims.SchoolID)
		c.Next()
	}
}
