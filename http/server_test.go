package http

import (
	"crypto/tls"
	"github.com/alexandreh2ag/lets-go-tls/config"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestCreateEcho(t *testing.T) {
	got := CreateEcho()

	assert.NotNil(t, got)
}

func TestGetApiPrefix(t *testing.T) {
	res := GetApiPrefix("certificates")
	assert.Equal(t, "/api/certificates", res)
}

func TestStartServerHTTP_SuccessListenHTTP(t *testing.T) {
	e := echo.New()

	go assert.NotPanics(t, func() {
		StartServerHTTP(e, "127.0.0.1:0", nil)
	})
	time.Sleep(200 * time.Millisecond)
	_ = e.Close()
}

func TestStartServerHTTP_SuccessListenHTTPS(t *testing.T) {
	e := echo.New()

	tlsConfig := &tls.Config{
		GetCertificate: func(info *tls.ClientHelloInfo) (*tls.Certificate, error) {
			cert, errLoadCert := tls.LoadX509KeyPair("./fixtures/cert.crt", "./fixtures/priv.key")
			if errLoadCert != nil {
				return nil, errLoadCert
			}
			return &cert, nil
		},
	}
	go assert.NotPanics(t, func() {
		StartServerHTTP(e, "127.0.0.1:0", tlsConfig)
	})
	time.Sleep(200 * time.Millisecond)
	_ = e.Close()
}

func TestStartServerHTTP_FailListen(t *testing.T) {
	e := echo.New()

	tlsConfig := &tls.Config{}
	assert.Panics(t, func() {
		StartServerHTTP(e, "127.0.0.1:0", tlsConfig)
	})
}

func TestCreateTLSConfig_Success(t *testing.T) {
	cfg := config.TLSConfig{
		CertPath: "./fixtures/cert.crt",
		KeyPath:  "./fixtures/priv.key",
	}
	tlsConfig := CreateTLSConfig(cfg)
	assert.NotNil(t, tlsConfig)
	cert, err := tlsConfig.GetCertificate(&tls.ClientHelloInfo{})
	assert.NoError(t, err)
	assert.NotNil(t, cert)
}

func TestCreateTLSConfig_Fail(t *testing.T) {
	cfg := config.TLSConfig{}
	tlsConfig := CreateTLSConfig(cfg)
	assert.NotNil(t, tlsConfig)
	cert, err := tlsConfig.GetCertificate(&tls.ClientHelloInfo{})
	assert.Error(t, err)
	assert.Nil(t, cert)
}
