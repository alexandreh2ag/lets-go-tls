package http

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/alexandreh2ag/lets-go-tls/apps/server/config"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/alexandreh2ag/lets-go-tls/types/acme"
	"github.com/labstack/echo/v4"
	"github.com/spf13/afero"
)

var _ acme.Challenge = &ChallengeHTTP{}

type ChallengeHTTP struct {
	cache               types.Cache
	logger              *slog.Logger
	fs                  afero.Fs
	httpChallengeConfig config.HttpChallengeConfig
}

func NewChallenge(logger *slog.Logger, fs afero.Fs, cache types.Cache, httpChallengeConfig config.HttpChallengeConfig) *ChallengeHTTP {
	return &ChallengeHTTP{logger: logger, fs: fs, cache: cache, httpChallengeConfig: httpChallengeConfig}
}

func (ch *ChallengeHTTP) ID() string {
	return acme.TypeHTTP01
}

func (ch *ChallengeHTTP) Type() string {
	return acme.TypeHTTP01
}

func (ch *ChallengeHTTP) GetCacheKey(token, domain string) string {
	return fmt.Sprintf("acme_http_%s_%s", token, domain)
}

func (ch *ChallengeHTTP) Present(domain, token, keyAuth string) error {
	ch.logger.Debug(fmt.Sprintf("present domain=%s,token=%s,keyauth=%s", domain, token, keyAuth), "provider", ch.Type())
	cacheKey := ch.GetCacheKey(token, domain)
	ctx := context.Background()
	err := ch.cache.Set(ctx, cacheKey, keyAuth)
	if err != nil {
		return fmt.Errorf("failed to store in cache keyAuth for token %s - domain %s: %v", token, domain, err)
	}
	return nil
}

func (ch *ChallengeHTTP) CleanUp(domain, token, _ string) error {
	ch.logger.Debug(fmt.Sprintf("clean up domain=%s,token=%s", domain, token), "provider", "http-01")
	cacheKey := ch.GetCacheKey(token, domain)
	ctx := context.Background()
	err := ch.cache.Delete(ctx, cacheKey)
	if err != nil {
		return fmt.Errorf("failed to delete in cache keyAuth for token %s - domain %s: %v", token, domain, err)
	}

	return nil
}

func (ch *ChallengeHTTP) Handler(c echo.Context) error {
	ch.logger.Debug(fmt.Sprintf("challenge start"), "provider", "http-01")
	reg := c.Request()
	host := reg.Host
	token := c.Param("token")

	if token == "" {
		return c.NoContent(http.StatusNotFound)
	}

	ch.logger.Debug(fmt.Sprintf("challenge host=%s,token=%s", host, token), "provider", "http-01")
	cacheKey := ch.GetCacheKey(token, host)
	ctx := context.Background()
	keyAuth, err := ch.cache.Get(ctx, cacheKey)
	if err != nil {
		ch.logger.Error(fmt.Sprintf("failed to get in cache keyAuth for token %s - domain %s: %v", token, host, err))
	}

	if keyAuth == "" && ch.httpChallengeConfig.EnableDocumentRoot {
		tokenPath := filepath.Join(ch.httpChallengeConfig.DocumentRoot, token)
		keyAuth, err = ch.getFileChallengeKeyAuth(tokenPath)
		if err != nil {
			keyAuth = ""
			ch.logger.Warn(
				fmt.Sprintf(
					"failed to get keyAuth (%s - domain %s) for domain from file (%s): %v",
					token,
					host,
					tokenPath,
					err,
				),
			)
		}
	}

	if keyAuth != "" {
		return c.String(http.StatusOK, keyAuth)
	}

	return c.NoContent(http.StatusNotFound)
}

func (ch *ChallengeHTTP) getFileChallengeKeyAuth(path string) (string, error) {
	keyAuth, err := afero.ReadFile(ch.fs, path)

	return string(keyAuth), err
}
