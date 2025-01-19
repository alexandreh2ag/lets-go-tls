package types

import (
	"cmp"
	"net"
	"slices"
	"strings"
)

type Domain string

func (d Domain) FormatSubdomainToWildcard() Domain {
	split := strings.Split(string(d), ".")
	if len(split) >= 3 {
		split[0] = "*"
		return Domain(strings.Join(split, "."))
	}

	return d
}

func (d Domain) IsWildcard() bool {
	return strings.HasPrefix(string(d), "*")
}

type Domains []Domain

func (ds Domains) ToStringSlice() []string {
	domains := []string{}
	for _, domain := range ds {
		domains = append(domains, string(domain))
	}
	return domains
}

func (ds Domains) Sort() {
	slices.SortFunc(ds, func(a, b Domain) int {
		return cmp.Compare(string(a), string(b))
	})
}

func (ds Domains) ContainsWildcard() bool {
	for _, d := range ds {
		if d.IsWildcard() {
			return true
		}
	}
	return false
}

type DomainRequest struct {
	Domains   Domains   `json:"domains"`
	Requester Requester `json:"-"`
}

func (dr DomainRequest) IsIP() bool {
	for _, domain := range dr.Domains {
		ip := net.ParseIP(string(domain))
		if ip != nil {
			return true
		}
	}
	return false
}

func SortDomainsRequests(domainsRequests []*DomainRequest) {
	slices.SortFunc(domainsRequests, func(a, b *DomainRequest) int {
		a.Domains.Sort()
		b.Domains.Sort()
		aContainsWildcard := 0
		bContainsWildcard := 0

		if a.Domains.ContainsWildcard() {
			aContainsWildcard++
		}

		if b.Domains.ContainsWildcard() {
			bContainsWildcard++
		}
		return cmp.Or(
			-cmp.Compare(aContainsWildcard, bContainsWildcard),
			-cmp.Compare(len(a.Domains.ToStringSlice()), len(b.Domains.ToStringSlice())),
			cmp.Compare(strings.Join(a.Domains.ToStringSlice(), ","), strings.Join(b.Domains.ToStringSlice(), ",")),
		)
	})
}
