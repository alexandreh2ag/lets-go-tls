package types

import (
	"fmt"
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

func TestCertificate_GetPemFilename(t *testing.T) {
	c := Certificate{Identifier: "foo"}
	assert.Equal(t, "foo.pem", c.GetPemFilename())
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

func TestCertificates_UsedCertificates(t *testing.T) {
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
			want: Certificates{cert1, cert2},
		},
		{
			name: "SuccessOneNotUsed",
			domainsRequests: []*DomainRequest{
				{Domains: Domains{Domain("example.com")}},
			},
			want: Certificates{cert1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, certificates.UsedCertificates(tt.domainsRequests), "UsedCertificates(%v)", tt.domainsRequests)
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

func TestGetPemFilename(t *testing.T) {
	identifier := "foo"
	assert.Equal(t, "foo.pem", GetPemFilename(identifier))
}

func TestGetX509Certificate(t *testing.T) {
	certPEM := []byte(`
-----BEGIN CERTIFICATE-----
MIIB1zCCAUCgAwIBAgIBATANBgkqhkiG9w0BAQsFADAWMRQwEgYDVQQKEwtFeGVt
cGxlIE9yZzAeFw0yNTA2MjIxMTU1MDFaFw0yNjA2MjIxMTU1MDFaMBYxFDASBgNV
BAoTC0V4ZW1wbGUgT3JnMIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQC+xqcl
8nmNTvPEiWpdJyWpsJSr8DTb6xPFnb1+I+ACzFS6Qyv7pT8RSTPuy/2DC4cSeqsV
HiaBSNxQxoWODwgE41/dVx0p0G8US2Ds7M5TCGm3zHnz5oABagjcaI2ZT8YEveNL
5w0Xstj6RmoytkGhEosDV04CZyhDWwzQB7+dkwIDAQABozUwMzAOBgNVHQ8BAf8E
BAMCBaAwEwYDVR0lBAwwCgYIKwYBBQUHAwEwDAYDVR0TAQH/BAIwADANBgkqhkiG
9w0BAQsFAAOBgQBnOYaKCp3VfgZD925AydTQ5PuUeo4LK5pE8AKLfOrgHvTzui1/
34zByxGjWHv7p4rX2txg9EpN5BuAIvddIn3eMb802+FqjznbGVaMQ1iaftGlJnir
B0JJs8CQfk8HQCPK5pdaZrsc+1cWqPDuUSpuDTi06NGbPD7xYGIwcKoEAg==
-----END CERTIFICATE-----
`)
	tests := []struct {
		name            string
		cert            []byte
		wantRawNotEmpty bool
		wantErr         assert.ErrorAssertionFunc
	}{
		{
			name:            "Success",
			cert:            certPEM,
			wantRawNotEmpty: true,
			wantErr:         assert.NoError,
		},
		{
			name:            "FailParsePem",
			cert:            []byte("wrong"),
			wantRawNotEmpty: false,
			wantErr:         assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetX509Certificate(tt.cert)
			if !tt.wantErr(t, err, fmt.Sprintf("GetX509Certificate(%v)", tt.cert)) {
				return
			}

			if tt.wantRawNotEmpty {
				assert.NotEmpty(t, got.Raw)
			}
		})
	}
}

func TestCertificate_GetPemContent(t *testing.T) {
	want := []byte(`-----BEGIN CERTIFICATE-----
MIIC+zCCAeOgAwIBAgIJAO0r1z8wDQYJKoZIhvcNAQELBQAwEjEQMA4GA1UEAwwH
...
-----END CERTIFICATE-----

-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASC...
-----END PRIVATE KEY-----
`)
	cert := []byte(`-----BEGIN CERTIFICATE-----
MIIC+zCCAeOgAwIBAgIJAO0r1z8wDQYJKoZIhvcNAQELBQAwEjEQMA4GA1UEAwwH
...
-----END CERTIFICATE-----`)

	key := []byte(`-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASC...
-----END PRIVATE KEY-----`)
	c := Certificate{

		Certificate: cert,
		Key:         key,
	}
	got := c.GetPemContent()
	assert.Equalf(t, string(want), string(got), "GetPemContent()")
}
