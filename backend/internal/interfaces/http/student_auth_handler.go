package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/interfaces/http/middleware"
	studentauthuc "eventlineup/internal/usecase/studentauth"
)

// studentOAuthStateCookie is the student flow's CSRF state cookie. It is distinct
// from the organizer cookie name so the two OAuth round-trips can be in flight in
// the same browser without clobbering each other.
const studentOAuthStateCookie = "student_oauth_state"

// StudentAuthHandler serves student Google sign-in (2026-07-05-student-google-
// login-design). It parallels AuthHandler (organizer) but on a distinct route so
// the Google client can register a second redirect URI, and it issues a
// role=student JWT.
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

// redirect sends the student to Google's consent screen, seeding a CSRF state
// value in both the redirect URL and a short-lived cookie (mirrors AuthHandler).
func (h *StudentAuthHandler) redirect(c *gin.Context) {
	state, err := randomState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(studentOAuthStateCookie, state, int(oauthStateTTL.Seconds()), "/", "", h.secureCookies, true)
	c.Redirect(http.StatusFound, h.uc.AuthCodeURL(state))
}

// callback validates CSRF state, exchanges the code for a student JWT, and
// redirects to the frontend student login route with the token — matching the
// organizer callback's redirect-with-token delivery. A Google user on no roster
// gets 403 ("ask your teacher"); enrollment is teacher-controlled so we don't
// leak whether the account exists. The code and token are never logged.
func (h *StudentAuthHandler) callback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	cookieState, _ := c.Cookie(studentOAuthStateCookie)

	if code == "" || state == "" || cookieState == "" || state != cookieState {
		// Booleans only — no code, state value, or token is ever logged.
		slog.Warn("student oauth callback rejected: state precheck failed",
			"request_id", middleware.GetRequestID(c),
			"has_code", code != "",
			"has_state", state != "",
			"has_state_cookie", cookieState != "",
			"state_matches_cookie", state == cookieState,
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "oauth exchange failed"})
		return
	}
	// One-time use: clear the state cookie (same attributes as when it was set).
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(studentOAuthStateCookie, "", deleteCookieMaxAge, "/", "", h.secureCookies, true)

	jwt, err := h.uc.HandleCallback(c.Request.Context(), code)
	if err != nil {
		// No roster match → 403 with a clear, non-leaking message. Enrollment is
		// teacher-controlled, so a stranger's Google account is turned away here.
		if errors.Is(err, apperrors.ErrNotFound) {
			slog.Warn("student oauth callback: no roster match",
				"request_id", middleware.GetRequestID(c),
			)
			c.JSON(http.StatusForbidden, gin.H{"error": "you're not on a roster yet — ask your teacher to add your email"})
			return
		}
		// The wrapped error carries Google's real reason (no secret, no auth code).
		slog.Error("student oauth callback failed: code exchange / userinfo",
			"request_id", middleware.GetRequestID(c),
			"err", err,
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "oauth exchange failed"})
		return
	}
	c.Redirect(http.StatusFound, h.frontendURL+"/student/login?token="+url.QueryEscape(jwt))
}
