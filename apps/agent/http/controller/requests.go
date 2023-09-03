package controller

import (
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/context"
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/http/middleware"
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/requester"
	"github.com/hashicorp/go-multierror"
	"github.com/labstack/echo/v4"
	"net/http"
)

func GetRequests(c echo.Context) error {
	ctx := c.Get(middleware.ContextKey).(*context.AgentContext)
	domainsRequests, err := requester.FetchRequests(ctx.Requesters)
	if merr, ok := err.(*multierror.Error); ok {
		for _, errRequester := range merr.Errors {
			ctx.Logger.Error(errRequester.Error())
		}
	}
	return c.JSON(http.StatusOK, domainsRequests)
}
