package certificate

import "github.com/alexandreh2ag/lets-go-tls/types"

type Storages map[string]Storage

func (s Storages) Get(key string) Storage {
	if storage, ok := s[key]; ok {
		return storage
	}
	return nil
}

type Storage interface {
	ID() string
	Save(certificates types.Certificates) []error
	Delete(certificates types.Certificates) []error
}
