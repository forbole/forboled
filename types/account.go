package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

// var _ auth.Account = (*AppAccount)(nil)

// Custom extensions for this application.  This is just an example of
// extending auth.BaseAccount with custom fields.
//
// This is compatible with the stock auth.AccountStore, since
// auth.AccountStore uses the flexible go-amino library.
type AppAccount struct {
	auth.BaseAccount
	Name string `json:"name"`
}

// nolint
func (acc AppAccount) GetName() string      { return acc.Name }
func (acc *AppAccount) SetName(name string) { acc.Name = name }

// Get the AccountDecoder function for the custom AppAccount
func GetAccountDecoder(cdc *wire.Codec) auth.AccountDecoder {
	return func(accBytes []byte) (res auth.Account, err error) {
		if len(accBytes) == 0 {
			return nil, sdk.ErrTxDecode("accBytes are empty")
		}
		acct := new(AppAccount)
		err = cdc.UnmarshalBinaryBare(accBytes, &acct)
		if err != nil {
			panic(err)
		}
		return acct, err
	}
}

//___________________________________________________________________________________

type ReputeAccount struct {
	auth.BaseAccount
	Name   string `json:"name"`
	Repute int64  `json:"repute"`
	Role   string `json:"role"`
}

func ProtoReputeAccount() auth.Account {
	return &ReputeAccount{}
}

// nolint
func (acc ReputeAccount) GetName() string         { return acc.Name }
func (acc *ReputeAccount) SetName(name string)    { acc.Name = name }
func (acc ReputeAccount) GetRepute() int64        { return acc.Repute }
func (acc *ReputeAccount) SetRepute(repute int64) { acc.Repute = repute }
func (acc ReputeAccount) GetRole() string         { return acc.Role }
func (acc *ReputeAccount) SetRole(role string)    { acc.Role = role }

// Get the AccountDecoder function for the ReputeAccount
func GetReputeAccountDecoder(cdc *wire.Codec) auth.AccountDecoder {
	return func(accBytes []byte) (res auth.Account, err error) {
		if len(accBytes) == 0 {
			return nil, sdk.ErrTxDecode("accBytes are empty")
		}
		acct := new(ReputeAccount)
		err = cdc.UnmarshalBinaryBare(accBytes, &acct)
		if err != nil {
			panic(err)
		}
		return acct, err
	}
}
