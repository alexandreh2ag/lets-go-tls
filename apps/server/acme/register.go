package acme

import (
	"fmt"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/alexandreh2ag/lets-go-tls/types/storage/state"
	"github.com/go-acme/lego/v4/registration"
)

func RegisterAccount(state *types.State, stateStorage state.Storage, defaultResolver types.Resolver) error {
	if state.Account.Registration == nil {
		// create private key + email
		reg, errRegister := defaultResolver.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
		if errRegister != nil {
			return fmt.Errorf("error when register ACME account: %v", errRegister)
		}
		state.Account.Registration = reg
		// lock state
		err := stateStorage.Save(state)
		// unlock state
		if err != nil {
			return fmt.Errorf("failed to save state: %v", err)
		}
	}
	return nil
}
