package app

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/server/config"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/stake"

	"github.com/spf13/pflag"
	"github.com/tendermint/tendermint/crypto"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/forbole/forboled/types"
)

// DefaultKeyPass contains the default key password for genesis transactions
const DefaultKeyPass = "12345678"

var (
	flagName       = "name"
	flagClientHome = "home-client"
	flagOWK        = "owk"

	// bonded tokens given to genesis validators/accounts
	freeFermionVal  = int64(100)
	freeFermionsAcc = int64(50)
)

// State to Unmarshal
type GenesisState struct {
	// cosmos 0.18.0rc it's []*GenesisAccount
	Accounts  []GenesisAccount   `json:"accounts"`
	Admins    []GenesisAdmin     `json:"admins"`
	StakeData stake.GenesisState `json:"stake"`
}

// GenesisAccount doesn't need pubkey or sequence
type GenesisAccount struct {
	Address sdk.AccAddress `json:"address"`
	Coins   sdk.Coins      `json:"coins"`
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
	Address sdk.AccAddress `json:"address"`
	Role    string         `json:"role"`
	// Repute  int64       `json:"repute"`
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

// get app init parameters for server init command
func ForboleAppInit() server.AppInit {
	fsAppGenState := pflag.NewFlagSet("", pflag.ContinueOnError)

	fsAppGenTx := pflag.NewFlagSet("", pflag.ContinueOnError)
	fsAppGenTx.String(server.FlagName, "", "validator moniker, required")
	fsAppGenTx.String(server.FlagClientHome, DefaultCLIHome,
		"home directory for the client, used for key generation")
	fsAppGenTx.Bool(server.FlagOWK, false, "overwrite the accounts created")

	return server.AppInit{
		FlagsAppGenState: fsAppGenState,
		FlagsAppGenTx:    fsAppGenTx,
		AppGenTx:         ForboleAppGenTx,
		AppGenState:      ForboleAppGenStateJSON,
	}
}

// simple genesis tx
type ForboleGenTx struct {
	Name    string         `json:"name"`
	Address sdk.AccAddress `json:"address"`
	PubKey  string         `json:"pub_key"`
}

// Generate a forbole genesis transaction
// GaiaAppGenTx generates a Gaia genesis transaction.
func ForboleAppGenTx(
	cdc *wire.Codec, pk crypto.PubKey, genTxConfig config.GenTx,
) (appGenTx, cliPrint json.RawMessage, validator tmtypes.GenesisValidator, err error) {
	if genTxConfig.Name == "" {
		return nil, nil, tmtypes.GenesisValidator{}, errors.New("Must specify --name (validator moniker)")
	}

	buf := client.BufferStdin()
	prompt := fmt.Sprintf("Password for account '%s' (default %s):", genTxConfig.Name, DefaultKeyPass)
	keyPass, err := client.GetPassword(prompt, buf)
	if err != nil && keyPass != "" {
		// An error was returned that either failed to read the password from
		// STDIN or the given password is not empty but failed to meet minimum
		// length requirements.
		return appGenTx, cliPrint, validator, err
	}
	if keyPass == "" {
		keyPass = DefaultKeyPass
	}
	addr, secret, err := server.GenerateSaveCoinKey(
		genTxConfig.CliRoot,
		genTxConfig.Name,
		keyPass,
		genTxConfig.Overwrite,
	)
	if err != nil {
		return appGenTx, cliPrint, validator, err
	}
	mm := map[string]string{"secret": secret}
	bz, err := cdc.MarshalJSON(mm)
	if err != nil {
		return appGenTx, cliPrint, validator, err
	}
	cliPrint = json.RawMessage(bz)

	appGenTx, _, validator, err = ForboleAppGenTxNF(cdc, pk, addr, genTxConfig.Name)

	return appGenTx, cliPrint, validator, err
}

func ForboleAppGenTxNF(cdc *wire.Codec, pk crypto.PubKey, addr sdk.AccAddress, name string) (
	appGenTx, cliPrint json.RawMessage, validator tmtypes.GenesisValidator, err error) {

	var bz []byte
	forboleGenTx := ForboleGenTx{
		Name:    name,
		Address: addr,
		PubKey:  sdk.MustBech32ifyAccPub(pk),
	}
	bz, err = wire.MarshalJSONIndent(cdc, forboleGenTx)
	if err != nil {
		return
	}
	appGenTx = json.RawMessage(bz)

	validator = tmtypes.GenesisValidator{
		PubKey: pk,
		Power:  freeFermionVal,
	}
	return
}

// Create the core parameters for genesis initialization for forbole
// note that the pubkey input is this machines pubkey
func ForboleAppGenState(cdc *wire.Codec, appGenTxs []json.RawMessage) (genesisState GenesisState, err error) {

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
			{"money", sdk.NewInt(1000000)},
			{"steak", sdk.NewInt(freeFermionsAcc)},
		}
		acc := NewGenesisAccount(&accAuth)
		genaccs[i] = acc
		admin := NewGenesisAdmin(&accAuth)
		admins[i] = admin
		stakeData.Pool.LooseTokens = stakeData.Pool.LooseTokens.Add(sdk.NewRat(freeFermionsAcc)) // increase the supply

		// add the validator
		if len(genTx.Name) > 0 {
			desc := stake.NewDescription(genTx.Name, "", "", "")
			validator := stake.NewValidator(genTx.Address,
				sdk.MustGetAccPubKeyBech32(genTx.PubKey), desc)

			stakeData.Pool.LooseTokens = stakeData.Pool.LooseTokens.Add(sdk.NewRat(freeFermionVal)) // increase the supply

			// add some new shares to the validator
			var issuedDelShares sdk.Rat
			validator, stakeData.Pool, issuedDelShares = validator.AddTokensFromDel(stakeData.Pool, freeFermionVal)
			stakeData.Validators = append(stakeData.Validators, validator)

			// create the self-delegation from the issuedDelShares
			delegation := stake.Delegation{
				DelegatorAddr: validator.Owner,
				ValidatorAddr: validator.Owner,
				Shares:        issuedDelShares,
				Height:        0,
			}

			stakeData.Bonds = append(stakeData.Bonds, delegation)
		}
	}

	// create the final app state
	genesisState = GenesisState{
		Accounts:  genaccs,
		Admins:    admins,
		StakeData: stakeData,
	}
	// appState, err = wire.MarshalJSONIndent(cdc, genesisState)
	return
}

// ForboleAppGenState but with JSON
func ForboleAppGenStateJSON(cdc *wire.Codec, appGenTxs []json.RawMessage) (appState json.RawMessage, err error) {

	// create the final app state
	genesisState, err := ForboleAppGenState(cdc, appGenTxs)
	if err != nil {
		return nil, err
	}
	appState, err = wire.MarshalJSONIndent(cdc, genesisState)
	return
}
