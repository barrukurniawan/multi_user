package tests

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/barrukurniawan/multi_user/core/dto"
	"github.com/barrukurniawan/multi_user/core/helpers"
	"github.com/gin-gonic/gin"
)

func errorCode(err error) string {
	var appErr *helpers.AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return ""
}

func newTestServer() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(helpers.LoggingMiddleware(), helpers.RecoveryMiddleware())
	r.GET("/ok", func(c *gin.Context) { c.JSON(http.StatusOK, dto.Success("ok")) })
	r.GET("/panic", func(c *gin.Context) { panic("rahasia: password database bocor") })
	r.GET("/private", helpers.AuthMiddleware(helpers.NewJWTManager("test-secret", 24)), func(c *gin.Context) {})
	return r
}

func call(r *gin.Engine, path string) *httptest.ResponseRecorder {
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	return w
}

func TestError(t *testing.T) {
	r := newTestServer()

	for _, tc := range []struct {
		path   string
		status int
		code   string
	}{
		{"/private", 401, "UNAUTHORIZED"},
		{"/panic", 500, "INTERNAL_ERROR"},
	} {
		w := call(r, tc.path)

		var body map[string]any
		json.Unmarshal(w.Body.Bytes(), &body)
		if w.Code != tc.status || body["code"] != tc.code {
			t.Fatalf("%s: dapat %d %v, seharusnya %d %s", tc.path, w.Code, body["code"], tc.status, tc.code)
		}
		for _, field := range []string{"status", "code", "message", "timestamp"} {
			if body[field] == nil {
				t.Fatalf("%s: field %q tidak ada di response error: %s", tc.path, field, w.Body)
			}
		}
		if strings.Contains(w.Body.String(), "rahasia") {
			t.Fatalf("%s: detail internal bocor ke response: %s", tc.path, w.Body)
		}
	}
}
