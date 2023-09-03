package http

import (
	"errors"
	"fmt"
	appCtx "github.com/alexandreh2ag/lets-go-tls/apps/server/context"
	mockTypes "github.com/alexandreh2ag/lets-go-tls/mocks/types"
	"github.com/alexandreh2ag/lets-go-tls/types/acme"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewChallenge_Success(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	cache := mockTypes.NewMockCache[string](ctrl)
	want := &ChallengeHTTP{logger: ctx.Logger, cache: cache}
	got := NewChallenge(ctx.Logger, cache)
	assert.Equal(t, want, got)
}

func TestChallenge_ID(t *testing.T) {
	challenge := &ChallengeHTTP{}
	assert.Equal(t, acme.TypeHTTP01, challenge.ID())
}

func TestChallenge_Type(t *testing.T) {
	challenge := &ChallengeHTTP{}
	assert.Equal(t, acme.TypeHTTP01, challenge.Type())
}

func TestChallenge_GetCacheKey(t *testing.T) {
	want := "acme_http_token_example.com"
	challenge := &ChallengeHTTP{}
	assert.Equal(t, want, challenge.GetCacheKey("token", "example.com"))
}

func TestChallenge_Present_Success(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	token := "xxx"
	domain := "example.com"
	keyAuth := "foo"
	cache := mockTypes.NewMockCache[string](ctrl)
	cache.EXPECT().Set(gomock.Any(), gomock.Eq(fmt.Sprintf("acme_http_%s_%s", token, domain)), keyAuth).Times(1).Return(nil)
	challenge := &ChallengeHTTP{logger: ctx.Logger, cache: cache}

	err := challenge.Present(domain, token, keyAuth)
	assert.NoError(t, err)
}

func TestChallenge_Present_FailWriteToCache(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	token := "xxx"
	domain := "example.com"
	keyAuth := "foo"
	cache := mockTypes.NewMockCache[string](ctrl)
	cache.EXPECT().Set(gomock.Any(), gomock.Eq(fmt.Sprintf("acme_http_%s_%s", token, domain)), keyAuth).Times(1).Return(errors.New("error"))
	challenge := &ChallengeHTTP{logger: ctx.Logger, cache: cache}

	err := challenge.Present(domain, token, keyAuth)
	assert.Error(t, err)
}

func TestChallenge_CleanUp_Success(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	token := "xxx"
	domain := "example.com"
	keyAuth := "foo"
	cache := mockTypes.NewMockCache[string](ctrl)
	cache.EXPECT().Delete(gomock.Any(), gomock.Eq(fmt.Sprintf("acme_http_%s_%s", token, domain))).Times(1).Return(nil)
	challenge := &ChallengeHTTP{logger: ctx.Logger, cache: cache}

	err := challenge.CleanUp(domain, token, keyAuth)
	assert.NoError(t, err)
}

func TestChallenge_CleanUp_FailWriteToCache(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	token := "xxx"
	domain := "example.com"
	keyAuth := "foo"
	cache := mockTypes.NewMockCache[string](ctrl)
	cache.EXPECT().Delete(gomock.Any(), gomock.Eq(fmt.Sprintf("acme_http_%s_%s", token, domain))).Times(1).Return(errors.New("error"))
	challenge := &ChallengeHTTP{logger: ctx.Logger, cache: cache}

	err := challenge.CleanUp(domain, token, keyAuth)
	assert.Error(t, err)
}

func TestChallenge_Handler(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name       string
		token      string
		mockFn     func(*mockTypes.MockCache[string])
		wantStatus int
	}{
		{
			name:  "Success",
			token: "xxx",
			mockFn: func(cache *mockTypes.MockCache[string]) {
				cache.EXPECT().Get(gomock.Any(), gomock.Any()).Times(1).Return("test", nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "FailWithEmptyToken",
			token: "",
			mockFn: func(cache *mockTypes.MockCache[string]) {
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:  "FailWithGetCacheError",
			token: "xxx",
			mockFn: func(cache *mockTypes.MockCache[string]) {
				cache.EXPECT().Get(gomock.Any(), gomock.Any()).Times(1).Return("", errors.New("error"))
			},
			wantStatus: http.StatusNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cache := mockTypes.NewMockCache[string](ctrl)
			tt.mockFn(cache)
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/:token")
			c.SetParamNames("token")
			c.SetParamValues(tt.token)
			challenge := &ChallengeHTTP{logger: ctx.Logger, cache: cache}
			err := challenge.Handler(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}
