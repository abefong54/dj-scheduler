package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"eventlineup/internal/domain/apperrors"
	"eventlineup/internal/interfaces/http/middleware"
)

// respondError maps a usecase/domain error to a safe HTTP response (EL-039).
//
// Known application errors carry client-safe messages authored in code:
//   - ValidationError → 400 with its message
//   - ConflictError   → 409 with structured conflict details
//   - ErrNotFound     → 404 "not found"
//   - ErrForbidden    → 403 "forbidden"
//
// Anything else is treated as an internal fault: the real error is logged
// server-side (with the request id) and the client gets a fixed
// 500 {"error":"internal error"} — raw pgx/driver text is never echoed back,
// which would otherwise help an attacker fingerprint the schema.
func respondError(c *gin.Context, err error) {
	var validation *apperrors.ValidationError
	var conflict *apperrors.ConflictError
	switch {
	case errors.As(err, &validation):
		c.JSON(http.StatusBadRequest, gin.H{"error": validation.Message})
	case errors.As(err, &conflict):
		c.JSON(http.StatusConflict, gin.H{
			"error":               "conflict",
			"type":                conflict.Type,
			"conflicting_slot_id": conflict.ConflictingSlotID,
		})
	case errors.Is(err, apperrors.ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
	case errors.Is(err, apperrors.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
	default:
		slog.Error("request failed",
			"request_id", middleware.GetRequestID(c),
			"method", c.Request.Method,
			"route", c.FullPath(),
			"err", err,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
	}
}
