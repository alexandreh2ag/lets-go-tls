package middleware

import (
	appCtx "github.com/alexandreh2ag/lets-go-tls/apps/agent/context"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlerContext(t *testing.T) {
	ctx := appCtx.TestContext(nil)

	e := echo.New()
	e.Use(HandlerContext(ctx))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	handler := HandlerContext(ctx)(func(c echo.Context) error {
		return nil
	})
	err := handler(c)
	got := c.Get(ContextKey)
	assert.NoError(t, err)
	assert.Equal(t, ctx, got)
}
