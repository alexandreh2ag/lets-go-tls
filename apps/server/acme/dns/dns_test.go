package dns

import (
	"github.com/alexandreh2ag/lets-go-tls/apps/server/config"
	"github.com/alexandreh2ag/lets-go-tls/apps/server/context"
	mockTypesAcme "github.com/alexandreh2ag/lets-go-tls/mocks/types/acme"
	"github.com/alexandreh2ag/lets-go-tls/types/acme"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestCreateDnsChallenge_Success(t *testing.T) {
	ctx := context.TestContext(nil)
	ctrl := gomock.NewController(t)
	key := "dummy"
	challenge := mockTypesAcme.NewMockChallenge(ctrl)
	TypeDnsProviderMapping[key] = func(ctx *context.ServerContext, id string, config map[string]interface{}) (acme.Challenge, error) {
		return challenge, nil
	}
	cfg := config.ResolverConfig{
		Type:   key,
		Config: map[string]interface{}{},
	}
	got, err := CreateDnsChallenge(ctx, "foo", cfg)
	assert.NoError(t, err)
	assert.NotNil(t, got)
}

func TestCreateDnsChallenge_Fail(t *testing.T) {
	ctx := context.TestContext(nil)
	cfg := config.ResolverConfig{
		Type:   "wrong",
		Config: map[string]interface{}{},
	}
	got, err := CreateDnsChallenge(ctx, "foo", cfg)
	assert.Nil(t, got)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config dns challenge id 'foo' (type wrong) does not exist")
}
