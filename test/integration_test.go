package test

import (
	"fmt"
	agentCli "github.com/alexandreh2ag/lets-go-tls/apps/agent/cli"
	agentConfig "github.com/alexandreh2ag/lets-go-tls/apps/agent/config"
	agentCtx "github.com/alexandreh2ag/lets-go-tls/apps/agent/context"
	srvCli "github.com/alexandreh2ag/lets-go-tls/apps/server/cli"
	srvConfig "github.com/alexandreh2ag/lets-go-tls/apps/server/config"
	srvCtx "github.com/alexandreh2ag/lets-go-tls/apps/server/context"
	"github.com/alexandreh2ag/lets-go-tls/config"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-jose/go-jose/v4/json"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
	"gopkg.in/yaml.v2"
	"io"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestIntegration_Full(t *testing.T) {
	fs := afero.NewMemMapFs()
	viper.Reset()
	viper.SetFs(fs)

	srvAddress := "0.0.0.0:2123"
	srvBasePath := "/app/server"
	srvConfigPath := filepath.Join(srvBasePath, "config.yaml")
	srvStatePath := filepath.Join(srvBasePath, "state.json")

	agentAddress := "0.0.0.0:2124"
	agentBasePath := "/app/agent"
	agentConfigPath := filepath.Join(agentBasePath, "config.yaml")
	agentStatePath := filepath.Join(agentBasePath, "state.json")
	agentStoragePath := filepath.Join(agentBasePath, "storage")

	_, apiURL := tester.SetupFakeAPI(t)

	ctxSrv := srvCtx.TestContext(nil)
	ctxSrv.Fs = fs
	cfgSrv := &srvConfig.Config{}
	cfgSrv.HTTP.Listen = srvAddress
	cfgSrv.Cache.Type = "memory"
	cfgSrv.JWT.Key = "supersecret"
	cfgSrv.JWT.Method = "HS256"
	cfgSrv.Acme.Email = "dev@example.com"
	cfgSrv.Acme.CAServer = apiURL + "/dir"
	cfgSrv.Acme.RenewPeriod = ctxSrv.Config.Acme.RenewPeriod
	cfgSrv.Acme.MaxAttempt = 3
	cfgSrv.Acme.DelayFailed = time.Hour
	cfgSrv.Interval = 1 * time.Second
	cfgSrv.LockDuration = 5 * time.Second
	cfgSrv.UnusedRetentionDuration = 5 * time.Minute
	cfgSrv.State.Type = "fs"
	cfgSrv.State.Config = map[string]interface{}{"path": srvStatePath}
	cfgSrv.Requesters = []config.RequesterConfig{
		{Id: "agent", Type: "agent", Config: map[string]interface{}{"addresses": []string{fmt.Sprintf("http://%s", agentAddress)}}},
	}

	cfgSrvOut := map[string]interface{}{}
	_ = mapstructure.Decode(cfgSrv, &cfgSrvOut)
	cfgSrvRaw, errYamlCfgSrv := yaml.Marshal(cfgSrvOut)
	assert.NoError(t, errYamlCfgSrv)
	errWriteCfgSrv := afero.WriteFile(ctxSrv.Fs, srvConfigPath, cfgSrvRaw, 0644)
	assert.NoError(t, errWriteCfgSrv)

	stateSrv := &types.State{
		Certificates: types.Certificates{
			&types.Certificate{
				Identifier:     "example.com-0",
				Main:           "example.com",
				Domains:        types.Domains{"example.com"},
				ExpirationDate: time.Now().Add(time.Hour * 744), // 1 month
				Key:            []byte("key"),
				Certificate:    []byte("certificate"),
			},
		},
	}
	_ = ctxSrv.Fs.MkdirAll(srvBasePath, 0775)

	stateSrvRaw, errJsonStateSrv := json.Marshal(stateSrv)
	assert.NoError(t, errJsonStateSrv)
	errWriteStateSrv := afero.WriteFile(ctxSrv.Fs, srvStatePath, stateSrvRaw, 0644)
	assert.NoError(t, errWriteStateSrv)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"exp": time.Now().Add(time.Hour * 24).Unix(),
		})

	tokenString, errJwtTokenSign := token.SignedString([]byte(cfgSrv.JWT.Key))
	assert.NoError(t, errJwtTokenSign)

	ctxAgent := agentCtx.TestContext(nil)
	ctxAgent.Fs = fs
	ctxAgent.HttpClient = &fasthttp.Client{}
	cfgAgent := &agentConfig.Config{}
	cfgAgent.HTTP.Listen = agentAddress
	cfgAgent.Interval = 1 * time.Second
	cfgAgent.State.Type = "fs"
	cfgAgent.State.Config = map[string]interface{}{"path": agentStatePath}
	cfgAgent.Manager.Address = fmt.Sprintf("http://%s", srvAddress)
	cfgAgent.Manager.TokenJWT = tokenString
	cfgAgent.Requesters = []config.RequesterConfig{
		{Id: "file_req", Type: "static", Config: map[string]interface{}{"domains": [][]string{{"example.com"}}}},
	}
	cfgAgent.Storages = []agentConfig.StorageConfig{
		{Id: "file_storage", Type: "fs", Config: map[string]interface{}{"path": agentStoragePath}},
	}

	_ = ctxAgent.Fs.MkdirAll(agentBasePath, 0775)
	_ = ctxAgent.Fs.MkdirAll(agentStoragePath, 0775)

	cfgAgentOut := map[string]interface{}{}
	_ = mapstructure.Decode(cfgAgent, &cfgAgentOut)
	cfgAgentRaw, errYamlCfgAgent := yaml.Marshal(cfgAgentOut)
	assert.NoError(t, errYamlCfgAgent)
	errWriteCfgAgent := afero.WriteFile(ctxAgent.Fs, agentConfigPath, cfgAgentRaw, 0644)
	assert.NoError(t, errWriteCfgAgent)

	go func() {
		cmdSrvRoot := srvCli.GetRootCmd(ctxSrv)
		cmdSrvRoot.SetOut(io.Discard)
		cmdSrvRoot.SetErr(io.Discard)

		cmdSrvRoot.SetArgs([]string{
			"start",
			"--config", srvConfigPath,
		})
		errExecCmdSrvRoot := cmdSrvRoot.Execute()
		assert.NoError(t, errExecCmdSrvRoot)
	}()
	time.Sleep(500 * time.Millisecond)

	go func() {
		cmdAgentRoot := agentCli.GetRootCmd(ctxAgent)
		cmdAgentRoot.SetOut(io.Discard)
		cmdAgentRoot.SetErr(io.Discard)

		cmdAgentRoot.SetArgs([]string{
			"start",
			"--config", agentConfigPath,
		})
		errExecCmdAgentRoot := cmdAgentRoot.Execute()
		assert.NoError(t, errExecCmdAgentRoot)
	}()
	time.Sleep(2 * time.Second)

	certContent, certErr := afero.ReadFile(fs, filepath.Join(agentStoragePath, "example.com-0.crt"))
	assert.NoError(t, certErr)
	assert.Equal(t, "certificate", string(certContent))
	keyContent, keyErr := afero.ReadFile(fs, filepath.Join(agentStoragePath, "example.com-0.key"))
	assert.NoError(t, keyErr)
	assert.Equal(t, "key", string(keyContent))

	ctxSrv.Signal() <- syscall.SIGINT
	ctxAgent.Signal() <- syscall.SIGINT

}
