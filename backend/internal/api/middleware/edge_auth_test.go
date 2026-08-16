package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ajaxe/email-ingestion/internal/api/middleware"
	"github.com/ajaxe/email-ingestion/internal/ingest"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestEdgeAuth(t *testing.T) {
	e := echo.New()
	secret := "my-secret-edge-token-12345"

	handler := middleware.EdgeAuth(secret)(func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	})

	t.Run("valid token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set(ingest.HeaderEdgeToken, secret)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("valid token with trailing whitespace or newline", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set(ingest.HeaderEdgeToken, secret+"  \n")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("empty token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler(c)
		assert.Error(t, err)
		he, ok := err.(*echo.HTTPError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusUnauthorized, he.Code)
	})

	t.Run("mismatched token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/", nil)
		req.Header.Set(ingest.HeaderEdgeToken, "wrong-token")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		err := handler(c)
		assert.Error(t, err)
		he, ok := err.(*echo.HTTPError)
		assert.True(t, ok)
		assert.Equal(t, http.StatusUnauthorized, he.Code)
	})
}
