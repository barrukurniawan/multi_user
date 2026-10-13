package helpers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/barrukurniawan/multi_user/core/dto"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func WriteError(c *gin.Context, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		c.JSON(appErr.StatusCode, dto.Error(appErr.Code, appErr.Message))
		return
	}

	slog.Error("internal error", "err", err, "request_id", c.GetString("request_id"))
	c.JSON(http.StatusInternalServerError, dto.Error("INTERNAL_ERROR", "terjadi kesalahan pada server"))
}

func WriteBindError(c *gin.Context, err error) {
	c.JSON(http.StatusBadRequest, dto.Error("INVALID_INPUT", "request tidak valid: "+err.Error()))
}

func TaskIDParam(c *gin.Context) (string, bool) {
	id := c.Param("id")
	if _, err := uuid.Parse(id); err != nil {
		c.JSON(http.StatusBadRequest, dto.Error("INVALID_TASK_ID", "id task harus berformat UUID"))
		return "", false
	}
	return id, true
}

func CurrentUserID(c *gin.Context) string {
	return c.GetString("user_id")
}
