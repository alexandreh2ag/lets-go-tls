package manager

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/acme"
	appCtx "github.com/alexandreh2ag/lets-go-tls/apps/server/context"
	"github.com/alexandreh2ag/lets-go-tls/types"
	typesAcme "github.com/alexandreh2ag/lets-go-tls/types/acme"
	typesStorageState "github.com/alexandreh2ag/lets-go-tls/types/storage/state"
	"github.com/eko/gocache/lib/v4/store"
	legoCertificate "github.com/go-acme/lego/v4/certificate"
	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	"github.com/jonboulle/clockwork"
	"github.com/prometheus/client_golang/prometheus"
	"strings"
	"sync"
	"time"
)

const (
	CacheProcessLockKey = "manager_run_process_lock"

	runCountMetric        = "run_count"
	fetchErrorMetric      = "fetch_error_number"
	obtainCertErrorMetric = "obtain_certificate_error_number"
)

var _ Manager = &CertifierManager{}

type Manager interface {
	Start(ctx *appCtx.ServerContext) error
}

type CertifierManager struct {
	ephemeralID  string
	stateStorage typesStorageState.Storage
	resolvers    types.Resolvers

	clock clockwork.Clock

	metricsInit  bool
	mutexMetrics sync.Mutex
}

func NewManager(stateStorage typesStorageState.Storage) *CertifierManager {
	return &CertifierManager{stateStorage: stateStorage, clock: clockwork.NewRealClock()}
}

func CreateManager(ctx *appCtx.ServerContext) (Manager, error) {
	mgr := NewManager(ctx.StateStorage)

	mgr.ephemeralID = uuid.NewString()
	return mgr, nil
}

func (cm *CertifierManager) Start(ctx *appCtx.ServerContext) error {
	tickFunc := func() {
		err := cm.Run(ctx)
		if err != nil {
			ctx.Logger.Error(err.Error())
		}
	}

	tickFunc()
	ticker := cm.clock.NewTicker(ctx.Config.Interval)
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

func (cm *CertifierManager) Run(ctx *appCtx.ServerContext) error {
	var errCreateAccountError, errCreateResolvers error
	state, errLoad := cm.stateStorage.Load()
	if errLoad != nil {
		return fmt.Errorf("failed to load state: %v", errLoad)
	}

	cm.initMetrics(ctx, state)

	hasLock, errLock := cm.obtainLock(ctx)
	if errLock != nil {
		return fmt.Errorf("unable to lock manager process with: %v", errLock)
	}
	if !hasLock {
		ctx.Logger.Info("tick skipped due process is already running")
		return nil
	}
	defer func() {
		errLock = cm.releaseLock(ctx)
		if errLock != nil {
			ctx.Logger.Error(fmt.Sprintf("unable to unlock manager process with: %v", errLock))
		}
	}()

	ctx.MetricsRegister.MustGetCounter(runCountMetric).Inc()
	// Create new account
	if state.Account == nil || state.Account.Key == nil {
		state.Account, errCreateAccountError = typesAcme.NewAccount(ctx.Config.Acme.Email)
		if errCreateAccountError != nil {
			return fmt.Errorf("failed to create account: %s", errCreateAccountError)
		}
	}

	if cm.resolvers == nil {
		cm.resolvers, errCreateResolvers = acme.CreateResolvers(ctx, state.Account)
		if errCreateResolvers != nil {
			return errCreateResolvers
		}
	}

	defaultResolver := cm.resolvers[types.DefaultKey]
	// Init account typesAcme
	errRegisterAccountError := acme.RegisterAccount(state, cm.stateStorage, defaultResolver)
	if errRegisterAccountError != nil {
		return errRegisterAccountError
	}

	domainsRequests, errFetch := cm.FetchRequests(ctx)

	if len(errFetch) > 0 {
		ctx.MetricsRegister.MustGetGauge(fetchErrorMetric).Set(1)
		ctx.Logger.Error(fmt.Sprintf("unable to fetch requests: %v", errFetch))
	} else {
		ctx.MetricsRegister.MustGetGauge(fetchErrorMetric).Set(0)
	}

	cm.MatchingRequests(ctx, state, domainsRequests)

	// Run acme challenge for new cert or renew cert need
	errObtainCerts := cm.ObtainCertificates(ctx, state)
	if errObtainCerts.ErrorOrNil() != nil {
		ctx.MetricsRegister.MustGetGauge(obtainCertErrorMetric).Set(1)
		for _, errObtainCert := range errObtainCerts.WrappedErrors() {
			ctx.Logger.Error(errObtainCert.Error())
		}
		ctx.Logger.Error("failed to obtain certificates")
	} else {
		ctx.MetricsRegister.MustGetGauge(obtainCertErrorMetric).Set(0)
	}

	// remove UnusedAt when a certificate is reuse again
	// remove unused certificates when retention expired or mark for retention and only if errFetch is nil
	if len(errFetch) == 0 {
		ctx.Logger.Info(fmt.Sprintf("clean unused flag when certificates have been reuse agin"))
		cm.MarkCertificatesAsReused(state.Certificates, domainsRequests)

		ctx.Logger.Info(fmt.Sprintf("clean up unused certificates"))
		state.Certificates = cm.CleanUnusedCertificates(ctx, state.Certificates, domainsRequests)
	}

	cm.updateMetrics(ctx, state)

	return cm.stateStorage.Save(state)
}

func (cm *CertifierManager) CleanUnusedCertificates(ctx *appCtx.ServerContext, certificates types.Certificates, domainsRequests []*types.DomainRequest) types.Certificates {
	toDeleteCertificates := types.Certificates{}
	unusedCertificates := certificates.UnusedCertificates(domainsRequests)
	for _, certificate := range unusedCertificates {
		if certificate.UnusedAt.IsZero() {
			ctx.Logger.Info(fmt.Sprintf("certificate %s is detected unused", certificate.Identifier))
			certificate.UnusedAt = cm.clock.Now()
		} else if certificate.UnusedAt.Before(time.Now().Add(ctx.Config.UnusedRetentionDuration * -1)) {
			ctx.Logger.Info(fmt.Sprintf("certificate %s will be deleted", certificate.Identifier))
			toDeleteCertificates = append(toDeleteCertificates, certificate)
		}
	}
	return certificates.Deletes(toDeleteCertificates)
}

func (cm *CertifierManager) MarkCertificatesAsReused(certificates types.Certificates, domainsRequests []*types.DomainRequest) {
	for _, certificate := range certificates.UsedCertificates(domainsRequests) {
		certificate.UnusedAt = time.Time{}
	}
}

func (cm *CertifierManager) ObtainCertificates(ctx *appCtx.ServerContext, state *types.State) *multierror.Error {
	cfgAcme := ctx.Config.Acme
	var err error
	merr := &multierror.Error{}
	for _, certificate := range state.Certificates {
		var certAcme *legoCertificate.Resource
		resolver := cm.resolvers.FindResolver(certificate)

		if resolver.TypeChallenge() == typesAcme.TypeHTTP01 && certificate.Domains.ContainsWildcard() {
			certificate.ObtainFailCount++
			certificate.ObtainFailDate = cm.clock.Now()
			merr = multierror.Append(
				merr,
				fmt.Errorf(
					"unable to obtain wildcard certificate without ACME DNS challange %s",
					certificate.Identifier,
				),
			)
			continue
		}

		if !certificate.ObtainFailDate.IsZero() && certificate.ObtainFailCount >= cfgAcme.MaxAttempt &&
			cm.clock.Now().Before(certificate.ObtainFailDate.Add(cfgAcme.DelayFailed)) {
			ctx.Logger.Warn(fmt.Sprintf("skip certificate %s due to max obtain fail reach", certificate.Identifier))
			continue
		}

		if certificate.Key == nil || certificate.Certificate == nil {
			request := legoCertificate.ObtainRequest{
				Domains:    certificate.Domains.ToStringSlice(),
				Bundle:     true,
				MustStaple: false,
			}
			ctx.Logger.Info(fmt.Sprintf("obtain certificate %s (%v)", certificate.Identifier, certificate.Domains.ToStringSlice()))
			certAcme, err = resolver.Obtain(request)
		} else if certificate.ExpirationDate.Before(time.Now().Add(cfgAcme.RenewPeriod * -1)) {
			certRes := legoCertificate.Resource{
				Domain:      string(certificate.Domains[0]),
				PrivateKey:  certificate.Key,
				Certificate: certificate.Certificate,
			}
			options := &legoCertificate.RenewOptions{Bundle: true, MustStaple: false}
			ctx.Logger.Info(fmt.Sprintf("renew certificate %s (%v)", certificate.Identifier, certificate.Domains.ToStringSlice()))
			certAcme, err = resolver.RenewWithOptions(certRes, options)
		} else {
			ctx.Logger.Debug(fmt.Sprintf("nothing to do for certificate %s", certificate.Identifier))
			continue
		}
		if err != nil {
			certificate.ObtainFailCount++
			certificate.ObtainFailDate = cm.clock.Now()
			merr = multierror.Append(merr, fmt.Errorf("unable to obtain/renew certificate %s : %v", certificate.Identifier, err))
			continue
		}
		certificate.Key = certAcme.PrivateKey
		certificate.Certificate = certAcme.Certificate
		block, _ := pem.Decode(certificate.Certificate)
		if block == nil {
			certificate.ObtainFailCount++
			certificate.ObtainFailDate = cm.clock.Now()
			merr = multierror.Append(merr, fmt.Errorf("failed to decode certificate for: %s", certificate.Identifier))
			continue
		}
		cert, errParse := x509.ParseCertificate(block.Bytes)
		if errParse != nil {
			certificate.ObtainFailCount++
			certificate.ObtainFailDate = cm.clock.Now()
			merr = multierror.Append(merr, fmt.Errorf("failed to parse certificate for %s: %v", certificate.Identifier, errParse))
			continue
		}
		certificate.ExpirationDate = cert.NotAfter
		certificate.ObtainFailCount = 0
		certificate.ObtainFailDate = time.Time{}
	}
	return merr
}

func (cm *CertifierManager) FetchRequests(ctx *appCtx.ServerContext) ([]*types.DomainRequest, map[string]error) {
	domainsRequests := []*types.DomainRequest{}
	wg := sync.WaitGroup{}
	lock := sync.Mutex{}
	errs := map[string]error{}
	for _, requester := range ctx.Requesters {
		wg.Add(1)
		go func() {
			defer wg.Done()
			requests, err := requester.Fetch()
			if err != nil {
				errs[requester.ID()] = err
				return
			}
			lock.Lock()
			defer lock.Unlock()
			domainsRequests = append(domainsRequests, requests...)
		}()
	}
	wg.Wait()
	types.SortDomainsRequests(domainsRequests)
	return domainsRequests, errs
}

func (cm *CertifierManager) MatchingRequests(ctx *appCtx.ServerContext, state *types.State, domainsRequests []*types.DomainRequest) {
	// check domainRequest is already in typesStorageState.Certificates or add it
	for _, request := range domainsRequests {
		cert := state.Certificates.Match(request, false)
		if cert == nil {
			cert = &types.Certificate{Domains: request.Domains, Main: string(request.Domains[0])}
			// generate name
			baseIdentifier := strings.ReplaceAll(cert.Main, "*", "wildcard")
			i := 0
			cert.Identifier = fmt.Sprintf("%s-%v", baseIdentifier, i)
			for !state.Certificates.CheckIdentifierUnique(cert.Identifier) {
				cert.Identifier = fmt.Sprintf("%s-%v", baseIdentifier, i)
				i++
			}
			ctx.Logger.Info(fmt.Sprintf("create new certificate %s (%v)", cert.Identifier, cert.Domains))
			state.Certificates = append(state.Certificates, cert)
			cm.registerNewCertificateMetrics(ctx, cert)
		}

	}
}

func (cm *CertifierManager) obtainLock(ctx *appCtx.ServerContext) (bool, error) {
	id, err := ctx.Cache.Get(context.Background(), CacheProcessLockKey)
	if err != nil {
		if _, ok := err.(*store.NotFound); !ok {
			return false, err
		}
	}
	if id != "" && id != cm.ephemeralID {
		return false, nil
	}
	err = ctx.Cache.Set(context.Background(), CacheProcessLockKey, cm.ephemeralID, func(o *store.Options) {
		o.Expiration = ctx.Config.LockDuration
	})
	if err != nil {
		return false, err
	}
	time.Sleep(time.Millisecond * 500)

	id, err = ctx.Cache.Get(context.Background(), CacheProcessLockKey)
	if err != nil {
		if _, ok := err.(*store.NotFound); !ok {
			return false, err
		}
	}
	if id != "" && id != cm.ephemeralID {
		return false, nil
	}

	return true, nil
}

func (cm *CertifierManager) releaseLock(ctx *appCtx.ServerContext) error {
	return ctx.Cache.Delete(context.Background(), CacheProcessLockKey)
}

func (cm *CertifierManager) initMetrics(ctx *appCtx.ServerContext, state *types.State) {
	if cm.metricsInit {
		return
	}

	defer func() {
		cm.metricsInit = true
	}()

	registry := ctx.MetricsRegister
	for _, certificate := range state.Certificates {
		cm.registerNewCertificateMetrics(ctx, certificate)
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

	gaugeObtainCertErrorMetrics := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: registry.FormatName(obtainCertErrorMetric),
		Help: "Number of error for obtain certificate process",
	})
	gaugeObtainCertErrorMetrics.Set(0)
	ctx.MetricsRegister.MustAddGauge(obtainCertErrorMetric, gaugeObtainCertErrorMetrics)

}

func (cm *CertifierManager) updateMetrics(ctx *appCtx.ServerContext, state *types.State) {
	registry := ctx.MetricsRegister
	cm.mutexMetrics.Lock()
	// clean metrics for deleted certificate
	for certIdentifier, gauge := range registry.GetGaugeCertificates() {
		if certificate := state.Certificates.GetCertificate(certIdentifier); certificate != nil {
			gauge.Set(float64(certificate.ExpirationDate.Unix()))
		} else {
			registry.MustDeleteGaugeCertificate(certIdentifier)
		}
	}
	cm.mutexMetrics.Unlock()

}

func (cm *CertifierManager) registerNewCertificateMetrics(ctx *appCtx.ServerContext, certificate *types.Certificate) {
	gauge := ctx.MetricsRegister.CreateGaugeCertificate(certificate)
	gauge.Set(float64(certificate.ExpirationDate.Unix()))
	ctx.MetricsRegister.MustAddGaugeCertificate(certificate.Identifier, gauge)
}
