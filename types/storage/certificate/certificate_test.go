package certificate

import (
	"github.com/alexandreh2ag/lets-go-tls/hook"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/stretchr/testify/assert"
	"testing"
)

var _ Storage = &dummyStorage{}

type dummyStorage struct {
}

func (d dummyStorage) ID() string {
	panic("implement me")
}

func (d dummyStorage) Save(certificates types.Certificates, _ chan<- *hook.Hook) []error {
	panic("implement me")
}

func (d dummyStorage) Delete(certificates types.Certificates, _ chan<- *hook.Hook) []error {
	panic("implement me")
}

func TestStorages_Get_SuccessFound(t *testing.T) {
	storage := &dummyStorage{}
	storages := Storages{"foo": storage}
	got := storages.Get("foo")
	assert.Equal(t, storage, got)
}

func TestStorages_Get_SuccessNotFound(t *testing.T) {
	storages := Storages{}
	got := storages.Get("foo")
	assert.Nil(t, got)
}
