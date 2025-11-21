package types

import (
	"testing"

	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/registration"
	"github.com/stretchr/testify/assert"
)

var _ Resolver = &dummyResolver{}

type dummyResolver struct {
	id    string
	match bool
}

func (d dummyResolver) ID() string {
	return d.id
}

func (d dummyResolver) TypeChallenge() string {
	panic("implement me")
}

func (d dummyResolver) Obtain(request certificate.ObtainRequest) (*certificate.Resource, error) {
	panic("implement me")
}

func (d dummyResolver) RenewWithOptions(certRes certificate.Resource, options *certificate.RenewOptions) (*certificate.Resource, error) {
	panic("implement me")
}

func (d dummyResolver) Register(options registration.RegisterOptions) (*registration.Resource, error) {
	panic("implement me")
}

func (d dummyResolver) Match(certificate *Certificate) bool {
	return d.match
}

func TestResolvers_FindResolver_Success(t *testing.T) {
	defaultResolver := &dummyResolver{id: DefaultKey}
	resolver := &dummyResolver{id: "foo", match: true}
	resolvers := Resolvers{"foo": resolver, DefaultKey: defaultResolver}
	cert := &Certificate{Domains: Domains{"example.com"}}
	got := resolvers.FindResolver(cert)
	assert.Equal(t, resolver, got)
}

func TestResolvers_FindResolver_SuccessDefault(t *testing.T) {
	defaultResolver := &dummyResolver{id: DefaultKey}
	resolver := &dummyResolver{id: "foo", match: false}
	resolvers := Resolvers{"foo": resolver, DefaultKey: defaultResolver}
	cert := &Certificate{Domains: Domains{"example.com"}}
	got := resolvers.FindResolver(cert)
	assert.Equal(t, defaultResolver, got)
}
