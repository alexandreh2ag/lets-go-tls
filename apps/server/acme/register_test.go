package acme

import (
	"errors"
	mockTypes "github.com/alexandreh2ag/lets-go-tls/mocks/types"
	mockTypesStorageState "github.com/alexandreh2ag/lets-go-tls/mocks/types/storage/state"
	"github.com/alexandreh2ag/lets-go-tls/types"
	"github.com/alexandreh2ag/lets-go-tls/types/acme"
	"github.com/go-acme/lego/v4/registration"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"testing"
)

func TestRegisterAccount_SuccessRegister(t *testing.T) {
	ctrl := gomock.NewController(t)
	state := &types.State{Account: &acme.Account{Registration: nil}}
	resolver := mockTypes.NewMockResolver(ctrl)
	resolver.EXPECT().Register(gomock.Any()).Times(1).Return(&registration.Resource{}, nil)
	stateStorage := mockTypesStorageState.NewMockStorage(ctrl)
	stateStorage.EXPECT().Save(gomock.Any()).Times(1).Return(nil)

	err := RegisterAccount(state, stateStorage, resolver)
	assert.NoError(t, err)
}

func TestRegisterAccount_SuccessNoRegister(t *testing.T) {
	ctrl := gomock.NewController(t)
	state := &types.State{Account: &acme.Account{Registration: &registration.Resource{}}}
	resolver := mockTypes.NewMockResolver(ctrl)
	stateStorage := mockTypesStorageState.NewMockStorage(ctrl)

	err := RegisterAccount(state, stateStorage, resolver)
	assert.NoError(t, err)
}

func TestRegisterAccount_FailRegister(t *testing.T) {
	ctrl := gomock.NewController(t)
	state := &types.State{Account: &acme.Account{Registration: nil}}
	resolver := mockTypes.NewMockResolver(ctrl)
	resolver.EXPECT().Register(gomock.Any()).Times(1).Return(nil, errors.New("error register"))
	stateStorage := mockTypesStorageState.NewMockStorage(ctrl)

	err := RegisterAccount(state, stateStorage, resolver)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error register")
}

func TestRegisterAccount_FailSave(t *testing.T) {
	ctrl := gomock.NewController(t)
	state := &types.State{Account: &acme.Account{Registration: nil}}
	resolver := mockTypes.NewMockResolver(ctrl)
	resolver.EXPECT().Register(gomock.Any()).Times(1).Return(&registration.Resource{}, nil)
	stateStorage := mockTypesStorageState.NewMockStorage(ctrl)
	stateStorage.EXPECT().Save(gomock.Any()).Times(1).Return(errors.New("error save"))

	err := RegisterAccount(state, stateStorage, resolver)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error save")
}
