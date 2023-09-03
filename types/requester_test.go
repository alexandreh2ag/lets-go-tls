package types

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

var _ Requester = &dummyProvider{}

type dummyProvider struct {
}

func (d dummyProvider) ID() string {
	panic("implement me")
}

func (d dummyProvider) Fetch() ([]*DomainRequest, error) {
	panic("implement me")
}

func TestProviders_Get_SuccessFound(t *testing.T) {
	provider := &dummyProvider{}
	providers := Requesters{"foo": provider}
	got := providers.Get("foo")
	assert.Equal(t, provider, got)
}

func TestProviders_Get_SuccessNotFound(t *testing.T) {
	providers := Requesters{}
	got := providers.Get("foo")
	assert.Nil(t, got)
}
