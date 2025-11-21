package acme

import (
	"testing"

	"github.com/alexandreh2ag/lets-go-tls/apps/server/acme/dns"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/config"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/context"
	mockTypesAcme "github.com/alexandreh2ag/lets-go-tls/mocks/types/acme"
	"github.com/alexandreh2ag/lets-go-tls/types/acme"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCreateResolvers_Success(t *testing.T) {
	ctx := context.TestContext(nil)
	_, apiURL := tester.SetupFakeAPI(t)
	ctx.Config.Acme.CAServer = apiURL + "/dir"
	account, _ := acme.NewAccount("dev@example.com")
	got, err := CreateResolvers(ctx, account)
	assert.NoError(t, err)
	assert.Len(t, got, 1)
}

func TestCreateResolvers_Fail(t *testing.T) {
	ctx := context.TestContext(nil)
	_, apiURL := tester.SetupFakeAPI(t)
	ctx.Config.Acme.CAServer = apiURL + "/dir"
	ctx.Config.Acme.Resolvers = map[string]config.ResolverConfig{"test": {Type: "wrong"}}
	account, _ := acme.NewAccount("dev@example.com")
	got, err := CreateResolvers(ctx, account)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config dns challenge id 'test' (type wrong) does not exist")
	assert.Len(t, got, 0)
}

func Test_createResolver_Success(t *testing.T) {
	ctx := context.TestContext(nil)
	_, apiURL := tester.SetupFakeAPI(t)

	account, _ := acme.NewAccount("dev@example.com")
	configAcme := lego.NewConfig(account)
	configAcme.CADirURL = apiURL + "/dir"
	cfg := config.ResolverConfig{
		Type:    acme.TypeHTTP01,
		Filters: []string{"*"},
	}
	got, err := createResolver(ctx, "foo", cfg, configAcme)
	assert.NoError(t, err)
	assert.NotNil(t, got)
}

func Test_createResolver_SuccessWithDns(t *testing.T) {
	ctx := context.TestContext(nil)
	_, apiURL := tester.SetupFakeAPI(t)
	ctrl := gomock.NewController(t)
	key := "dummy"
	challenge := mockTypesAcme.NewMockChallenge(ctrl)
	dns.TypeDnsProviderMapping[key] = func(ctx *context.ServerContext, id string, config map[string]interface{}) (acme.Challenge, error) {
		return challenge, nil
	}
	account, _ := acme.NewAccount("dev@example.com")
	configAcme := lego.NewConfig(account)
	configAcme.CADirURL = apiURL + "/dir"
	cfg := config.ResolverConfig{
		Type:    key,
		Filters: []string{"*"},
	}
	got, err := createResolver(ctx, "foo", cfg, configAcme)
	assert.NoError(t, err)
	assert.NotNil(t, got)
}

func Test_createResolver_FailCreateLegoClient(t *testing.T) {
	ctx := context.TestContext(nil)

	account, _ := acme.NewAccount("dev@example.com")
	configAcme := lego.NewConfig(account)
	configAcme.CADirURL = ""
	cfg := config.ResolverConfig{
		Type:    acme.TypeHTTP01,
		Filters: []string{"*"},
	}
	got, err := createResolver(ctx, "foo", cfg, configAcme)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to init acme client for resolver")
	assert.Nil(t, got)
}

func Test_createResolver_FailCreateDnsProvider(t *testing.T) {
	ctx := context.TestContext(nil)
	_, apiURL := tester.SetupFakeAPI(t)

	account, _ := acme.NewAccount("dev@example.com")
	configAcme := lego.NewConfig(account)
	configAcme.CADirURL = apiURL + "/dir"
	cfg := config.ResolverConfig{
		Type:    "wrong",
		Filters: []string{"*"},
	}
	got, err := createResolver(ctx, "foo", cfg, configAcme)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config dns challenge id 'foo' (type wrong) does not exist")
	assert.Nil(t, got)
}
