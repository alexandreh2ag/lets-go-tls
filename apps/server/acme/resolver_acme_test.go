package acme

import (
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/alexandreh2ag/lets-go-tls/types/acme"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	legoLog "github.com/go-acme/lego/v4/log"
	"github.com/go-acme/lego/v4/platform/tester"
	"github.com/go-acme/lego/v4/registration"
	"github.com/stretchr/testify/assert"
	"io"
	"log"
	"testing"
)

type dummyChallenge struct {
	id string
}

func (d dummyChallenge) Present(domain, token, keyAuth string) error {
	return nil
}

func (d dummyChallenge) CleanUp(domain, token, keyAuth string) error {
	return nil
}

func (d dummyChallenge) ID() string {
	return d.id
}

func (d dummyChallenge) Type() string {
	return "dummy"
}

func TestResolverAcme_ID(t *testing.T) {
	id := "foo"
	r := &ResolverAcme{Id: id}
	assert.Equal(t, id, r.ID())
}

func TestResolverAcme_TypeChallenge(t *testing.T) {
	r := &ResolverAcme{Challenge: &dummyChallenge{}}
	assert.Equal(t, "dummy", r.TypeChallenge())
}

func TestResolverAcme_Register(t *testing.T) {
	_, apiURL := tester.SetupFakeAPI(t)
	legoLog.Logger = log.New(io.Discard, "", log.LstdFlags)
	account, err := acme.NewAccount("dev@example.com")
	assert.NoError(t, err)
	cfgAcme := lego.NewConfig(account)
	cfgAcme.CADirURL = apiURL + "/dir"
	client, err := lego.NewClient(cfgAcme)
	assert.NoError(t, err)
	r := &ResolverAcme{Client: client}
	_, err = r.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	assert.Error(t, err)
}

func TestResolverAcme_Obtain(t *testing.T) {
	_, apiURL := tester.SetupFakeAPI(t)
	legoLog.Logger = log.New(io.Discard, "", log.LstdFlags)
	account, err := acme.NewAccount("dev@example.com")
	assert.NoError(t, err)
	cfgAcme := lego.NewConfig(account)
	cfgAcme.CADirURL = apiURL + "/dir"
	client, err := lego.NewClient(cfgAcme)
	assert.NoError(t, err)
	r := &ResolverAcme{Client: client}
	request := certificate.ObtainRequest{
		Domains: []string{},
		Bundle:  false,
	}
	_, err = r.Obtain(request)
	assert.Error(t, err)
}

func TestResolverAcme_RenewWithOptions(t *testing.T) {
	_, apiURL := tester.SetupFakeAPI(t)
	legoLog.Logger = log.New(io.Discard, "", log.LstdFlags)
	account, err := acme.NewAccount("dev@example.com")
	assert.NoError(t, err)
	cfgAcme := lego.NewConfig(account)
	cfgAcme.CADirURL = apiURL + "/dir"
	client, err := lego.NewClient(cfgAcme)
	assert.NoError(t, err)
	r := &ResolverAcme{Client: client}
	request := certificate.Resource{}
	options := &certificate.RenewOptions{}
	_, err = r.RenewWithOptions(request, options)
	assert.Error(t, err)
}

func TestResolverAcme_Match(t *testing.T) {

	tests := []struct {
		name        string
		filters     []string
		certificate *types.Certificate
		want        bool
	}{
		{
			name:        "EmptyFilters",
			certificate: &types.Certificate{Domains: types.Domains{"example.com"}},
			want:        false,
		},
		{
			name:        "EmptyCertificateDomains",
			filters:     []string{"example.com"},
			certificate: &types.Certificate{Domains: types.Domains{}},
			want:        false,
		},
		{
			name:        "NoMatch",
			filters:     []string{"example.com"},
			certificate: &types.Certificate{Domains: types.Domains{"example2.com"}},
			want:        false,
		},
		{
			name:        "NoMatchWithOneMatching",
			filters:     []string{"example.com"},
			certificate: &types.Certificate{Domains: types.Domains{"example.com", "example2.com"}},
			want:        false,
		},
		{
			name:        "OneMatch",
			filters:     []string{"example.com"},
			certificate: &types.Certificate{Domains: types.Domains{"example.com"}},
			want:        true,
		},
		{
			name:        "TwoMatch",
			filters:     []string{"example.com"},
			certificate: &types.Certificate{Domains: types.Domains{"example.com", "sub.example.com"}},
			want:        true,
		},
		{
			name:        "WildcardMatch",
			filters:     []string{"example.com"},
			certificate: &types.Certificate{Domains: types.Domains{"example.com", "*.example.com"}},
			want:        true,
		},
		{
			name:        "DifferentDomainsCertificateMatch",
			filters:     []string{"example.com", "example.dev"},
			certificate: &types.Certificate{Domains: types.Domains{"example.com", "*.example.dev"}},
			want:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := ResolverAcme{
				Filters: tt.filters,
			}
			assert.Equalf(t, tt.want, r.Match(tt.certificate), "Match(%v)", tt.certificate)
		})
	}
}
