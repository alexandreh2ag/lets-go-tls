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
	traefikConfigDynamic "github.com/traefik/traefik/v3/pkg/config/dynamic"
	traefikHttpmuxer "github.com/traefik/traefik/v3/pkg/muxer/http"
	"github.com/valyala/fasthttp"
	"log/slog"
	"net/http"
	"slices"
	"sync"
	"time"
)

const (
	TraefikV2Key = "traefikV2"
	TraefikV3Key = "traefikV3"
)

type TraefikResponseHttpRouter struct {
	Rule string                       `json:"rule,omitempty"`
	TLS  TraefikResponseHttpRouterTLS `json:"tls,omitempty"`
}

type TraefikResponseHttpRouterTLS struct {
	Domains []TraefikResponseHttpRouterTLSDomain `json:"domains,omitempty"`
}

type TraefikResponseHttpRouterTLSDomain struct {
	Main string   `json:"main,omitempty"`
	SANS []string `json:"sans,omitempty"`
}

func init() {
	requester.TypeRequesterMapping[TraefikV2Key] = createTraefikV2Provider
	requester.TypeRequesterMapping[TraefikV3Key] = createTraefikV2Provider
}

var _ types.Requester = &traefik{}

type ConfigTraefik struct {
	Addresses []string `mapstructure:"addresses" validate:"required,min=1"`
}

type traefik struct {
	id         string
	httpClient appHttp.Client
	logger     *slog.Logger
	addresses  []string
}

func (t *traefik) ID() string {
	return t.id
}

func (t *traefik) Fetch() ([]*types.DomainRequest, error) {
	merr := &multierror.Error{}
	wg := sync.WaitGroup{}
	lock := sync.Mutex{}
	domains := []*types.DomainRequest{}

	for _, address := range t.addresses {
		wg.Add(1)
		go func() {
			defer wg.Done()
			domainsAgent, err := t.FetchInstance(address)
			if err != nil {
				formatError := fmt.Errorf("requester (%s) failed to fetch with: %v", t.id, err)
				t.logger.Error(formatError.Error())
				merr = multierror.Append(merr, fmt.Errorf("requester (%s) failed to fetch with: %v", t.id, formatError))
			}

			lock.Lock()
			defer lock.Unlock()
			domains = append(domains, domainsAgent...)
		}()
	}
	wg.Wait()

	domains = slices.DeleteFunc(domains, func(item *types.DomainRequest) bool {
		if item.IsIP() {
			return true
		}
		return false
	})

	for _, domain := range domains {
		domain.Requester = t
	}
	return domains, merr.ErrorOrNil()
}

func (t *traefik) FetchInstance(address string) ([]*types.DomainRequest, error) {
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
	uri := fmt.Sprintf("%s%s", address, "/api/http/routers")
	req.Header.SetMethod(http.MethodGet)
	req.SetRequestURI(uri)

	err := t.httpClient.DoTimeout(req, resp, 1*time.Second)

	if err != nil {
		return domains, err
	}

	if resp.StatusCode() != http.StatusOK {
		return domains, errors.New(fmt.Sprintf("response have invalid status code %v", resp.StatusCode()))
	}
	routers := []traefikConfigDynamic.Router{}
	err = json.Unmarshal(resp.Body(), &routers)
	if err != nil {
		return domains, err
	}

	return t.FormatRouters(routers)
}

func (t *traefik) FormatRouters(routers []traefikConfigDynamic.Router) ([]*types.DomainRequest, error) {
	requests := []*types.DomainRequest{}

	for _, router := range routers {
		domains := types.Domains{}
		if router.TLS == nil {
			continue
		}

		if len(router.TLS.Domains) > 0 {
			for _, domain := range router.TLS.Domains {
				domains = append(domains, types.Domain(domain.Main))
				for _, san := range domain.SANs {
					domains = append(domains, types.Domain(san))
				}
			}

		} else {
			parsedDomains, err := traefikHttpmuxer.ParseDomains(router.Rule)
			if err != nil {
				return requests, err
			}
			if len(parsedDomains) == 0 {
				continue
			}
			for _, parsedDomain := range parsedDomains {
				domains = append(domains, types.Domain(parsedDomain))
			}
		}
		requests = append(requests, &types.DomainRequest{Domains: domains})
	}
	return requests, nil
}

func createTraefikV2Provider(ctx context.Context, cfg config.RequesterConfig) (types.Requester, error) {
	instanceConfig := ConfigTraefik{}
	err := mapstructure.Decode(cfg.Config, &instanceConfig)
	if err != nil {
		return nil, err
	}

	validate := validator.New()
	err = validate.Struct(instanceConfig)
	if err != nil {
		return nil, err
	}

	instance := &traefik{id: cfg.Id}
	instance.logger = ctx.GetLogger()
	instance.addresses = instanceConfig.Addresses
	instance.httpClient = ctx.GetHttpClient()

	return instance, nil
}
