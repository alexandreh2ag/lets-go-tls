#!/bin/bash

go install go.uber.org/mock/mockgen@latest

rm -rf mocks
mockgen -destination=mocks/types/types.go -package=mockTypes github.com/alexandreh2ag/lets-go-tls/types Requester,Cache,Resolver
mockgen -destination=mocks/types/acme/acme.go -package=mockTypesAcme github.com/alexandreh2ag/lets-go-tls/types/acme Challenge
mockgen -destination=mocks/types/storage/state/state.go -package=mockTypesStorageState github.com/alexandreh2ag/lets-go-tls/types/storage/state Storage
mockgen -destination=mocks/types/storage/certificate/storage.go -package=mockTypesStorageCertificate github.com/alexandreh2ag/lets-go-tls/types/storage/certificate Storage
mockgen -destination=mocks/http/http.go -package=mockHttp github.com/alexandreh2ag/lets-go-tls/http Client
mockgen -destination=mocks/prometheus/registry.go -package=mockPrometheus github.com/alexandreh2ag/lets-go-tls/prometheus Registry
mockgen -destination=mocks/afero/fs.go -package=mockAfero github.com/spf13/afero Fs
