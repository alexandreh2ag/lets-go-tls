package http

import (
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/context"
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/http/controller"
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/http/middleware"
	"github.com/alexandreh2ag/lets-go-tls/http"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
)

func CreateServerHTTP(ctx *context.AgentContext) (*echo.Echo, error) {
	e := http.CreateEcho()

	if ctx.Config.HTTP.MetricsEnable {
		e.Use(echoprometheus.NewMiddlewareWithConfig(echoprometheus.MiddlewareConfig{
			Registerer: ctx.MetricsRegister,
		}))
		e.GET("/metrics", echoprometheus.NewHandlerWithConfig(echoprometheus.HandlerConfig{
			Gatherer: ctx.MetricsRegister,
		}))
	}
	e.Use(middleware.HandlerContext(ctx))
	e.GET(http.GetApiPrefix(http.AgentApiRequests), controller.GetRequests)

	go func() {
		err := e.Start(ctx.Config.HTTP.Listen)
		ctx.Logger.Error(err.Error())

	}()
	return e, nil
}
