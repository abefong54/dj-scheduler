package middleware

import (
	"context"

	"github.com/gin-gonic/gin"

	"eventlineup/internal/domain/model"
)

// SchoolIDKey / RoleKey are the gin.Context keys for the caller's active school
// and their role in it (EL-104). Scoped handlers read SchoolIDKey and pass it to
// the repositories, which filter by school_id. Empty when the user hasn't joined
// a school yet — scoped queries then return nothing (fail-closed), and the
// onboarding flow (EL-106) routes them to register one.
const (
	SchoolIDKey = "school_id"
	RoleKey     = "role"
)

// noSchoolSentinel is set as the school_id when the caller has no active
// membership. It's a syntactically valid UUID that matches no real school, so
// scoped queries return empty (fail-closed) instead of erroring on an empty
// string cast to uuid. Handlers that must distinguish "no school" (e.g. /api/me)
// resolve the membership themselves rather than trusting this value.
const noSchoolSentinel = "00000000-0000-0000-0000-000000000000"

// SchoolResolver resolves the caller's active membership. Implemented by the
// membership repository; kept as a local interface so middleware doesn't depend
// on the infrastructure layer.
type SchoolResolver interface {
	ActiveForOrganizer(ctx context.Context, organizerID string) (model.Membership, error)
}

// School runs after Auth. It looks up the authenticated organizer's active
// membership and stashes their school_id + role in the context. It never aborts:
// a user with no membership simply has no school set (so /api/me and onboarding
// still work, and scoped reads return empty).
func School(resolver SchoolResolver) gin.HandlerFunc {
	return func(c *gin.Context) {
		schoolID := noSchoolSentinel
		organizerID, _ := c.Get(OrganizerIDKey)
		if oid, ok := organizerID.(string); ok && oid != "" {
			if m, err := resolver.ActiveForOrganizer(c.Request.Context(), oid); err == nil {
				schoolID = m.SchoolID
				c.Set(RoleKey, m.Role)
			}
		}
		c.Set(SchoolIDKey, schoolID)
		c.Next()
	}
}
