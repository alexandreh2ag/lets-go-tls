package requester

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alexandreh2ag/lets-go-tls/config"
	"github.com/alexandreh2ag/lets-go-tls/context"
	appHttp "github.com/alexandreh2ag/lets-go-tls/http"
	"github.com/alexandreh2ag/lets-go-tls/requester"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/go-playground/validator/v10"
	"github.com/hashicorp/go-multierror"
	"github.com/mitchellh/mapstructure"
	"github.com/valyala/fasthttp"
	"log/slog"
	"net/http"
	"sync"
	"time"
)

const (
	AgentKey = "agent"
)

func init() {
	requester.TypeRequesterMapping[AgentKey] = createAgentProvider
}

var _ types.Requester = &agent{}

type ConfigAgent struct {
	Addresses []string `mapstructure:"addresses" validate:"required,min=1,dive,http_url"`
}

type agent struct {
	id         string
	httpClient appHttp.Client
	logger     *slog.Logger
	addresses  []string
}

func (a *agent) ID() string {
	return a.id
}

func (a *agent) Fetch() ([]*types.DomainRequest, error) {
	merr := &multierror.Error{}
	wg := sync.WaitGroup{}
	lock := sync.Mutex{}
	domains := []*types.DomainRequest{}

	for _, address := range a.addresses {
		wg.Add(1)
		go func() {
			defer wg.Done()
			domainsAgent, err := a.FetchAgent(address)
			if err != nil {
				formatError := fmt.Errorf("agent (%s) failed to fetch with: %v", a.id, err)
				a.logger.Error(formatError.Error())
				merr = multierror.Append(merr, fmt.Errorf("agent (%s) failed to fetch with: %v", a.id, formatError))
			}
			lock.Lock()
			defer lock.Unlock()
			domains = append(domains, domainsAgent...)
		}()
	}
	wg.Wait()

	for _, domain := range domains {
		domain.Requester = a
	}
	return domains, merr.ErrorOrNil()
}

func (a *agent) FetchAgent(address string) ([]*types.DomainRequest, error) {
	domains := []*types.DomainRequest{}

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

	uri := fmt.Sprintf("%s%s", address, appHttp.GetApiPrefix(appHttp.AgentApiRequests))
	req.Header.SetMethod(http.MethodGet)
	req.SetRequestURI(uri)

	err := a.httpClient.DoTimeout(req, resp, 1*time.Second)

	if err != nil {
		return domains, err
	}

	if resp.StatusCode() != http.StatusOK {
		return domains, errors.New(fmt.Sprintf("response have invalid status code %v", resp.StatusCode()))
	}

	err = json.Unmarshal(resp.Body(), &domains)
	if err != nil {
		return domains, err
	}

	return domains, nil
}

func createAgentProvider(ctx context.Context, cfg config.RequesterConfig) (types.Requester, error) {
	instanceConfig := ConfigAgent{}
	err := mapstructure.Decode(cfg.Config, &instanceConfig)
	if err != nil {
		return nil, err
	}

	validate := validator.New()
	err = validate.Struct(instanceConfig)
	if err != nil {
		return nil, err
	}

	instance := &agent{id: cfg.Id}
	instance.logger = ctx.GetLogger()
	instance.addresses = instanceConfig.Addresses
	instance.httpClient = ctx.GetHttpClient()

	return instance, nil
}
