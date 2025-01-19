package http

import (
	"github.com/alexandreh2ag/lets-go-tls/apps/server/acme/http"
	appCtx "github.com/alexandreh2ag/lets-go-tls/apps/server/context"
	appProm "github.com/alexandreh2ag/lets-go-tls/prometheus"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCreateServerHTTP(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctx.Config.HTTP.Listen = "127.0.0.1:0"
	ctx.Config.HTTP.MetricsEnable = true
	ctx.MetricsRegister = appProm.NewRegistry(types.NameServerMetrics, prometheus.NewRegistry())
	httpProvider := http.NewChallenge(ctx.Logger, ctx.Cache)
	got, err := CreateServerHTTP(ctx, httpProvider)
	time.Sleep(200 * time.Millisecond)
	assert.NoError(t, err)
	assert.NotNil(t, got)
}

func TestCreateServerHTTPWithTLS_Fail(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctx.Config.HTTP.Listen = "127.0.0.1:0"
	ctx.Config.HTTP.MetricsEnable = true
	ctx.MetricsRegister = appProm.NewRegistry(types.NameServerMetrics, prometheus.NewRegistry())
	ctx.Config.HTTP.TLS.Enable = true
	ctx.Config.HTTP.TLS.CertPath = "./fixtures/cert.crt"
	ctx.Config.HTTP.TLS.KeyPath = "./fixtures/priv.key"
	httpProvider := http.NewChallenge(ctx.Logger, ctx.Cache)

	got, err := CreateServerHTTP(ctx, httpProvider)
	time.Sleep(200 * time.Millisecond)
	assert.NoError(t, err)
	assert.NotNil(t, got)

}
