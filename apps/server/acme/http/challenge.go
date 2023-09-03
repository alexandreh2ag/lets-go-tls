package http

import (
	"context"
	"fmt"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/alexandreh2ag/lets-go-tls/types/acme"
	"github.com/labstack/echo/v4"
	"log/slog"
	"net/http"
)

var _ acme.Challenge = &ChallengeHTTP{}

type ChallengeHTTP struct {
	cache  types.Cache
	logger *slog.Logger
}

func NewChallenge(logger *slog.Logger, cache types.Cache) *ChallengeHTTP {
	return &ChallengeHTTP{logger: logger, cache: cache}
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
	if keyAuth != "" {
		return c.String(http.StatusOK, keyAuth)
	}

	return c.NoContent(http.StatusNotFound)
}
