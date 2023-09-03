package types

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDomain_FormatSubdomainToWildcard(t *testing.T) {
	tests := []struct {
		name string
		d    Domain
		want Domain
	}{
		{
			name: "RootDomain",
			d:    Domain("example.com"),
			want: Domain("example.com"),
		},
		{
			name: "RootDomainMultiTldDot",
			d:    Domain("example.co.uk"),
			want: Domain("*.co.uk"),
		},
		{
			name: "Subdomain",
			d:    Domain("foo.example.com"),
			want: Domain("*.example.com"),
		},
		{
			name: "SubdomainMultiTldDot",
			d:    Domain("foo.example.co.uk"),
			want: Domain("*.example.co.uk"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.d.FormatSubdomainToWildcard(), "FormatSubdomainToWildcard()")
		})
	}
}

func TestDomains_ToStringSlice(t *testing.T) {
	domains := Domains{"example.com", "api.example.fr"}
	want := []string{"example.com", "api.example.fr"}
	assert.Equal(t, want, domains.ToStringSlice())
}

func TestDomains_Sort(t *testing.T) {
	tests := []struct {
		name string
		d    Domains
		want Domains
	}{
		{
			name: "Success",
			d:    Domains{"example.b", "example.ba", "example.a", "example.c"},
			want: Domains{"example.a", "example.b", "example.ba", "example.c"},
		},
		{
			name: "SuccessAlreadySorted",
			d:    Domains{"example.a", "example.b", "example.c"},
			want: Domains{"example.a", "example.b", "example.c"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.d.Sort()
			assert.Equal(t, tt.want, tt.d)
		})
	}
}

func TestSortDomainsRequests(t *testing.T) {

	tests := []struct {
		name            string
		domainsRequests []*DomainRequest
		want            []*DomainRequest
	}{
		{
			name: "Success",
			domainsRequests: []*DomainRequest{
				{Domains: Domains{"example.a"}},
				{Domains: Domains{"foo.foo.a", "example.a", "*.example.a"}},
				{Domains: Domains{"foo.foo.a", "example.a", "a.example.a"}},
				{Domains: Domains{"foo.example.b", "bar.example.b"}},
				{Domains: Domains{"*.example.b"}},
				{Domains: Domains{"a.example.a"}},
				{Domains: Domains{"example.cb"}},
				{Domains: Domains{"example.z"}},
				{Domains: Domains{"example.c", "example.cb"}},
				{Domains: Domains{"example.b", "example.a"}},
			},
			want: []*DomainRequest{
				{Domains: Domains{"*.example.a", "example.a", "foo.foo.a"}},
				{Domains: Domains{"*.example.b"}},
				{Domains: Domains{"a.example.a", "example.a", "foo.foo.a"}},
				{Domains: Domains{"bar.example.b", "foo.example.b"}},
				{Domains: Domains{"example.a", "example.b"}},
				{Domains: Domains{"example.c", "example.cb"}},
				{Domains: Domains{"a.example.a"}},
				{Domains: Domains{"example.a"}},
				{Domains: Domains{"example.cb"}},
				{Domains: Domains{"example.z"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SortDomainsRequests(tt.domainsRequests)
			assert.Equal(t, tt.want, tt.domainsRequests)
		})
	}
}

func TestDomain_IsWildcard(t *testing.T) {
	tests := []struct {
		name string
		d    Domain
		want bool
	}{
		{
			name: "RootDomain",
			d:    Domain("example.com"),
			want: false,
		},
		{
			name: "DomainNotWildcard",
			d:    Domain("foo.example.com"),
			want: false,
		},
		{
			name: "DomainWildcard",
			d:    Domain("*.example.com"),
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.d.IsWildcard(), "IsWildcard()")
		})
	}
}

func TestDomains_ContainsWildcard(t *testing.T) {
	tests := []struct {
		name string
		ds   Domains
		want bool
	}{
		{
			name: "RootDomain",
			ds:   Domains{"example.com"},
			want: false,
		},
		{
			name: "DomainNotWildcard",
			ds:   Domains{"foo.example.com", "bar.example.com"},
			want: false,
		},
		{
			name: "DomainWildcard",
			ds:   Domains{"foo.example.com", "*.example.com"},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, tt.ds.ContainsWildcard(), "ContainsWildcard()")
		})
	}
}
