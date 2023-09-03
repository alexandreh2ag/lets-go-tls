package http

import (
	"github.com/alexandreh2ag/lets-go-tls/apps/server/acme/http"
	appCtx "github.com/alexandreh2ag/lets-go-tls/apps/server/context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCreateServerHTTP(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctx.Config.HTTP.Listen = "127.0.0.1:0"
	ctx.Config.HTTP.MetricsEnable = true
	httpProvider := http.NewChallenge(ctx.Logger, ctx.Cache)
	got, err := CreateServerHTTP(ctx, httpProvider)
	time.Sleep(200 * time.Millisecond)
	assert.NoError(t, err)
	assert.NotNil(t, got)
}
