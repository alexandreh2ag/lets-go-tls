package requester

import (
	"github.com/alexandreh2ag/lets-go-tls/config"
	"github.com/alexandreh2ag/lets-go-tls/context"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_createRequester_Success(t *testing.T) {
	ctx := context.TestContext(nil)
	want := &static{id: "foo"}
	want.domainRequests = []*types.DomainRequest{{Domains: types.Domains{"foo.com"}, Requester: want}}
	cfg := config.RequesterConfig{
		Id:     "foo",
		Type:   "static",
		Config: map[string]interface{}{"domains": [][]string{{"foo.com"}}},
	}
	got, err := createRequester(ctx, cfg)
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func Test_createRequester_Fail(t *testing.T) {
	ctx := context.TestContext(nil)
	cfg := config.RequesterConfig{
		Id:     "foo",
		Type:   "wrong",
		Config: map[string]interface{}{},
	}
	got, err := createRequester(ctx, cfg)
	assert.Nil(t, got)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config requester type 'wrong' does not exist")
}

func TestCreateRequesters_Success(t *testing.T) {
	ctx := context.TestContext(nil)
	requesters := []config.RequesterConfig{
		{
			Id:     "foo",
			Type:   "static",
			Config: map[string]interface{}{"domains": [][]string{{"foo.com"}}},
		},
	}
	staticP := &static{id: "foo"}
	staticP.domainRequests = []*types.DomainRequest{{Domains: types.Domains{"foo.com"}, Requester: staticP}}
	want := types.Requesters{"foo": staticP}
	got, err := CreateRequesters(ctx, requesters)
	assert.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestCreateRequesters_Fail(t *testing.T) {
	ctx := context.TestContext(nil)
	requesters := []config.RequesterConfig{
		{
			Id:     "foo",
			Type:   "wrong",
			Config: map[string]interface{}{},
		},
	}

	got, err := CreateRequesters(ctx, requesters)
	assert.Nil(t, got)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "config requester type 'wrong' does not exist")
}
