package types

import (
	"fmt"
	"slices"
	"time"
)

type Certificates []*Certificate

func (c Certificates) Match(request *DomainRequest, onlyValid bool) *Certificate {
	for _, certificate := range c {
		if onlyValid && !certificate.IsValid() {
			continue
		}
		if certificate.Match(request.Domains) {
			return certificate
		}
	}
	return nil
}

func (c Certificates) CheckIdentifierUnique(identifier string) bool {
	if identifier == "" {
		return false
	}

	if c.GetCertificate(identifier) != nil {
		return false
	}
	return true
}

func (c Certificates) GetCertificate(identifier string) *Certificate {
	for _, certificate := range c {
		if certificate.Identifier == identifier {
			return certificate
		}
	}
	return nil
}

func (c Certificates) UnusedCertificates(domainsRequests []*DomainRequest) Certificates {
	unusedCertificates := Certificates{}
	for _, cert := range c {
		found := false
		for _, request := range domainsRequests {
			if cert.Match(request.Domains) {
				found = true
				break
			}
		}

		if !found {
			unusedCertificates = append(unusedCertificates, cert)
		}
	}
	return unusedCertificates
}

func (c Certificates) Deletes(removeCertificates Certificates) Certificates {
	return slices.DeleteFunc(c, func(certificate *Certificate) bool {
		if removeCertificates.GetCertificate(certificate.Identifier) != nil {
			return true
		}
		return false
	})
}

type Certificate struct {
	Identifier     string    `json:"identifier,omitempty"`
	Main           string    `json:"main,omitempty"`
	Domains        Domains   `json:"domain,omitempty"`
	Certificate    []byte    `json:"certificate,omitempty"`
	Key            []byte    `json:"key,omitempty"`
	ExpirationDate time.Time `json:"expiration_date,omitempty"`

	ObtainFailCount int       `json:"obtain_fail_count,omitempty"`
	ObtainFailDate  time.Time `json:"obtain_fail_date,omitempty"`

	UnusedAt time.Time `json:"unused_at,omitempty"`
}

func (c Certificate) IsValid() bool {
	if c.Key == nil || c.Certificate == nil {
		return false
	}
	return true
}

func (c Certificate) GetKeyFilename() string {
	return GetKeyFilename(c.Identifier)
}

func (c Certificate) GetCertificateFilename() string {
	return GetCertificateFilename(c.Identifier)
}

func (c Certificate) Match(domains Domains) bool {
	if len(c.Domains) > 0 && len(domains) > 0 {
		for _, domain := range domains {
			if !slices.Contains(c.Domains, domain) && !slices.Contains(c.Domains, domain.FormatSubdomainToWildcard()) {
				return false
			}
		}
		return true
	}
	return false
}

func GetKeyFilename(identifier string) string {
	return fmt.Sprintf("%s.%s", identifier, "key")
}

func GetCertificateFilename(identifier string) string {
	return fmt.Sprintf("%s.%s", identifier, "crt")
}
