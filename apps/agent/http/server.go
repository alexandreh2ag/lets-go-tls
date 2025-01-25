package http

import (
	"fmt"
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/context"
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/http/controller"
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/http/middleware"
	"github.com/alexandreh2ag/lets-go-tls/http"
	"github.com/labstack/echo-contrib/echoprometheus"
	"github.com/labstack/echo/v4"
	buildinHttp "net/http"
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
		if err != nil && err != buildinHttp.ErrServerClosed {
			panic(fmt.Errorf("fail to start http server with %v", err))
		}

	}()
	return e, nil
}
