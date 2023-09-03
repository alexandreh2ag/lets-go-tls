package http

import (
	"fmt"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
)

const (
	ServerApiGetCertificates = "certificates"

	AgentApiRequests = "requests"
)

func GetApiPrefix(path string) string {
	return fmt.Sprintf("/api/%s", path)
}

func CreateEcho() *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(echoMiddleware.GzipWithConfig(echoMiddleware.GzipConfig{
		Level: 5,
	}))

	return e
}
