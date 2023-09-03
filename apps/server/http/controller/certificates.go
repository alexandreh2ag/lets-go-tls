package controller

import (
	"fmt"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/context"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/http/middleware"
	appHttp "github.com/alexandreh2ag/lets-go-tls/http"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/labstack/echo/v4"
	"net/http"
)

func GetCertificatesFromRequests(c echo.Context) error {
	ctx := c.Get(middleware.ContextKey).(*context.ServerContext)
	response := appHttp.ResponseCertificatesFromRequests{
		Certificates: types.Certificates{},
		Requests: appHttp.ResponseRequests{
			Found:    []*types.DomainRequest{},
			NotFound: []*types.DomainRequest{},
		},
	}
	domainsRequests := []*types.DomainRequest{}
	if err := c.Bind(&domainsRequests); err != nil || domainsRequests == nil || len(domainsRequests) == 0 {
		ctx.Logger.Error(fmt.Sprintf(
			"http request (%s): failed to parse body: %v",
			appHttp.GetApiPrefix(appHttp.ServerApiGetCertificates),
			err,
		))
		return c.JSON(http.StatusInternalServerError, response)
	}

	state, errLoad := ctx.StateStorage.Load()
	if errLoad != nil {
		ctx.Logger.Error(
			fmt.Sprintf(
				"http request (%s): failed to load state: %v",
				appHttp.GetApiPrefix(appHttp.ServerApiGetCertificates),
				errLoad,
			),
		)
		return c.JSON(http.StatusInternalServerError, response)
	}

	for _, request := range domainsRequests {
		ctx.Logger.Debug(fmt.Sprintf(
			"http request (%s): search matching certificate for %v",
			appHttp.GetApiPrefix(appHttp.ServerApiGetCertificates),
			request,
		))

		cert := state.Certificates.Match(request, true)
		if cert != nil {
			response.Certificates = append(response.Certificates, cert)
			response.Requests.Found = append(response.Requests.Found, request)
		} else {
			response.Requests.NotFound = append(response.Requests.NotFound, request)
			ctx.Logger.Warn(
				fmt.Sprintf(
					"http request (%s): does not found certificate for %v",
					appHttp.GetApiPrefix(appHttp.ServerApiGetCertificates),
					request,
				),
			)
		}
	}

	return c.JSON(http.StatusOK, response)
}
