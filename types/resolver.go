package types

import (
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/registration"
)

const (
	DefaultKey = "default"
)

type Resolvers map[string]Resolver

func (r Resolvers) FindResolver(certificate *Certificate) Resolver {
	defaultResolver := r[DefaultKey]

	for _, resolver := range r {

		if resolver.Match(certificate) {
			return resolver
		}
	}

	return defaultResolver
}

type Resolver interface {
	ID() string
	TypeChallenge() string
	Obtain(request certificate.ObtainRequest) (*certificate.Resource, error)
	RenewWithOptions(certRes certificate.Resource, options *certificate.RenewOptions) (*certificate.Resource, error)
	Register(options registration.RegisterOptions) (*registration.Resource, error)
	Match(certificate *Certificate) bool
}
