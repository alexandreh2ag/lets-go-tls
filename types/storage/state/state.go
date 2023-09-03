package state

import (
	"github.com/alexandreh2ag/lets-go-tls/types"
)

type Storage interface {
	Type() string
	Load() (*types.State, error)
	Save(state *types.State) error
}
