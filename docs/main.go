package main

import (
	"fmt"
	agentConfig "github.com/alexandreh2ag/lets-go-tls/apps/agent/config"
	agentRequester "github.com/alexandreh2ag/lets-go-tls/apps/agent/requester"
	"github.com/alexandreh2ag/lets-go-tls/apps/agent/storage/certificate"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/acme/dns/gandiv5"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/acme/dns/httpreq"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/acme/dns/ovh"
	serverConfig "github.com/alexandreh2ag/lets-go-tls/apps/server/config"
	serverRequester "github.com/alexandreh2ag/lets-go-tls/apps/server/requester"
	"github.com/alexandreh2ag/lets-go-tls/config"
	"github.com/alexandreh2ag/lets-go-tls/hook"
	"github.com/alexandreh2ag/lets-go-tls/requester"
	"github.com/alexandreh2ag/lets-go-tls/storage/state"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/mitchellh/mapstructure"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
	"time"
)

const (
	cfgPath = "examples"
)

func main() {

	serverCfg := getServerConfig()

	agentCfg := getAgentConfig()

	// generate files
	serverCfgRaw, err := yaml.Marshal(decodeToMap(serverCfg))
	if err != nil {
		panic(fmt.Errorf("faild marshal server config: %v", err))
	}

	agentCfgRaw, err := yaml.Marshal(decodeToMap(agentCfg))
	if err != nil {
		panic(fmt.Errorf("faild marshal agent config: %v", err))
	}

	err = os.WriteFile(filepath.Join(cfgPath, "server.cfg.yml"), serverCfgRaw, 0660)
	if err != nil {
		panic(fmt.Errorf("faild write server config: %v", err))
	}

	err = os.WriteFile(filepath.Join(cfgPath, "agent.cfg.yml"), agentCfgRaw, 0660)
	if err != nil {
		panic(fmt.Errorf("faild write agent config: %v", err))
	}
}

func getServerConfig() serverConfig.Config {
	serverCfg := serverConfig.DefaultConfig()

	serverCfg.Acme.Email = "acme@example.com"
	serverCfg.JWT.Key = "superSecret"

	serverCfg.State.Type = state.FsKey

	serverCfg.State.Config = decodeToMap(state.ConfigFs{Path: "/var/lib/lets-go-tls/state.json"})

	serverCfg.Requesters = []config.RequesterConfig{
		{
			Id:   "static",
			Type: requester.StaticKey,
			Config: decodeToMap(requester.ConfigStatic{ListDomains: [][]string{
				{"example.com"},
				{"foo.com", "bar.com"},
			}}),
		},
		{
			Id:   "agents",
			Type: serverRequester.AgentKey,
			Config: decodeToMap(serverRequester.ConfigAgent{
				Addresses: []string{"127.0.0.1:8080"},
			}),
		},
	}

	serverCfg.Acme.Resolvers = map[string]serverConfig.ResolverConfig{
		ovh.KeyDnsOVH: {
			Type: ovh.KeyDnsOVH,
			Config: decodeToMap(ovh.ConfigOvh{
				AccessToken:        "",
				ApplicationKey:     "",
				ApplicationSecret:  "",
				ClientID:           "",
				ClientSecret:       "",
				ConsumerKey:        "",
				Endpoint:           "ovh-eu",
				PropagationTimeout: time.Second * 60,
				PollingInterval:    2 * time.Second,
				HttpTimeout:        180 * time.Second,
			}),
			Filters: []string{"foo.com"},
		},
		httpreq.KeyDnsHttpReq: {
			Type: httpreq.KeyDnsHttpReq,
			Config: decodeToMap(httpreq.ConfigHttpReq{
				Endpoint:           "",
				Mode:               "",
				Username:           "",
				Password:           "",
				PropagationTimeout: time.Second * 60,
				PollingInterval:    2 * time.Second,
				HttpTimeout:        180 * time.Second,
			}),
			Filters: []string{"foo.com"},
		},
		gandiv5.KeyDnsGandiV5: {
			Type: gandiv5.KeyDnsGandiV5,
			Config: decodeToMap(gandiv5.ConfigGandiV5{
				APIKey:             "",
				PropagationTimeout: time.Second * 60,
				PollingInterval:    2 * time.Second,
				HttpTimeout:        180 * time.Second,
			}),
			Filters: []string{"foo.com"},
		},
	}

	return serverCfg
}

func getAgentConfig() agentConfig.Config {
	agentCfg := agentConfig.DefaultConfig()
	agentCfg.State.Type = state.FsKey

	agentCfg.Manager.Address = "127.0.0.1:8080"
	agentCfg.Manager.TokenJWT = "tokenJwt"

	agentCfg.State.Config = decodeToMap(state.ConfigFs{Path: "/var/lib/lets-go-tls/state.json"})

	agentCfg.Requesters = []config.RequesterConfig{
		{
			Id:   "static",
			Type: requester.StaticKey,
			Config: decodeToMap(requester.ConfigStatic{ListDomains: [][]string{
				{"example.com"},
				{"foo.com", "bar.com"},
			}}),
		},
		{
			Id:   "traefikV2",
			Type: agentRequester.TraefikV2Key,
			Config: decodeToMap(agentRequester.ConfigTraefik{
				Addresses: []string{"http://127.0.0.1"},
			}),
		},
		{
			Id:   "traefikV3",
			Type: agentRequester.TraefikV3Key,
			Config: decodeToMap(agentRequester.ConfigTraefik{
				Addresses: []string{"http://127.0.0.1"},
			}),
		},
	}

	agentCfg.Storages = []agentConfig.StorageConfig{
		{
			Id:   "fs",
			Type: certificate.FsKey,
			Config: decodeToMap(certificate.ConfigFs{
				Path:           "/var/lib/lets-go-tls/ssl",
				PrefixFilename: "",
				Owner:          "root",
				Group:          "root",
				SpecificDomains: []certificate.ConfigSpecificDomain{
					{Identifier: "custom", Domains: types.Domains{"example.com"}},
					{Identifier: "custom", Path: "foo", Domains: types.Domains{"example.com"}},
				},
				PostHook: &hook.Hook{
					Cmd:     "echo 1",
					Timeout: time.Second * 60,
				},
			}),
		},
		{
			Id:   "traefikV2",
			Type: certificate.TraefikV2Key,
			Config: decodeToMap(certificate.ConfigTraefik{
				Path:           "/etc/traefik/config",
				PrefixFilename: "",
				Owner:          "root",
				Group:          "root",
			}),
		},
		{
			Id:   "traefikV3",
			Type: certificate.TraefikV3Key,
			Config: decodeToMap(certificate.ConfigTraefik{
				Path:           "/etc/traefik/config",
				PrefixFilename: "",
				Owner:          "root",
				Group:          "root",
			}),
		},
		{
			Id:   "haproxy",
			Type: certificate.HaproxyKey,
			Config: decodeToMap(certificate.ConfigHaproxy{
				ConfigFs: certificate.ConfigFs{
					Path:           "/etc/haproxy/ssl",
					PrefixFilename: "",
					Owner:          "root",
					Group:          "root",
					AddPem:         true,
				},
				CrtListPath: "/etc/haproxy/crt-list.txt",
			}),
		},
	}

	return agentCfg
}

func decodeToMap(input interface{}) map[string]interface{} {
	output := map[string]interface{}{}
	_ = mapstructure.Decode(input, &output)
	return output
}
