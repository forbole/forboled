package app

import (
	"encoding/json"
	"errors"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	crypto "github.com/tendermint/go-crypto"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/server"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/stake"

	"github.com/forbole/forboled/types"
)

// State to Unmarshal
type GenesisState struct {
	Accounts  []GenesisAccount   `json:"accounts"`
	Admins    []GenesisAdmin     `json:"admins"`
	StakeData stake.GenesisState `json:"stake"`
}

// GenesisAccount doesn't need pubkey or sequence
type GenesisAccount struct {
	Address sdk.Address `json:"address"`
	Coins   sdk.Coins   `json:"coins"`
}

func NewGenesisAccount(acc *auth.BaseAccount) GenesisAccount {
	return GenesisAccount{
		Address: acc.Address,
		Coins:   acc.Coins,
	}
}

func NewGenesisAccountI(acc auth.Account) GenesisAccount {
	return GenesisAccount{
		Address: acc.GetAddress(),
		Coins:   acc.GetCoins(),
	}
}

// convert GenesisAccount to auth.BaseAccount
func (ga *GenesisAccount) ToAccount() (acc *auth.BaseAccount) {
	return &auth.BaseAccount{
		Address: ga.Address,
		Coins:   ga.Coins.Sort(),
	}
}

// GenesisAdmin doesn't need pubkey or sequence
type GenesisAdmin struct {
	Address sdk.Address `json:"address"`
	// Repute  int64       `json:"repute"`
	Role string `json:"role"`
}

func NewGenesisAdmin(acc *auth.BaseAccount) GenesisAdmin {
	return GenesisAdmin{
		Address: acc.Address,
		Role:    "Admin",
	}
}

// convert GenesisAdmin to ReputeAccount
func (ga *GenesisAdmin) ToReputeAccount() (acc *types.ReputeAccount) {
	baseAcc := auth.BaseAccount{
		Address: ga.Address,
		Coins:   nil,
	}
	return &types.ReputeAccount{
		BaseAccount: baseAcc,
		// Repute:      ga.Repute,
		Role: ga.Role,
	}
}

var (
	flagName       = "name"
	flagClientHome = "home-client"
	flagOWK        = "owk"

	// bonded tokens given to genesis validators/accounts
	freeFermionVal  = int64(100)
	freeFermionsAcc = int64(50)
)

// get app init parameters for server init command
func ForboleAppInit() server.AppInit {
	fsAppGenState := pflag.NewFlagSet("", pflag.ContinueOnError)

	fsAppGenTx := pflag.NewFlagSet("", pflag.ContinueOnError)
	fsAppGenTx.String(flagName, "", "validator moniker, if left blank, do not add validator")
	fsAppGenTx.String(flagClientHome, DefaultCLIHome, "home directory for the client, used for key generation")
	fsAppGenTx.Bool(flagOWK, false, "overwrite the accounts created")

	return server.AppInit{
		FlagsAppGenState: fsAppGenState,
		FlagsAppGenTx:    fsAppGenTx,
		AppGenTx:         ForboleAppGenTx,
		AppGenState:      ForboleAppGenState,
	}
}

// simple genesis tx
type ForboleGenTx struct {
	Name    string        `json:"name"`
	Address sdk.Address   `json:"address"`
	PubKey  crypto.PubKey `json:"pub_key"`
}

// Generate a forbole genesis transaction
func ForboleAppGenTx(cdc *wire.Codec, pk crypto.PubKey) (
	appGenTx, cliPrint json.RawMessage, validator tmtypes.GenesisValidator, err error) {

	var addr sdk.Address
	var secret string
	clientRoot := viper.GetString(flagClientHome)
	overwrite := viper.GetBool(flagOWK)
	name := viper.GetString(flagName)
	addr, secret, err = server.GenerateSaveCoinKey(clientRoot, name, "1234567890", overwrite)
	if err != nil {
		return
	}

	var bz []byte
	forboleGenTx := ForboleGenTx{
		Name:    name,
		Address: addr,
		PubKey:  pk,
	}
	bz, err = wire.MarshalJSONIndent(cdc, forboleGenTx)
	if err != nil {
		return
	}
	appGenTx = json.RawMessage(bz)

	mm := map[string]string{"secret": secret}
	bz, err = cdc.MarshalJSON(mm)
	if err != nil {
		return
	}
	cliPrint = json.RawMessage(bz)

	validator = tmtypes.GenesisValidator{
		PubKey: pk,
		Power:  freeFermionVal,
	}
	return
}

// Create the core parameters for genesis initialization for forbole
// note that the pubkey input is this machines pubkey
func ForboleAppGenState(cdc *wire.Codec, appGenTxs []json.RawMessage) (appState json.RawMessage, err error) {

	if len(appGenTxs) == 0 {
		err = errors.New("must provide at least genesis transaction")
		return
	}

	// start with the default staking genesis state
	stakeData := stake.DefaultGenesisState()

	// get genesis flag account information
	genaccs := make([]GenesisAccount, len(appGenTxs))
	admins := make([]GenesisAdmin, len(appGenTxs))
	for i, appGenTx := range appGenTxs {

		var genTx ForboleGenTx
		err = cdc.UnmarshalJSON(appGenTx, &genTx)
		if err != nil {
			return
		}

		// create the genesis account, give'm few steaks and a buncha token with there name
		accAuth := auth.NewBaseAccountWithAddress(genTx.Address)
		accAuth.Coins = sdk.Coins{
			{"money", 1000000},
			{"steak", freeFermionsAcc},
		}
		acc := NewGenesisAccount(&accAuth)
		genaccs[i] = acc
		admin := NewGenesisAdmin(&accAuth)
		admins[i] = admin
		stakeData.Pool.LooseUnbondedTokens += freeFermionsAcc // increase the supply

		// add the validator
		if len(genTx.Name) > 0 {
			desc := stake.NewDescription(genTx.Name, "", "", "")
			validator := stake.NewValidator(genTx.Address, genTx.PubKey, desc)
			validator.PoolShares = stake.NewBondedShares(sdk.NewRat(freeFermionVal))
			stakeData.Validators = append(stakeData.Validators, validator)

			// pool logic
			stakeData.Pool.BondedTokens += freeFermionVal
			stakeData.Pool.BondedShares = sdk.NewRat(stakeData.Pool.BondedTokens)
		}
	}

	// create the final app state
	genesisState := GenesisState{
		Accounts:  genaccs,
		Admins:    admins,
		StakeData: stakeData,
	}
	appState, err = wire.MarshalJSONIndent(cdc, genesisState)
	return
}
