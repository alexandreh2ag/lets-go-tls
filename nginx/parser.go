package nginx

import (
	"fmt"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/tufanbarisyildirim/gonginx/parser"
	"log/slog"
	"regexp"
)

type VhostConfig struct {
	ServerName types.Domains
	KeyPath    string
	CertPath   string
}

type VhostConfigs []VhostConfig

func isValidDomainName(domain string) bool {
	regex := `^(?i)[a-z0-9]+(-[a-z0-9]+)*(\.[a-z0-9]+(-[a-z0-9]+)*)*$`
	matched, _ := regexp.MatchString(regex, domain)
	return matched
}

func isValidDomainNames(domains types.Domains) bool {
	for _, domain := range domains {
		if !isValidDomainName(string(domain)) {
			return false
		}
	}
	return true
}

func ParseConfig(logger *slog.Logger, cfgPath string) (VhostConfigs, error) {
	vhostConfigs := VhostConfigs{}
	p, err := parser.NewParser(cfgPath, parser.WithIncludeParsing())
	if err != nil {
		return vhostConfigs, err
	}

	conf, errParse := p.Parse()
	if errParse != nil {
		return vhostConfigs, fmt.Errorf("failed to parse config (%s): %w", cfgPath, errParse)
	}

	for _, serverNameDirective := range conf.FindDirectives("server_name") {
		sslCertDirectives := serverNameDirective.GetParent().GetBlock().FindDirectives("ssl_certificate")
		sslCertKeyDirectives := serverNameDirective.GetParent().GetBlock().FindDirectives("ssl_certificate_key")

		if len(sslCertDirectives) > 0 && len(sslCertKeyDirectives) > 0 {
			vhostConfig := VhostConfig{}
			domains := types.Domains{}
			for _, domain := range serverNameDirective.GetParameters() {
				domains = append(domains, types.Domain(domain.String()))
			}

			if !isValidDomainNames(domains) {
				logger.Debug(fmt.Sprintf("Skipping domain '%s' due to invalid domain", domains.ToStringSlice()))
				continue
			}

			vhostConfig.ServerName = domains
			for _, sslCertDirective := range sslCertDirectives {
				if len(sslCertDirective.GetParameters()) == 1 {
					vhostConfig.CertPath = sslCertDirective.GetParameters()[0].String()
				}
			}
			for _, sslCertKeyDirective := range sslCertKeyDirectives {
				if len(sslCertKeyDirective.GetParameters()) == 1 {
					vhostConfig.KeyPath = sslCertKeyDirective.GetParameters()[0].String()
				}
			}

			vhostConfigs = append(vhostConfigs, vhostConfig)
		}
	}
	return vhostConfigs, nil
}
