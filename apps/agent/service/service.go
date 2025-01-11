package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/config"
	appCtx "github.com/alexandreh2ag/lets-go-tls/apps/agent/context"
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/requester"
	"github.com/alexandreh2ag/lets-go-tls/hook"
	appHttp "github.com/alexandreh2ag/lets-go-tls/http"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/alexandreh2ag/lets-go-tls/types/storage/certificate"
	typesStorageState "github.com/alexandreh2ag/lets-go-tls/types/storage/state"
	"github.com/hashicorp/go-multierror"
	"github.com/jonboulle/clockwork"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/valyala/fasthttp"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

const (
	runCountMetric                   = "run_count"
	fetchErrorMetric                 = "fetch_error_number"
	domainRequestsMetric             = "domains_requests_number"
	domainRequestsCertFoundMetric    = "domains_requests_certificate_found_number"
	domainRequestsCertNotFoundMetric = "domains_requests_certificate_not_found_number"
)

var _ Service = &AgentService{}

type (
	Service interface {
		Start(ctx *appCtx.AgentContext) error
	}
)

type AgentService struct {
	stateStorage  typesStorageState.Storage
	storages      certificate.Storages
	httpClient    appHttp.Client
	logger        *slog.Logger
	managerConfig config.ManagerConfig

	clock clockwork.Clock

	hookManager *hook.ManagerHook

	mutexMetrics sync.Mutex
	metricsInit  bool
}

func NewService(ctx *appCtx.AgentContext) *AgentService {
	return &AgentService{
		managerConfig: ctx.Config.Manager,
		stateStorage:  ctx.StateStorage,
		storages:      ctx.Storages,
		httpClient:    ctx.GetHttpClient(),
		logger:        ctx.GetLogger(),
		clock:         clockwork.NewRealClock(),
		hookManager:   hook.NewManagerHook(ctx.Logger),
	}
}

func (as *AgentService) Start(ctx *appCtx.AgentContext) error {
	go as.hookManager.Start()

	tickFunc := func() {
		err := as.Run(ctx)
		if err != nil {
			ctx.Logger.Error(err.Error())
		}
	}

	tickFunc()
	ticker := as.clock.NewTicker(ctx.Config.Interval)
	defer ticker.Stop()
	ctx.Logger.Debug("wait for tick")
	for {
		select {
		case <-ticker.Chan():
			ctx.Logger.Debug("tick received")
			tickFunc()
		case <-ctx.Done():
			ctx.Logger.Info(fmt.Sprintf("stop asked by app, exiting..."))
			return nil
		}

	}
}

func (as *AgentService) Run(ctx *appCtx.AgentContext) error {
	state, errLoad := as.stateStorage.Load()
	if errLoad != nil {
		return fmt.Errorf("failed to load state: %v", errLoad)
	}
	as.initMetrics(ctx, state)

	ctx.MetricsRegister.MustGetCounter(runCountMetric).Inc()
	domainsRequests, errFetch := requester.FetchRequests(ctx.Requesters)
	if merr, ok := errFetch.(*multierror.Error); ok {
		for _, errRequester := range merr.Errors {
			ctx.Logger.Error(errRequester.Error())
		}
		ctx.MetricsRegister.MustGetGauge(fetchErrorMetric).Set(1)
	} else {
		ctx.MetricsRegister.MustGetGauge(fetchErrorMetric).Set(0)
	}

	ctx.MetricsRegister.MustGetGauge(domainRequestsMetric).Set(float64(len(domainsRequests)))
	if len(domainsRequests) > 0 {
		managerResponse, errManager := as.GetRequestManager(domainsRequests)
		if errManager != nil {
			return errManager
		}

		ctx.MetricsRegister.MustGetGauge(domainRequestsCertFoundMetric).Set(float64(len(managerResponse.Requests.Found)))
		ctx.MetricsRegister.MustGetGauge(domainRequestsCertNotFoundMetric).Set(float64(len(managerResponse.Requests.NotFound)))
		ctx.Logger.Debug(fmt.Sprintf("found %d certificates", len(managerResponse.Certificates)))
		ctx.Logger.Debug(fmt.Sprintf("%d domains requests found", len(managerResponse.Requests.Found)))
		ctx.Logger.Debug(fmt.Sprintf("%d domains requests not found", len(managerResponse.Requests.NotFound)))
		if len(managerResponse.Requests.NotFound) > 0 {
			as.logger.Warn(fmt.Sprintf("some domains requests not found (%d)", len(managerResponse.Requests.NotFound)))
		}

		for _, certManager := range managerResponse.Certificates {
			certificateState := state.Certificates.GetCertificate(certManager.Identifier)
			if certificateState != nil {
				certificateState.Main = certManager.Main
				certificateState.Domains = certManager.Domains
				certificateState.Key = certManager.Key
				certificateState.Certificate = certManager.Certificate
				certificateState.ExpirationDate = certManager.ExpirationDate
			} else {
				state.Certificates = append(state.Certificates, certManager)
			}
		}
	}

	// find unused certificates only if fetch succeeded
	unusedCertificates := types.Certificates{}
	if errFetch == nil {
		unusedCertificates = state.Certificates.UnusedCertificates(domainsRequests)
		ctx.Logger.Debug(fmt.Sprintf("found %d unused certificates", len(unusedCertificates)))
	}

	stateDeleteUnusedCertificates := false
	for _, storage := range as.storages {
		ctx.Logger.Debug("save certificates in storage")
		errStorageSave := storage.Save(state.Certificates, as.hookManager.GetHookChan())
		if len(errStorageSave) > 0 {
			for _, err := range errStorageSave {
				ctx.Logger.Error(fmt.Sprintf("storage %s, failed to save certificates: %v", storage.ID(), err))
			}
		}

		errStorageDelete := storage.Delete(unusedCertificates, as.hookManager.GetHookChan())
		if len(errStorageDelete) > 0 {
			stateDeleteUnusedCertificates = true
			for _, err := range errStorageDelete {
				ctx.Logger.Error(fmt.Sprintf("storage %s, failed to delete unused certificates: %v", storage.ID(), err))
			}
		}
	}

	if !stateDeleteUnusedCertificates && len(unusedCertificates) > 0 {
		ctx.Logger.Debug("delete unused certificates in state")
		state.Certificates = state.Certificates.Deletes(unusedCertificates)
	}

	as.hookManager.RunHooks()

	ctx.Logger.Debug("save state")
	return as.stateStorage.Save(state)
}

func (as *AgentService) GetRequestManager(domainRequests []*types.DomainRequest) (appHttp.ResponseCertificatesFromRequests, error) {
	response := appHttp.ResponseCertificatesFromRequests{}
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	req.SetConnectionClose()
	defer func() {
		if req != nil {
			fasthttp.ReleaseRequest(req)
		}
		if resp != nil {
			fasthttp.ReleaseResponse(resp)
		}
	}()

	body, err := json.Marshal(domainRequests)
	if err != nil {
		return response, fmt.Errorf("failed to marshal domains requests: %w", err)
	}

	uri := fmt.Sprintf("%s%s", as.managerConfig.Address, appHttp.GetApiPrefix(appHttp.ServerApiGetCertificates))
	req.Header.SetMethod(http.MethodPost)
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", as.managerConfig.TokenJWT))
	req.Header.Add("Content-Type", "application/json")
	req.SetRequestURI(uri)
	req.SetBody(body)

	err = as.httpClient.DoTimeout(req, resp, 1*time.Second)

	if err != nil {
		return response, err
	}

	if resp.StatusCode() != http.StatusOK {
		return response, errors.New(fmt.Sprintf("response have invalid status code %v", resp.StatusCode()))
	}

	err = json.Unmarshal(resp.Body(), &response)
	if err != nil {
		return response, err
	}
	return response, nil
}

func (as *AgentService) initMetrics(ctx *appCtx.AgentContext, state *types.State) {
	if as.metricsInit {
		return
	}

	defer func() {
		as.metricsInit = true
	}()

	registry := ctx.MetricsRegister
	for _, cert := range state.Certificates {
		gauge := ctx.MetricsRegister.CreateGaugeCertificate(cert)
		gauge.Set(float64(cert.ExpirationDate.Unix()))
		ctx.MetricsRegister.MustAddGaugeCertificate(cert.Identifier, gauge)
	}

	countRunMetrics := prometheus.NewCounter(prometheus.CounterOpts{
		Name: registry.FormatName(runCountMetric),
		Help: "Count of run",
	})
	registry.MustAddCounter(runCountMetric, countRunMetrics)

	gaugeFetchErrorMetrics := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: registry.FormatName(fetchErrorMetric),
		Help: "Number of error for fetch process",
	})
	registry.MustAddGauge(fetchErrorMetric, gaugeFetchErrorMetrics)

	gaugeDomainRequestsCountMetrics := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: registry.FormatName(domainRequestsMetric),
		Help: "Number of domains requests",
	})
	gaugeDomainRequestsCountMetrics.Set(0)
	ctx.MetricsRegister.MustAddGauge(domainRequestsMetric, gaugeDomainRequestsCountMetrics)

	gaugeDomainRequestsCertFoundCountMetrics := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: registry.FormatName(domainRequestsCertFoundMetric),
		Help: "Number of domains requests certificate found",
	})
	gaugeDomainRequestsCertFoundCountMetrics.Set(0)
	ctx.MetricsRegister.MustAddGauge(domainRequestsCertFoundMetric, gaugeDomainRequestsCertFoundCountMetrics)

	gaugeDomainRequestsCertNotFoundCountMetrics := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: registry.FormatName(domainRequestsCertNotFoundMetric),
		Help: "Number of domains requests certificate not found",
	})
	gaugeDomainRequestsCertFoundCountMetrics.Set(0)
	ctx.MetricsRegister.MustAddGauge(domainRequestsCertNotFoundMetric, gaugeDomainRequestsCertNotFoundCountMetrics)

}
