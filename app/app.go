package app

import (
	"encoding/json"
	"os"

	abci "github.com/tendermint/abci/types"
	cmn "github.com/tendermint/tmlibs/common"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/ibc"
	"github.com/cosmos/cosmos-sdk/x/stake"
	bam "github.com/forbole/forboled/baseapp"

	"github.com/forbole/forboled/types"
	"github.com/forbole/forboled/x/contrib"
)

const (
	appName = "ForboleApp"
)

// default home directories for expected binaries
var (
	DefaultCLIHome  = os.ExpandEnv("$HOME/.fbcli")
	DefaultNodeHome = os.ExpandEnv("$HOME/.forboled")
)

// Extended ABCI application
type ForboleApp struct {
	*bam.BaseApp
	cdc *wire.Codec

	// keys to access the substores
	keyMain    *sdk.KVStoreKey
	keyAccount *sdk.KVStoreKey
	keyIBC     *sdk.KVStoreKey
	keyStake   *sdk.KVStoreKey
	keyContrib *sdk.KVStoreKey
	keyRepute  *sdk.KVStoreKey

	// Manage getting and setting accounts
	accountMapper       auth.AccountMapper
	feeCollectionKeeper auth.FeeCollectionKeeper
	reputeAccountMapper auth.AccountMapper
	coinKeeper          bank.Keeper
	ibcMapper           ibc.Mapper
	stakeKeeper         stake.Keeper
	contribKeeper       contrib.Keeper
}

func NewForboleApp(logger log.Logger, db dbm.DB) *ForboleApp {

	// Create app-level codec for txs and accounts.
	var cdc = MakeCodec()

	// Create your application object.
	var app = &ForboleApp{
		BaseApp:    bam.NewBaseApp(appName, cdc, logger, db),
		cdc:        cdc,
		keyMain:    sdk.NewKVStoreKey("main"),
		keyAccount: sdk.NewKVStoreKey("acc"),
		keyIBC:     sdk.NewKVStoreKey("ibc"),
		keyStake:   sdk.NewKVStoreKey("stake"),
		keyContrib: sdk.NewKVStoreKey("contrib"),
		keyRepute:  sdk.NewKVStoreKey("repute"),
	}

	// Define the accountMapper.
	app.accountMapper = auth.NewAccountMapper(
		app.cdc,
		app.keyAccount,      // target store
		&auth.BaseAccount{}, // prototype
	)
	app.reputeAccountMapper = auth.NewAccountMapper(
		app.cdc,
		app.keyRepute,          // target store
		&types.ReputeAccount{}, // prototype
	)

	// Add handlers.
	app.coinKeeper = bank.NewKeeper(app.accountMapper)
	app.ibcMapper = ibc.NewMapper(app.cdc, app.keyIBC, app.RegisterCodespace(ibc.DefaultCodespace))
	app.stakeKeeper = stake.NewKeeper(app.cdc, app.keyStake, app.coinKeeper, app.RegisterCodespace(stake.DefaultCodespace))
	app.contribKeeper = contrib.NewKeeper(app.cdc, app.reputeAccountMapper, app.keyContrib)
	app.Router().
		AddRoute("bank", bank.NewHandler(app.coinKeeper)).
		AddRoute("ibc", ibc.NewHandler(app.ibcMapper, app.coinKeeper)).
		AddRoute("stake", stake.NewHandler(app.stakeKeeper)).
		AddRoute("contrib", contrib.NewHandler(app.contribKeeper))

	// Initialize BaseApp.
	app.SetInitChainer(app.initChainer)
	app.SetEndBlocker(stake.NewEndBlocker(app.stakeKeeper))
	app.MountStoresIAVL(app.keyMain, app.keyAccount, app.keyIBC, app.keyStake, app.keyContrib, app.keyRepute)
	app.SetAnteHandlers(auth.NewAnteHandler(app.accountMapper, app.feeCollectionKeeper), auth.NewAnteHandler(app.reputeAccountMapper, app.feeCollectionKeeper))
	err := app.LoadLatestVersion(app.keyMain)
	if err != nil {
		cmn.Exit(err.Error())
	}

	return app
}

// Custom tx codec
func MakeCodec() *wire.Codec {
	var cdc = wire.NewCodec()
	wire.RegisterCrypto(cdc) // Register crypto.
	sdk.RegisterWire(cdc)    // Register Msgs
	bank.RegisterWire(cdc)
	stake.RegisterWire(cdc)
	ibc.RegisterWire(cdc)
	contrib.RegisterWire(cdc)

	// register custom AppAccount
	cdc.RegisterInterface((*auth.Account)(nil), nil)
	cdc.RegisterConcrete(&auth.BaseAccount{}, "auth/Account", nil)
	cdc.RegisterConcrete(&types.ReputeAccount{}, "forbole/Repute", nil)
	return cdc
}

// Custom logic for forbole initialization
func (app *ForboleApp) initChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	stateJSON := req.AppStateBytes

	var genesisState GenesisState
	err := app.cdc.UnmarshalJSON(stateJSON, &genesisState)
	if err != nil {
		panic(err) // TODO https://github.com/cosmos/cosmos-sdk/issues/468
		// return sdk.ErrGenesisParse("").TraceCause(err, "")
	}

	// load the accounts
	for _, gacc := range genesisState.Accounts {
		acc := gacc.ToAccount()
		app.accountMapper.SetAccount(ctx, acc)
	}

	for _, admin := range genesisState.Admins {
		acc := admin.ToReputeAccount()
		app.reputeAccountMapper.SetAccount(ctx, acc)
	}

	// load the initial stake information
	stake.InitGenesis(ctx, app.stakeKeeper, genesisState.StakeData)

	return abci.ResponseInitChain{}
}

// Custom logic for state export
func (app *ForboleApp) ExportAppStateJSON() (appState json.RawMessage, err error) {
	ctx := app.NewContext(true, abci.Header{})

	// iterate to get the accounts
	accounts := []GenesisAccount{}
	appendAccount := func(acc auth.Account) (stop bool) {
		account := NewGenesisAccountI(acc)
		accounts = append(accounts, account)
		return false
	}
	app.accountMapper.IterateAccounts(ctx, appendAccount)

	// iterate to get the admins
	admins := []GenesisAdmin{}
	appendAdmin := func(acc auth.Account) (stop bool) {
		role := acc.(*types.ReputeAccount).GetRole()
		if role == "Admin" {
			admin := GenesisAdmin{
				Address: acc.GetAddress(),
				Role:    role,
			}
			admins = append(admins, admin)
		}
		return false
	}
	app.reputeAccountMapper.IterateAccounts(ctx, appendAdmin)

	genState := GenesisState{
		Accounts:  accounts,
		Admins:    admins,
		StakeData: stake.WriteGenesis(ctx, app.stakeKeeper),
	}
	return wire.MarshalJSONIndent(app.cdc, genState)
}
