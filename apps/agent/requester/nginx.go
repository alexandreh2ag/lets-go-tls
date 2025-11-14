package requester

import (
	"log/slog"

	"github.com/alexandreh2ag/lets-go-tls/config"
	"github.com/alexandreh2ag/lets-go-tls/context"
	"github.com/alexandreh2ag/lets-go-tls/mapstructure"
	nginxParser "github.com/alexandreh2ag/lets-go-tls/nginx"
	"github.com/alexandreh2ag/lets-go-tls/requester"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/go-playground/validator/v10"
)

const (
	NginxKey = "nginx"
)

func init() {
	requester.TypeRequesterMapping[NginxKey] = createNginxProvider
}

var _ types.Requester = &nginx{}

type ConfigNginx struct {
	NginxCfgPath string `mapstructure:"nginx_cfg_path" validate:"required"`
}

type nginx struct {
	id  string
	cfg ConfigNginx

	logger *slog.Logger
}

func (n nginx) ID() string {
	return n.id
}

func (n nginx) Fetch() ([]*types.DomainRequest, error) {
	domains := []*types.DomainRequest{}
	vhostConfigs, err := nginxParser.ParseConfig(n.logger, n.cfg.NginxCfgPath)
	if err != nil {
		return domains, err
	}

	for _, vhostConfig := range vhostConfigs {
		domains = append(domains, &types.DomainRequest{Domains: vhostConfig.ServerName, Requester: n})
	}

	return domains, nil
}

func createNginxProvider(ctx context.Context, cfg config.RequesterConfig) (types.Requester, error) {
	instanceConfig := ConfigNginx{}
	err := mapstructure.Decode(cfg.Config, &instanceConfig)
	if err != nil {
		return nil, err
	}

	validate := validator.New()
	err = validate.Struct(instanceConfig)
	if err != nil {
		return nil, err
	}

	instance := &nginx{id: cfg.Id, cfg: instanceConfig, logger: ctx.GetLogger()}

	return instance, nil
}
