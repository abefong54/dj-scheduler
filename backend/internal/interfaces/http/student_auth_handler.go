package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	studentauthuc "eventlineup/internal/usecase/studentauth"
)

// StudentAuthHandler serves student Google sign-in (2026-07-05-student-google-
// login-design). It parallels AuthHandler (organizer) but on a distinct route so
// the Google client can register a second redirect URI, and it issues a
// role=student JWT. SCAFFOLD: routes are wired; methods return 501 until Track A
// (login backend) fills them.
type StudentAuthHandler struct {
	uc            *studentauthuc.UseCase
	frontendURL   string
	secureCookies bool
}

func NewStudentAuthHandler(uc *studentauthuc.UseCase, frontendURL string, secureCookies bool) *StudentAuthHandler {
	return &StudentAuthHandler{uc: uc, frontendURL: frontendURL, secureCookies: secureCookies}
}

// RegisterPublic mounts the student auth routes. Like the organizer auth routes
// they are NOT behind the JWT middleware — they are how a student obtains a JWT.
func (h *StudentAuthHandler) RegisterPublic(rg *gin.RouterGroup) {
	rg.GET("/student/auth/google", h.redirect)
	rg.GET("/student/auth/google/callback", h.callback)
}

// redirect sends the student to Google's consent screen. SCAFFOLD: not implemented.
func (h *StudentAuthHandler) redirect(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}

// callback exchanges the OAuth code for a student JWT, or 403 if the Google user
// is on no roster. SCAFFOLD: not implemented.
func (h *StudentAuthHandler) callback(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{"error": "not implemented"})
}
