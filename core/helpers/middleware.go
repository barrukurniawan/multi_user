package helpers

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
	"time"

	"github.com/barrukurniawan/multi_user/core/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func AuthMiddleware(jwtManager *JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, found := strings.CutPrefix(c.GetHeader("Authorization"), "Bearer ")
		if !found || token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Error("UNAUTHORIZED", "header Authorization: Bearer <token> wajib diisi"))
			return
		}

		claims, err := jwtManager.Validate(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, dto.Error("INVALID_TOKEN", "token tidak valid atau sudah expired"))
			return
		}

		c.Set("user_id", claims.UserID)
		c.Next()
	}
}

func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := uuid.NewString()
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)

		start := time.Now()
		c.Next()
		status := c.Writer.Status()

		level := slog.LevelInfo
		if status >= 500 {
			level = slog.LevelError
		} else if status >= 400 {
			level = slog.LevelWarn
		}

		slog.Log(c.Request.Context(), level, "request completed",
			"request_id", requestID,
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status_code", status,
			"latency_ms", time.Since(start).Milliseconds(),
		)
	}
}

func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				slog.Error("panic recovered",
					"request_id", c.GetString("request_id"),
					"panic", r,
					"stack", string(debug.Stack()),
				)

				c.AbortWithStatusJSON(http.StatusInternalServerError, dto.Error("INTERNAL_ERROR", "terjadi kesalahan pada server"))
			}
		}()
		c.Next()
	}
}
