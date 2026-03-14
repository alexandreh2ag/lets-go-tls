package http

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	appCtx "github.com/alexandreh2ag/lets-go-tls/apps/server/context"
	mockTypes "github.com/alexandreh2ag/lets-go-tls/mocks/types"
	"github.com/alexandreh2ag/lets-go-tls/types/acme"
	"github.com/labstack/echo/v4"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestNewChallenge_Success(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	cache := mockTypes.NewMockCache[string](ctrl)
	want := &ChallengeHTTP{logger: ctx.Logger, fs: ctx.GetFS(), cache: cache, httpChallengeConfig: ctx.Config.Acme.HttpChallengeConfig}
	got := NewChallenge(ctx.Logger, ctx.GetFS(), cache, ctx.Config.Acme.HttpChallengeConfig)
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

	baseDocRoot := "/var/www"
	ctx.Config.Acme.HttpChallengeConfig.EnableDocumentRoot = true
	ctx.Config.Acme.HttpChallengeConfig.DocumentRoot = baseDocRoot
	tokenFile := "xxx-file"
	_ = ctx.Fs.Mkdir(ctx.Config.Acme.HttpChallengeConfig.DocumentRoot, 0775)
	_ = afero.WriteFile(ctx.GetFS(), filepath.Join(baseDocRoot, tokenFile), []byte("bar"), 0644)

	tests := []struct {
		name       string
		token      string
		mockFn     func(*mockTypes.MockCache[string])
		wantStatus int
	}{
		{
			name:  "SuccessTokenInCache",
			token: "xxx",
			mockFn: func(cache *mockTypes.MockCache[string]) {
				cache.EXPECT().Get(gomock.Any(), gomock.Any()).Times(1).Return("test", nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "SuccessTokenInFS",
			token: tokenFile,
			mockFn: func(cache *mockTypes.MockCache[string]) {
				cache.EXPECT().Get(gomock.Any(), gomock.Any()).Times(1).Return("", nil)
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "FailWithEmptyToken",
			token:      "",
			mockFn:     func(cache *mockTypes.MockCache[string]) {},
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
			challenge := &ChallengeHTTP{
				logger:              ctx.Logger,
				cache:               cache,
				fs:                  ctx.GetFS(),
				httpChallengeConfig: ctx.Config.Acme.HttpChallengeConfig,
			}
			err := challenge.Handler(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantStatus, rec.Code)
		})
	}
}

func TestChallengeHTTP_getFileChallengeKeyAuth_FileFound(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	basePath := "/var/www"
	token := "foo"
	keyAuth := "bar"
	path := filepath.Join(basePath, token)
	ctx.Config.Acme.HttpChallengeConfig.DocumentRoot = basePath
	_ = ctx.Fs.Mkdir(basePath, 0775)
	_ = afero.WriteFile(ctx.GetFS(), path, []byte(keyAuth), 0644)
	challenge := &ChallengeHTTP{fs: ctx.GetFS()}
	got, err := challenge.getFileChallengeKeyAuth(path)
	assert.NoError(t, err)
	assert.Equal(t, got, keyAuth)
}

func TestChallengeHTTP_getFileChallengeKeyAuth_FileNotFound(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	basePath := "/var/www"
	token := "foo"
	path := filepath.Join(basePath, token)
	ctx.Config.Acme.HttpChallengeConfig.DocumentRoot = basePath

	challenge := &ChallengeHTTP{fs: ctx.GetFS()}
	got, err := challenge.getFileChallengeKeyAuth(path)
	assert.Error(t, err)
	assert.Equal(t, got, "")
}
