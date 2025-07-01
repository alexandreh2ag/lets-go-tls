package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/config"
	appCtx "github.com/alexandreh2ag/lets-go-tls/apps/agent/context"
	"github.com/alexandreh2ag/lets-go-tls/hook"
	appHttp "github.com/alexandreh2ag/lets-go-tls/http"
	mockHttp "github.com/alexandreh2ag/lets-go-tls/mocks/http"
	mockPrometheus "github.com/alexandreh2ag/lets-go-tls/mocks/prometheus"
	mockTypes "github.com/alexandreh2ag/lets-go-tls/mocks/types"
	mockTypesStorageCertificate "github.com/alexandreh2ag/lets-go-tls/mocks/types/storage/certificate"
	mockTypesStorageState "github.com/alexandreh2ag/lets-go-tls/mocks/types/storage/state"
	appProm "github.com/alexandreh2ag/lets-go-tls/prometheus"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/alexandreh2ag/lets-go-tls/types/storage/certificate"
	"github.com/jonboulle/clockwork"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"go.uber.org/mock/gomock"
	"log/slog"
	"net/http"
	"testing"
	"time"
)

var (
	domainRequestFoo     = &types.DomainRequest{Domains: types.Domains{types.Domain("foo.com")}}
	domainRequestBar     = &types.DomainRequest{Domains: types.Domains{types.Domain("bar.com")}}
	domainRequestNoFound = &types.DomainRequest{Domains: types.Domains{types.Domain("notfound.com")}}

	certificateFoo = &types.Certificate{Identifier: "foo.com-0", Domains: types.Domains{types.Domain("foo.com")}}
	certificateBar = &types.Certificate{Identifier: "bar.com-0", Domains: types.Domains{types.Domain("bar.com")}}
)

func TestNewService(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	stateStorage := mockTypesStorageState.NewMockStorage(ctrl)
	ctx.StateStorage = stateStorage
	hookManager := hook.NewManagerHook(ctx.Logger)
	want := &AgentService{
		managerConfig: ctx.Config.Manager,
		stateStorage:  stateStorage,
		storages:      ctx.Storages,
		httpClient:    ctx.GetHttpClient(),
		logger:        ctx.Logger,
		clock:         clockwork.NewRealClock(),
		hookManager:   hookManager,
	}
	got := NewService(ctx)
	got.hookManager = hookManager
	assert.Equal(t, want, got)
}

func TestAgentService_Run_Success(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx.MetricsRegister = appProm.NewRegistry(types.NameAgentMetrics, prometheus.NewRegistry())
	state := &types.State{Account: nil, Certificates: types.Certificates{certificateBar}}
	storageState := mockTypesStorageState.NewMockStorage(ctrl)
	storageState.EXPECT().Load().Times(1).Return(state, nil)
	storageState.EXPECT().Save(gomock.Any()).Times(1).Return(nil)

	requester := mockTypes.NewMockRequester(ctrl)
	requester.EXPECT().Fetch().Times(1).Return([]*types.DomainRequest{domainRequestFoo, domainRequestBar, domainRequestNoFound}, nil)
	ctx.Requesters = types.Requesters{"foo": requester}

	clientHttp := mockHttp.NewMockClient(ctrl)
	resp := fasthttp.Response{}
	resp.SetStatusCode(http.StatusOK)
	responseCertificate := appHttp.ResponseCertificatesFromRequests{
		Certificates: types.Certificates{certificateFoo, certificateBar},
		Requests: appHttp.ResponseRequests{
			Found:    []*types.DomainRequest{domainRequestFoo, domainRequestBar},
			NotFound: []*types.DomainRequest{domainRequestNoFound},
		},
	}
	body, _ := json.Marshal(responseCertificate)
	resp.SetBody(body)
	clientHttp.EXPECT().DoTimeout(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).SetArg(1, resp).Return(nil)

	storage := mockTypesStorageCertificate.NewMockStorage(ctrl)
	storage.EXPECT().Save(gomock.Any(), gomock.Any()).Times(1).Return(nil)
	storage.EXPECT().Delete(gomock.Any(), gomock.Any()).Times(1).Return(nil)

	as := &AgentService{
		logger:       ctx.Logger,
		stateStorage: storageState,
		httpClient:   clientHttp,
		storages:     certificate.Storages{"foo": storage},
		hookManager:  hook.NewManagerHook(ctx.Logger),
	}
	go as.hookManager.Start()

	err := as.Run(ctx)
	assert.NoError(t, err)
}

func TestAgentService_Run_FailLoadState(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx.MetricsRegister = appProm.NewRegistry(types.NameAgentMetrics, prometheus.NewRegistry())
	state := &types.State{Account: nil, Certificates: types.Certificates{certificateBar}}
	storageState := mockTypesStorageState.NewMockStorage(ctrl)
	storageState.EXPECT().Load().Times(1).Return(state, nil)
	storageState.EXPECT().Save(gomock.Any()).Times(1).Return(errors.New("error load"))

	as := &AgentService{
		logger:       ctx.Logger,
		stateStorage: storageState,
		hookManager:  hook.NewManagerHook(ctx.Logger),
	}
	go as.hookManager.Start()
	err := as.Run(ctx)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error load")
}

func TestAgentService_Run_SuccessWithErrorFetchRequest(t *testing.T) {
	b := bytes.NewBufferString("")
	ctx := appCtx.TestContext(b)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx.MetricsRegister = appProm.NewRegistry(types.NameAgentMetrics, prometheus.NewRegistry())
	state := &types.State{Account: nil, Certificates: types.Certificates{certificateBar}}
	storageState := mockTypesStorageState.NewMockStorage(ctrl)
	storageState.EXPECT().Load().Times(1).Return(state, nil)
	storageState.EXPECT().Save(gomock.Any()).Times(1).Return(nil)

	requester := mockTypes.NewMockRequester(ctrl)
	gomock.InOrder(
		requester.EXPECT().Fetch().Times(1).Return([]*types.DomainRequest{domainRequestFoo, domainRequestBar}, nil),
		requester.EXPECT().Fetch().Times(1).Return(nil, errors.New("error fetch")),
		requester.EXPECT().ID().Times(1).Return("bar"),
	)

	ctx.Requesters = types.Requesters{"foo": requester, "bar": requester}

	clientHttp := mockHttp.NewMockClient(ctrl)
	resp := fasthttp.Response{}
	resp.SetStatusCode(http.StatusOK)
	responseCertificate := appHttp.ResponseCertificatesFromRequests{
		Certificates: types.Certificates{certificateFoo, certificateBar},
		Requests: appHttp.ResponseRequests{
			Found:    []*types.DomainRequest{domainRequestFoo, domainRequestBar},
			NotFound: []*types.DomainRequest{},
		},
	}
	body, _ := json.Marshal(responseCertificate)
	resp.SetBody(body)
	clientHttp.EXPECT().DoTimeout(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).SetArg(1, resp).Return(nil)

	storage := mockTypesStorageCertificate.NewMockStorage(ctrl)
	storage.EXPECT().Save(gomock.Any(), gomock.Any()).Times(1).Return(nil)
	storage.EXPECT().Delete(gomock.Any(), gomock.Any()).Times(1).Return(nil)

	as := &AgentService{
		logger:       ctx.Logger,
		stateStorage: storageState,
		httpClient:   clientHttp,
		storages:     certificate.Storages{"foo": storage},
		hookManager:  hook.NewManagerHook(ctx.Logger),
	}
	go as.hookManager.Start()

	err := as.Run(ctx)
	assert.NoError(t, err)
	assert.Contains(t, b.String(), "requester (bar) fetching domains request failed: error fetch")
}

func TestAgentService_Run_SuccessWithErrorGetRequestManager(t *testing.T) {
	b := bytes.NewBufferString("")
	ctx := appCtx.TestContext(b)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx.MetricsRegister = appProm.NewRegistry(types.NameAgentMetrics, prometheus.NewRegistry())
	state := &types.State{Account: nil, Certificates: types.Certificates{certificateBar}}
	storageState := mockTypesStorageState.NewMockStorage(ctrl)
	storageState.EXPECT().Load().Times(1).Return(state, nil)

	requester := mockTypes.NewMockRequester(ctrl)
	requester.EXPECT().Fetch().Times(1).Return([]*types.DomainRequest{domainRequestFoo, domainRequestBar}, nil)

	ctx.Requesters = types.Requesters{"foo": requester}

	clientHttp := mockHttp.NewMockClient(ctrl)

	clientHttp.EXPECT().DoTimeout(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(errors.New("error get request manager"))

	as := &AgentService{
		logger:       ctx.Logger,
		stateStorage: storageState,
		httpClient:   clientHttp,
		hookManager:  hook.NewManagerHook(ctx.Logger),
	}
	go as.hookManager.Start()

	err := as.Run(ctx)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error get request manager")
}

func TestAgentService_Run_SuccessStorageSaveError(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx.MetricsRegister = appProm.NewRegistry(types.NameAgentMetrics, prometheus.NewRegistry())
	state := &types.State{Account: nil, Certificates: types.Certificates{certificateBar}}
	storageState := mockTypesStorageState.NewMockStorage(ctrl)
	storageState.EXPECT().Load().Times(1).Return(state, nil)
	storageState.EXPECT().Save(gomock.Any()).Times(1).Return(nil)

	requester := mockTypes.NewMockRequester(ctrl)
	requester.EXPECT().Fetch().Times(1).Return([]*types.DomainRequest{domainRequestFoo, domainRequestBar, domainRequestNoFound}, nil)
	ctx.Requesters = types.Requesters{"foo": requester}

	clientHttp := mockHttp.NewMockClient(ctrl)
	resp := fasthttp.Response{}
	resp.SetStatusCode(http.StatusOK)
	responseCertificate := appHttp.ResponseCertificatesFromRequests{
		Certificates: types.Certificates{certificateFoo, certificateBar},
		Requests: appHttp.ResponseRequests{
			Found:    []*types.DomainRequest{domainRequestFoo, domainRequestBar},
			NotFound: []*types.DomainRequest{domainRequestNoFound},
		},
	}
	body, _ := json.Marshal(responseCertificate)
	resp.SetBody(body)
	clientHttp.EXPECT().DoTimeout(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).SetArg(1, resp).Return(nil)

	storage := mockTypesStorageCertificate.NewMockStorage(ctrl)
	storage.EXPECT().Save(gomock.Any(), gomock.Any()).Times(1).Return([]error{errors.New("error")})
	storage.EXPECT().ID().Times(1).Return("foo")
	storage.EXPECT().Delete(gomock.Any(), gomock.Any()).Times(1).Return(nil)

	as := &AgentService{
		logger:       ctx.Logger,
		stateStorage: storageState,
		httpClient:   clientHttp,
		storages:     certificate.Storages{"foo": storage},
		hookManager:  hook.NewManagerHook(ctx.Logger),
	}
	go as.hookManager.Start()

	err := as.Run(ctx)
	assert.NoError(t, err)
}

func TestAgentService_Run_SuccessStorageDeleteError(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx.MetricsRegister = appProm.NewRegistry(types.NameAgentMetrics, prometheus.NewRegistry())
	state := &types.State{Account: nil, Certificates: types.Certificates{certificateBar}}
	storageState := mockTypesStorageState.NewMockStorage(ctrl)
	storageState.EXPECT().Load().Times(1).Return(state, nil)
	storageState.EXPECT().Save(gomock.Any()).Times(1).Return(nil)

	requester := mockTypes.NewMockRequester(ctrl)
	requester.EXPECT().Fetch().Times(1).Return([]*types.DomainRequest{domainRequestFoo, domainRequestBar, domainRequestNoFound}, nil)
	ctx.Requesters = types.Requesters{"foo": requester}

	clientHttp := mockHttp.NewMockClient(ctrl)
	resp := fasthttp.Response{}
	resp.SetStatusCode(http.StatusOK)
	responseCertificate := appHttp.ResponseCertificatesFromRequests{
		Certificates: types.Certificates{certificateFoo, certificateBar},
		Requests: appHttp.ResponseRequests{
			Found:    []*types.DomainRequest{domainRequestFoo, domainRequestBar},
			NotFound: []*types.DomainRequest{domainRequestNoFound},
		},
	}
	body, _ := json.Marshal(responseCertificate)
	resp.SetBody(body)
	clientHttp.EXPECT().DoTimeout(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).SetArg(1, resp).Return(nil)

	storage := mockTypesStorageCertificate.NewMockStorage(ctrl)
	storage.EXPECT().Save(gomock.Any(), gomock.Any()).Times(1).Return(nil)
	storage.EXPECT().Delete(gomock.Any(), gomock.Any()).Times(1).Return([]error{errors.New("error")})
	storage.EXPECT().ID().Times(1).Return("foo")

	as := &AgentService{
		logger:       ctx.Logger,
		stateStorage: storageState,
		httpClient:   clientHttp,
		storages:     certificate.Storages{"foo": storage},
		hookManager:  hook.NewManagerHook(ctx.Logger),
	}
	go as.hookManager.Start()

	err := as.Run(ctx)
	assert.NoError(t, err)
}

func TestAgentService_Start_Success(t *testing.T) {
	b := bytes.NewBufferString("")
	ctx := appCtx.TestContext(b)
	ctx.LogLevel.Set(slog.LevelDebug)
	ctx.Config.Interval = time.Millisecond * 500
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storage := mockTypesStorageState.NewMockStorage(ctrl)
	storage.EXPECT().Load().AnyTimes().Return(&types.State{}, nil)
	storage.EXPECT().Save(gomock.Any()).AnyTimes().Return(nil)

	ctx.MetricsRegister = appProm.NewRegistry(types.NameAgentMetrics, prometheus.NewRegistry())

	fakeClock := clockwork.NewFakeClockAt(time.Date(1970, time.January, 1, 0, 0, 59, 0, time.UTC))
	as := &AgentService{
		stateStorage: storage,
		clock:        fakeClock,
		hookManager:  hook.NewManagerHook(ctx.Logger),
	}

	go func() {
		err := as.Start(ctx)
		assert.NoError(t, err)
	}()
	fakeClock.BlockUntil(1)
	fakeClock.Advance(ctx.Config.Interval)
	time.Sleep(200 * time.Millisecond)
	ctx.Cancel()
	time.Sleep(200 * time.Millisecond)
	assert.Contains(t, b.String(), "tick received")
	assert.Contains(t, b.String(), "stop asked by app, exiting")
}

func TestAgentService_Start_Failed(t *testing.T) {
	b := bytes.NewBufferString("")
	ctx := appCtx.TestContext(b)
	ctx.LogLevel.Set(slog.LevelDebug)
	ctx.Config.Interval = time.Millisecond * 500
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	storage := mockTypesStorageState.NewMockStorage(ctrl)
	storage.EXPECT().Load().AnyTimes().Return(nil, errors.New("error load"))

	fakeClock := clockwork.NewFakeClockAt(time.Date(1970, time.January, 1, 0, 0, 59, 0, time.UTC))
	as := &AgentService{
		stateStorage: storage,
		clock:        fakeClock,
		hookManager:  hook.NewManagerHook(ctx.Logger),
	}
	go func() {
		err := as.Start(ctx)
		assert.NoError(t, err)
	}()
	fakeClock.BlockUntil(1)
	fakeClock.Advance(ctx.Config.Interval)
	time.Sleep(200 * time.Millisecond)
	ctx.Cancel()
	assert.Contains(t, b.String(), "error load")
}

func TestAgentService_GetRequestManager(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	mgrCfg := config.ManagerConfig{Address: "http://127.0.0.1", TokenJWT: "<PASSWORD>"}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name           string
		domainRequests []*types.DomainRequest
		mockFn         func(clientHttp *mockHttp.MockClient)
		want           appHttp.ResponseCertificatesFromRequests
		wantErr        assert.ErrorAssertionFunc
	}{
		{
			name: "Success",
			domainRequests: []*types.DomainRequest{
				domainRequestFoo,
				domainRequestBar,
				domainRequestNoFound,
			},
			mockFn: func(clientHttp *mockHttp.MockClient) {
				resp := fasthttp.Response{}
				resp.SetStatusCode(http.StatusOK)
				responseCertificate := appHttp.ResponseCertificatesFromRequests{
					Certificates: types.Certificates{certificateFoo, certificateBar},
					Requests: appHttp.ResponseRequests{
						Found:    []*types.DomainRequest{domainRequestFoo, domainRequestBar},
						NotFound: []*types.DomainRequest{domainRequestNoFound},
					},
				}
				body, _ := json.Marshal(responseCertificate)
				resp.SetBody(body)
				clientHttp.EXPECT().DoTimeout(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).SetArg(1, resp).Return(nil)
			},
			want: appHttp.ResponseCertificatesFromRequests{
				Certificates: types.Certificates{certificateFoo, certificateBar},
				Requests: appHttp.ResponseRequests{
					Found:    []*types.DomainRequest{domainRequestFoo, domainRequestBar},
					NotFound: []*types.DomainRequest{domainRequestNoFound},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "ErrorDoRequest",
			domainRequests: []*types.DomainRequest{
				domainRequestFoo,
				domainRequestBar,
				domainRequestNoFound,
			},
			mockFn: func(clientHttp *mockHttp.MockClient) {
				clientHttp.EXPECT().DoTimeout(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(errors.New("fail"))
			},
			want:    appHttp.ResponseCertificatesFromRequests{},
			wantErr: assert.Error,
		},
		{
			name: "ErrorStatusCode",
			domainRequests: []*types.DomainRequest{
				domainRequestFoo,
				domainRequestBar,
				domainRequestNoFound,
			},
			mockFn: func(clientHttp *mockHttp.MockClient) {
				resp := fasthttp.Response{}
				resp.SetStatusCode(http.StatusInternalServerError)
				clientHttp.EXPECT().DoTimeout(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).SetArg(1, resp).Return(nil)
			},
			want:    appHttp.ResponseCertificatesFromRequests{},
			wantErr: assert.Error,
		},
		{
			name: "FailUnmarshalResponse",
			domainRequests: []*types.DomainRequest{
				domainRequestFoo,
			},
			mockFn: func(clientHttp *mockHttp.MockClient) {
				resp := fasthttp.Response{}
				resp.SetStatusCode(http.StatusOK)
				resp.SetBody([]byte("{]"))
				clientHttp.EXPECT().DoTimeout(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).SetArg(1, resp).Return(nil)
			},
			want:    appHttp.ResponseCertificatesFromRequests{},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientHttp := mockHttp.NewMockClient(ctrl)
			tt.mockFn(clientHttp)
			as := &AgentService{
				httpClient:    clientHttp,
				logger:        ctx.Logger,
				managerConfig: mgrCfg,
			}
			got, err := as.GetRequestManager(tt.domainRequests)
			if !tt.wantErr(t, err, fmt.Sprintf("GetRequestManager(%v)", tt.domainRequests)) {
				return
			}
			assert.Equalf(t, tt.want, got, "GetRequestManager(%v)", tt.domainRequests)
		})
	}
}

func TestAgentService_initMetrics_Success(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	registry := appProm.NewRegistry(types.NameServerMetrics, prometheus.NewRegistry())
	cert1 := &types.Certificate{Identifier: "foo", Main: "example.com", Domains: types.Domains{"example.com"}}
	metricsRegistry := mockPrometheus.NewMockRegistry(ctrl)

	gomock.InOrder(
		metricsRegistry.EXPECT().CreateGaugeCertificate(gomock.Any()).Times(1).Return(registry.CreateGaugeCertificate(cert1)),
		metricsRegistry.EXPECT().MustAddGaugeCertificate(gomock.Eq(cert1.Identifier), gomock.Any()).Times(1),
		metricsRegistry.EXPECT().FormatName(gomock.Any()).Times(1).Return(runCountMetric),
		metricsRegistry.EXPECT().MustAddCounter(gomock.Any(), gomock.Any()).Times(1),

		metricsRegistry.EXPECT().FormatName(gomock.Any()).Times(1).Return(fetchErrorMetric),
		metricsRegistry.EXPECT().MustAddGauge(gomock.Any(), gomock.Any()).Times(1),

		metricsRegistry.EXPECT().FormatName(gomock.Any()).Times(1).Return(domainRequestsMetric),
		metricsRegistry.EXPECT().MustAddGauge(gomock.Any(), gomock.Any()).Times(1),

		metricsRegistry.EXPECT().FormatName(gomock.Any()).Times(1).Return(domainRequestsCertFoundMetric),
		metricsRegistry.EXPECT().MustAddGauge(gomock.Any(), gomock.Any()).Times(1),

		metricsRegistry.EXPECT().FormatName(gomock.Any()).Times(1).Return(domainRequestsCertNotFoundMetric),
		metricsRegistry.EXPECT().MustAddGauge(gomock.Any(), gomock.Any()).Times(1),
	)
	ctx.MetricsRegister = metricsRegistry

	as := &AgentService{metricsInit: false}
	as.initMetrics(ctx, &types.State{Certificates: types.Certificates{cert1}})
}

func TestAgentService_initMetrics_SuccessAlreadyInit(t *testing.T) {
	ctx := appCtx.TestContext(nil)
	as := &AgentService{metricsInit: true}
	as.initMetrics(ctx, &types.State{})
}
