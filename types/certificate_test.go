package types

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCertificate_Match(t *testing.T) {

	tests := []struct {
		name               string
		certificateDomains Domains
		domains            Domains
		want               bool
	}{
		{
			name:               "EmptyDomains",
			certificateDomains: Domains{},
			domains:            Domains{},
			want:               false,
		},
		{
			name:               "EmptyCertificateDomains",
			certificateDomains: Domains{},
			domains:            Domains{"example.com"},
			want:               false,
		},
		{
			name:               "EmptyArgDomains",
			certificateDomains: Domains{"example.com"},
			domains:            Domains{},
			want:               false,
		},
		{
			name:               "NoMatch",
			certificateDomains: Domains{"example.com"},
			domains:            Domains{"example.co.uk"},
			want:               false,
		},
		{
			name:               "NoMatchWildcard",
			certificateDomains: Domains{"*.example.com"},
			domains:            Domains{"foo.example.co.uk"},
			want:               false,
		},
		{
			name:               "NoMatchWildcardSubSubdomain",
			certificateDomains: Domains{"*.example.com"},
			domains:            Domains{"foo.bar.example.co.uk"},
			want:               false,
		},
		{
			name:               "Match",
			certificateDomains: Domains{"example.com"},
			domains:            Domains{"example.com"},
			want:               true,
		},
		{
			name:               "MatchWildcard",
			certificateDomains: Domains{"*.example.com"},
			domains:            Domains{"foo.example.com"},
			want:               true,
		},
		{
			name:               "MatchMultiCertificateDomains",
			certificateDomains: Domains{"example.com", "example.us", "example.co.uk"},
			domains:            Domains{"example.com"},
			want:               true,
		},
		{
			name:               "MatchMultiArgsDomains",
			certificateDomains: Domains{"example.com", "example.us", "example.co.uk"},
			domains:            Domains{"example.us", "example.com"},
			want:               true,
		},
		{
			name:               "MatchMixed",
			certificateDomains: Domains{"example.com", "*.example.us", "example.co.uk"},
			domains:            Domains{"foo.example.us", "example.com"},
			want:               true,
		},
		{
			name:               "NoMatchMixed",
			certificateDomains: Domains{"example.com", "*.example.us", "example.co.uk"},
			domains:            Domains{"foo.example.us", "example.com", "foo.foo.example.us"},
			want:               false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Certificate{
				Domains: tt.certificateDomains,
			}
			assert.Equalf(t, tt.want, c.Match(tt.domains), "Match(%v)", tt.domains)
		})
	}
}

func TestCertificate_GetCertificateFilename(t *testing.T) {
	c := Certificate{Identifier: "foo"}
	assert.Equal(t, "foo.crt", c.GetCertificateFilename())
}

func TestCertificate_GetKeyFilename(t *testing.T) {
	c := Certificate{Identifier: "foo"}
	assert.Equal(t, "foo.key", c.GetKeyFilename())
}

func TestCertificates_CheckIdentifierUnique(t *testing.T) {
	tests := []struct {
		name       string
		c          Certificates
		identifier string
		want       bool
	}{
		{
			name:       "UniqueEmptyCertificates",
			c:          Certificates{},
			identifier: "foo",
			want:       true,
		},
		{
			name:       "NoUniqueEmptyCertificatesAndIdentifier",
			c:          Certificates{},
			identifier: "",
			want:       false,
		},
		{
			name:       "NoUniqueEmptyIdentifier",
			c:          Certificates{{Identifier: "foo"}},
			identifier: "",
			want:       false,
		},
		{
			name:       "NoUniqueEmptyIdentifier",
			c:          Certificates{{Identifier: "foo"}, {Identifier: "bar"}},
			identifier: "bar",
			want:       false,
		},
		{
			name:       "UniqueEmptyIdentifier",
			c:          Certificates{{Identifier: "foo"}, {Identifier: "bar"}},
			identifier: "bar2",
			want:       true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.c.CheckIdentifierUnique(tt.identifier), "CheckIdentifierUnique(%v)", tt.identifier)
		})
	}
}

func TestCertificates_Match(t *testing.T) {
	cert := &Certificate{Domains: Domains{Domain("example.com")}, Certificate: []byte("certificate"), Key: []byte("key")}
	certNotValid := &Certificate{Domains: Domains{Domain("example-not-valid.com")}}
	tests := []struct {
		name      string
		c         Certificates
		request   *DomainRequest
		onlyValid bool
		want      *Certificate
	}{
		{
			name:      "Match",
			c:         Certificates{cert, certNotValid},
			request:   &DomainRequest{Domains: Domains{Domain("example.com")}},
			onlyValid: false,
			want:      cert,
		},
		{
			name:      "NotMatch",
			c:         Certificates{cert, certNotValid},
			request:   &DomainRequest{Domains: Domains{Domain("example2.com")}},
			onlyValid: false,
			want:      nil,
		},
		{
			name:      "NotMatchWithNotValid",
			c:         Certificates{cert, certNotValid},
			request:   &DomainRequest{Domains: Domains{Domain("example-not-valid.com")}},
			onlyValid: true,
			want:      nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.c.Match(tt.request, tt.onlyValid), "Match(%v)", tt.request)
		})
	}
}

func TestCertificates_GetCertificate(t *testing.T) {
	cert1 := &Certificate{Identifier: "foo"}
	cert2 := &Certificate{Identifier: "bar"}
	tests := []struct {
		name       string
		c          Certificates
		identifier string
		want       *Certificate
	}{
		{
			name:       "SuccessGetCertificate",
			c:          Certificates{cert1, cert2},
			identifier: cert2.Identifier,
			want:       cert2,
		},
		{
			name:       "SuccessNotFound",
			c:          Certificates{cert1, cert2},
			identifier: "wrong",
			want:       nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.c.GetCertificate(tt.identifier), "GetCertificate(%v)", tt.identifier)
		})
	}
}

func TestCertificate_IsValid(t *testing.T) {

	tests := []struct {
		name        string
		certificate Certificate
		want        bool
	}{
		{
			name:        "EmptyCertificate",
			certificate: Certificate{},
			want:        false,
		},
		{
			name: "CertificateValid",
			certificate: Certificate{
				Identifier:  "example.com-0",
				Main:        "example.com",
				Domains:     Domains{"example.com"},
				Key:         []byte("key"),
				Certificate: []byte("certificate"),
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			assert.Equalf(t, tt.want, tt.certificate.IsValid(), "IsValid()")
		})
	}
}

func TestCertificates_UnusedCertificates(t *testing.T) {
	cert1 := &Certificate{Domains: Domains{Domain("example.com")}}
	cert2 := &Certificate{Domains: Domains{Domain("example2.com")}}
	certificates := Certificates{cert1, cert2}

	tests := []struct {
		name            string
		domainsRequests []*DomainRequest
		want            Certificates
	}{
		{
			name: "SuccessAllUsed",
			domainsRequests: []*DomainRequest{
				{Domains: Domains{Domain("example.com")}},
				{Domains: Domains{Domain("example2.com")}},
			},
			want: Certificates{},
		},
		{
			name: "SuccessOneNotUsed",
			domainsRequests: []*DomainRequest{
				{Domains: Domains{Domain("example.com")}},
			},
			want: Certificates{cert2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, certificates.UnusedCertificates(tt.domainsRequests), "UnusedCertificates(%v)", tt.domainsRequests)
		})
	}
}

func TestCertificates_Deletes(t *testing.T) {
	cert1 := &Certificate{Identifier: "example.com-0", Domains: Domains{Domain("example.com")}}
	cert2 := &Certificate{Identifier: "example2.com-0", Domains: Domains{Domain("example2.com")}}
	certificates := Certificates{cert1, cert2}
	removeCertificates := Certificates{cert2}
	want := Certificates{cert1}
	got := certificates.Deletes(removeCertificates)
	assert.Equal(t, want, got)
}

func TestGetCertificateFilename(t *testing.T) {
	identifier := "foo"
	assert.Equal(t, "foo.crt", GetCertificateFilename(identifier))
}

func TestGetKeyFilename(t *testing.T) {
	identifier := "foo"
	assert.Equal(t, "foo.key", GetKeyFilename(identifier))
}
