package http

import "github.com/alexandreh2ag/lets-go-tls/types"

type ResponseCertificatesFromRequests struct {
	Certificates types.Certificates `json:"certificates"`
	Requests     ResponseRequests   `json:"requests"`
}

type ResponseRequests struct {
	Found    []*types.DomainRequest `json:"found"`
	NotFound []*types.DomainRequest `json:"not_found"`
}
