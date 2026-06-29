package handler

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"

	authuc "eventlineup/internal/usecase/auth"
)

const oauthStateCookie = "oauth_state"

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
	// 5 min, path=/, HttpOnly, SameSite=Lax (survives the Google redirect back).
	// Secure is config-driven: true in production, off for local http dev.
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(oauthStateCookie, state, 300, "/", "", h.secureCookies, true)
	c.Redirect(http.StatusFound, h.uc.AuthCodeURL(state))
}

// callback validates CSRF state, exchanges the code for an organizer JWT, and
// redirects to the frontend with the token. The code and token are never logged.
func (h *AuthHandler) callback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	cookieState, _ := c.Cookie(oauthStateCookie)

	if code == "" || state == "" || cookieState == "" || state != cookieState {
		c.JSON(http.StatusBadRequest, gin.H{"error": "oauth exchange failed"})
		return
	}
	// One-time use: clear the state cookie (same attributes as when it was set).
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(oauthStateCookie, "", -1, "/", "", h.secureCookies, true)

	jwt, err := h.uc.HandleCallback(c.Request.Context(), code)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "oauth exchange failed"})
		return
	}
	c.Redirect(http.StatusFound, h.frontendURL+"/auth/callback?token="+url.QueryEscape(jwt))
}

func randomState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
