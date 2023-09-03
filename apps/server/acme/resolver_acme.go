package acme

import (
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/alexandreh2ag/lets-go-tls/types/acme"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/registration"
	"strings"
)

var _ types.Resolver = &ResolverAcme{}

type ResolverAcme struct {
	Id        string
	Filters   []string
	Client    *lego.Client
	Challenge acme.Challenge
}

func (r ResolverAcme) ID() string {
	return r.Id
}

func (r ResolverAcme) TypeChallenge() string {
	return r.Challenge.Type()
}

func (r ResolverAcme) Register(options registration.RegisterOptions) (*registration.Resource, error) {
	return r.Client.Registration.Register(options)
}

func (r ResolverAcme) Obtain(request certificate.ObtainRequest) (*certificate.Resource, error) {
	return r.Client.Certificate.Obtain(request)
}

func (r ResolverAcme) RenewWithOptions(certRes certificate.Resource, options *certificate.RenewOptions) (*certificate.Resource, error) {
	return r.Client.Certificate.RenewWithOptions(certRes, options)
}

func (r ResolverAcme) Match(certificate *types.Certificate) bool {
	if len(r.Filters) > 0 && len(certificate.Domains) > 0 {
		for _, domain := range certificate.Domains {
			match := false
			for _, domainFilter := range r.Filters {
				if strings.Contains(string(domain), domainFilter) {
					match = true
					continue
				}
			}
			if !match {
				return false
			}
		}
		return true
	}
	return false
}
