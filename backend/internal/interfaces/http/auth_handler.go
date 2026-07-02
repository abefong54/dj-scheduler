package handler

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"

	"eventlineup/internal/interfaces/http/middleware"
	authuc "eventlineup/internal/usecase/auth"
)

const (
	// oauthStateCookie is the name of the short-lived CSRF cookie that pins the
	// OAuth state value across the Google consent round-trip.
	oauthStateCookie = "oauth_state"

	// oauthStateTTL is how long that cookie stays valid — long enough to finish
	// the Google consent flow, short enough to limit replay of a leaked state.
	oauthStateTTL = 5 * time.Minute

	// deleteCookieMaxAge is the Max-Age that tells the browser to drop a cookie
	// immediately (Gin/`net/http` convention: any negative value expires it).
	deleteCookieMaxAge = -1

	// stateEntropyBytes is the amount of cryptographic randomness behind each
	// CSRF state token (256 bits, hex-encoded to 64 chars).
	stateEntropyBytes = 32
)

type AuthHandler struct {
	uc            *authuc.UseCase
	frontendURL   string
	secureCookies bool
}

func NewAuthHandler(uc *authuc.UseCase, frontendURL string, secureCookies bool) *AuthHandler {
	return &AuthHandler{uc: uc, frontendURL: frontendURL, secureCookies: secureCookies}
}

// Register mounts the auth routes. These are intentionally NOT behind the JWT
// middleware — they are how a user obtains a JWT in the first place.
func (h *AuthHandler) Register(r gin.IRoutes) {
	r.GET("/auth/google", h.redirect)
	r.GET("/auth/google/callback", h.callback)
}

// redirect sends the user to Google's consent screen, seeding a CSRF state
// value in both the redirect URL and a short-lived cookie.
func (h *AuthHandler) redirect(c *gin.Context) {
	state, err := randomState()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	// path=/, HttpOnly, SameSite=Lax (survives the Google redirect back).
	// Secure is config-driven: true in production, off for local http dev.
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(oauthStateCookie, state, int(oauthStateTTL.Seconds()), "/", "", h.secureCookies, true)
	c.Redirect(http.StatusFound, h.uc.AuthCodeURL(state))
}

// callback validates CSRF state, exchanges the code for an organizer JWT, and
// redirects to the frontend with the token. The code and token are never logged.
func (h *AuthHandler) callback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	cookieState, _ := c.Cookie(oauthStateCookie)

	if code == "" || state == "" || cookieState == "" || state != cookieState {
		// Client gets a generic message; the server log records WHICH precondition
		// failed (booleans only — no code, state value, or token is ever logged) so
		// a state-cookie problem can be told apart from a token-exchange one.
		slog.Warn("oauth callback rejected: state precheck failed",
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
	c.SetCookie(oauthStateCookie, "", deleteCookieMaxAge, "/", "", h.secureCookies, true)

	jwt, err := h.uc.HandleCallback(c.Request.Context(), code)
	if err != nil {
		// The wrapped error carries Google's real reason (e.g. invalid_client,
		// invalid_grant). It contains no secret — the client secret and the auth
		// code are not part of the returned error — so it is safe to log.
		slog.Error("oauth callback failed: code exchange / userinfo",
			"request_id", middleware.GetRequestID(c),
			"err", err,
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "oauth exchange failed"})
		return
	}
	c.Redirect(http.StatusFound, h.frontendURL+"/auth/callback?token="+url.QueryEscape(jwt))
}

func randomState() (string, error) {
	b := make([]byte, stateEntropyBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
