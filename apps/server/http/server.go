package http

import (
	"fmt"
	acmeHttp "github.com/alexandreh2ag/lets-go-tls/apps/server/acme/http"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/context"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/http/controller"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/http/middleware"
	"github.com/alexandreh2ag/lets-go-tls/http"
	"github.com/labstack/echo-contrib/echoprometheus"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	buildinHttp "net/http"
)

const (
	AcmeEndpoint = "/.well-known/acme-challenge"
)

func CreateServerHTTP(ctx *context.ServerContext, httpProvider *acmeHttp.ChallengeHTTP) (*echo.Echo, error) {
	e := http.CreateEcho()

	if ctx.Config.HTTP.MetricsEnable {
		e.Use(echoprometheus.NewMiddlewareWithConfig(echoprometheus.MiddlewareConfig{
			Registerer: ctx.MetricsRegister,
		}))
		e.GET("/metrics", echoprometheus.NewHandlerWithConfig(echoprometheus.HandlerConfig{
			Gatherer: ctx.MetricsRegister,
		}))
	}
	e.Use(
		middleware.HandlerContext(ctx),
	)
	e.Any(fmt.Sprintf("%s/:token", AcmeEndpoint), httpProvider.Handler)

	authorizedGroup := e.Group("", echojwt.WithConfig(
		echojwt.Config{
			SigningMethod: ctx.Config.JWT.Method,
			SigningKey:    []byte(ctx.Config.JWT.Key),
		},
	))
	authorizedGroup.POST(http.GetApiPrefix(http.ServerApiGetCertificates), controller.GetCertificatesFromRequests)

	go func() {
		var err error
		if ctx.Config.HTTP.TLS.Enable {
			err = e.StartTLS(ctx.Config.HTTP.Listen, ctx.Config.HTTP.TLS.CertPath, ctx.Config.HTTP.TLS.KeyPath)
		} else {
			err = e.Start(ctx.Config.HTTP.Listen)
		}
		if err != nil && err != buildinHttp.ErrServerClosed {
			panic(fmt.Errorf("fail to start http server with %v", err))
		}
	}()

	return e, nil
}
