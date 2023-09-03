package controller

import (
	"encoding/json"
	"errors"
	appCtx "github.com/alexandreh2ag/lets-go-tls/apps/agent/context"
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/http/middleware"
	mockTypes "github.com/alexandreh2ag/lets-go-tls/mocks/types"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetRequests_Success(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	requester := mockTypes.NewMockRequester(ctrl)
	want1 := []*types.DomainRequest{
		{Requester: requester, Domains: types.Domains{types.Domain("foo.com")}},
		{Requester: requester, Domains: types.Domains{types.Domain("bar.com")}},
	}
	want2 := []*types.DomainRequest{
		{Requester: requester, Domains: types.Domains{types.Domain("bar.foo.com"), types.Domain("foo.bar.com")}},
	}
	want := append(want1, want2...)
	wantJson, _ := json.Marshal(&want)
	ctx.Requesters = types.Requesters{"foo": requester, "bar": requester}
	gomock.InOrder(
		requester.EXPECT().Fetch().Times(1).Return(want1, nil),
		requester.EXPECT().Fetch().Times(1).Return(want2, nil),
		//requester.EXPECT().ID().Times(1).Return("bar"),
	)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(middleware.ContextKey, ctx)
	err := GetRequests(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, string(wantJson)+"\n", rec.Body.String())
}

func TestGetRequests_FailMultiError(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	requester := mockTypes.NewMockRequester(ctrl)
	want := []*types.DomainRequest{
		{Requester: requester, Domains: types.Domains{types.Domain("foo.com")}},
		{Requester: requester, Domains: types.Domains{types.Domain("bar.com")}},
	}
	wantJson, _ := json.Marshal(&want)
	ctx.Requesters = types.Requesters{"foo": requester, "bar": requester}
	gomock.InOrder(
		requester.EXPECT().Fetch().Times(1).Return(want, nil),
		requester.EXPECT().Fetch().Times(1).Return([]*types.DomainRequest{}, errors.New("fail")),
		requester.EXPECT().ID().Times(1).Return("bar"),
	)

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(middleware.ContextKey, ctx)
	err := GetRequests(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, string(wantJson)+"\n", rec.Body.String())
}
