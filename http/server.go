package http

import (
	"crypto/tls"
	"fmt"
	"github.com/alexandreh2ag/lets-go-tls/config"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	buildinHttp "net/http"
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

func CreateTLSConfig(tlsConfig config.TLSConfig) *tls.Config {
	return &tls.Config{
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			cert, errLoadCert := tls.LoadX509KeyPair(tlsConfig.CertPath, tlsConfig.KeyPath)
			if errLoadCert != nil {
				return nil, errLoadCert
			}
			return &cert, nil
		},
	}
}

func StartServerHTTP(e *echo.Echo, listen string, tlsConfig *tls.Config) {
	server := buildinHttp.Server{
		Addr:    listen,
		Handler: e,
	}

	var errStart error
	if tlsConfig != nil {
		server.TLSConfig = tlsConfig
		errStart = server.ListenAndServeTLS("", "")
	} else {
		errStart = server.ListenAndServe()
	}
	if errStart != nil && errStart != buildinHttp.ErrServerClosed {
		panic(fmt.Errorf("fail to start http server with %v", errStart))
	}
}
