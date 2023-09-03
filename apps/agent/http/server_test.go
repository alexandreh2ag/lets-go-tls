package http

import (
	appCtx "github.com/alexandreh2ag/lets-go-tls/apps/agent/context"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCreateServerHTTP(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctx.Config.HTTP.Listen = "127.0.0.1:0"
	ctx.Config.HTTP.MetricsEnable = true

	got, err := CreateServerHTTP(ctx)
	time.Sleep(200 * time.Millisecond)
	assert.NoError(t, err)
	assert.NotNil(t, got)
}
