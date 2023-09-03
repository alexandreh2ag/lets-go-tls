package controller

import (
	"bytes"
	"encoding/json"
	"errors"
	appCtx "github.com/alexandreh2ag/lets-go-tls/apps/server/context"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/http/middleware"
	appHttp "github.com/alexandreh2ag/lets-go-tls/http"
	mockTypesStorageState "github.com/alexandreh2ag/lets-go-tls/mocks/types/storage/state"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGetCertificatesFromRequests_Success(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	response := appHttp.ResponseCertificatesFromRequests{}
	cert1 := &types.Certificate{Identifier: "foo", Main: "foo.com", Domains: types.Domains{types.Domain("foo.com")}, ExpirationDate: time.Now(), Certificate: []byte("cert"), Key: []byte("key")}
	cert2 := &types.Certificate{Identifier: "bar", Main: "bar.com", Domains: types.Domains{types.Domain("bar.com")}, ExpirationDate: time.Now(), Certificate: []byte("cert"), Key: []byte("key")}
	state := &types.State{Certificates: types.Certificates{
		cert1,
		cert2,
	}}

	request1 := types.DomainRequest{Domains: types.Domains{"foo.com"}}
	request2 := types.DomainRequest{Domains: types.Domains{"bar.foo.com"}}

	response.Certificates = types.Certificates{cert1}
	response.Requests.Found = []*types.DomainRequest{&request1}
	response.Requests.NotFound = []*types.DomainRequest{&request2}
	wantJson, _ := json.Marshal(response)
	stateStorage := mockTypesStorageState.NewMockStorage(ctrl)
	stateStorage.EXPECT().Load().Times(1).Return(state, nil)
	ctx.StateStorage = stateStorage
	e := echo.New()
	jsonBody, _ := json.Marshal([]types.DomainRequest{request1, request2})
	body := bytes.NewReader(jsonBody)
	req := httptest.NewRequest(http.MethodGet, "/", body)
	req.Header.Add("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(middleware.ContextKey, ctx)

	err := GetCertificatesFromRequests(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, string(wantJson)+"\n", rec.Body.String())
}

func TestGetCertificatesFromRequests_SuccessWithInvalidCertificate(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	response := appHttp.ResponseCertificatesFromRequests{}
	cert1 := &types.Certificate{Identifier: "foo", Main: "foo.com", Domains: types.Domains{types.Domain("foo.com")}, ExpirationDate: time.Now()}
	state := &types.State{Certificates: types.Certificates{
		cert1,
	}}

	request1 := types.DomainRequest{Domains: types.Domains{"foo.com"}}

	response.Certificates = types.Certificates{}
	response.Requests.Found = []*types.DomainRequest{}
	response.Requests.NotFound = []*types.DomainRequest{&request1}
	wantJson, _ := json.Marshal(response)
	stateStorage := mockTypesStorageState.NewMockStorage(ctrl)
	stateStorage.EXPECT().Load().Times(1).Return(state, nil)
	ctx.StateStorage = stateStorage
	e := echo.New()
	jsonBody, _ := json.Marshal([]types.DomainRequest{request1})
	body := bytes.NewReader(jsonBody)
	req := httptest.NewRequest(http.MethodGet, "/", body)
	req.Header.Add("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(middleware.ContextKey, ctx)

	err := GetCertificatesFromRequests(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, string(wantJson)+"\n", rec.Body.String())
}

func TestGetCertificatesFromRequests_Fail_ParseBody(t *testing.T) {
	ctx := appCtx.TestContext(nil)

	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(middleware.ContextKey, ctx)

	err := GetCertificatesFromRequests(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

func TestGetCertificatesFromRequests_Fail_LoadState(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	stateStorage := mockTypesStorageState.NewMockStorage(ctrl)
	stateStorage.EXPECT().Load().Times(1).Return(nil, errors.New("fail"))
	ctx.StateStorage = stateStorage
	e := echo.New()
	jsonBody, _ := json.Marshal([]types.DomainRequest{{Domains: types.Domains{"foo.com"}}})
	body := bytes.NewReader(jsonBody)
	req := httptest.NewRequest(http.MethodGet, "/", body)
	req.Header.Add("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set(middleware.ContextKey, ctx)

	err := GetCertificatesFromRequests(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
