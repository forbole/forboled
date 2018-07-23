package app

import (
	"encoding/json"
	"io"
	"os"

	abci "github.com/tendermint/tendermint/abci/types"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"

	// bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/gov"
	"github.com/cosmos/cosmos-sdk/x/ibc"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/stake"
	bam "github.com/forbole/cosmos-sdk/baseapp"

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
	keyMain          *sdk.KVStoreKey
	keyAccount       *sdk.KVStoreKey
	keyIBC           *sdk.KVStoreKey
	keyStake         *sdk.KVStoreKey
	keySlashing      *sdk.KVStoreKey
	keyContrib       *sdk.KVStoreKey
	keyGov           *sdk.KVStoreKey
	keyRepute        *sdk.KVStoreKey
	keyFeeCollection *sdk.KVStoreKey

	// Manage getting and setting accounts
	accountMapper       auth.AccountMapper
	feeCollectionKeeper auth.FeeCollectionKeeper
	reputeAccountMapper auth.AccountMapper
	coinKeeper          bank.Keeper
	ibcMapper           ibc.Mapper
	stakeKeeper         stake.Keeper
	slashingKeeper      slashing.Keeper
	govKeeper           gov.Keeper
	contribKeeper       contrib.Keeper
}

func NewForboleApp(logger log.Logger, db dbm.DB, traceStore io.Writer, baseAppOptions ...func(*bam.BaseApp)) *ForboleApp {

	// Create app-level codec for txs and accounts.
	var cdc = MakeCodec()

	bApp := bam.NewBaseApp(appName, cdc, logger, db, baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)

	var app = &ForboleApp{
		BaseApp:          bApp,
		cdc:              cdc,
		keyMain:          sdk.NewKVStoreKey("main"),
		keyAccount:       sdk.NewKVStoreKey("acc"),
		keyIBC:           sdk.NewKVStoreKey("ibc"),
		keyStake:         sdk.NewKVStoreKey("stake"),
		keySlashing:      sdk.NewKVStoreKey("slashing"),
		keyContrib:       sdk.NewKVStoreKey("contrib"),
		keyGov:           sdk.NewKVStoreKey("gov"),
		keyFeeCollection: sdk.NewKVStoreKey("fee"),
		keyRepute:        sdk.NewKVStoreKey("repute"),
	}

	// Define the accountMapper.
	app.accountMapper = auth.NewAccountMapper(
		app.cdc,
		app.keyAccount,        // target store
		auth.ProtoBaseAccount, // prototype
	)
	app.reputeAccountMapper = auth.NewAccountMapper(
		app.cdc,
		app.keyRepute,            // target store
		types.ProtoReputeAccount, // prototype //have to change here as well ????????????
	)

	// Add handlers.
	app.coinKeeper = bank.NewKeeper(app.accountMapper)
	app.ibcMapper = ibc.NewMapper(app.cdc, app.keyIBC, app.RegisterCodespace(ibc.DefaultCodespace))
	app.stakeKeeper = stake.NewKeeper(app.cdc, app.keyStake, app.coinKeeper, app.RegisterCodespace(stake.DefaultCodespace))
	app.slashingKeeper = slashing.NewKeeper(app.cdc, app.keySlashing, app.stakeKeeper, app.RegisterCodespace(slashing.DefaultCodespace))
	app.govKeeper = gov.NewKeeper(app.cdc, app.keyGov, app.coinKeeper, app.stakeKeeper, app.RegisterCodespace(gov.DefaultCodespace))
	app.feeCollectionKeeper = auth.NewFeeCollectionKeeper(app.cdc, app.keyFeeCollection)
	app.contribKeeper = contrib.NewKeeper(app.cdc, app.reputeAccountMapper, app.keyContrib)
	app.Router().
		// AddRoute("auth", auth.NewHandler(app.accountMapper)).
		AddRoute("bank", bank.NewHandler(app.coinKeeper)).
		AddRoute("ibc", ibc.NewHandler(app.ibcMapper, app.coinKeeper)).
		AddRoute("stake", stake.NewHandler(app.stakeKeeper)).
		AddRoute("contrib", contrib.NewHandler(app.contribKeeper)).
		AddRoute("slashing", slashing.NewHandler(app.slashingKeeper)).
		AddRoute("gov", gov.NewHandler(app.govKeeper))

	// Initialize BaseApp.
	app.SetInitChainer(app.initChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)
	// app.SetAnteHandler(auth.NewAnteHandler(app.accountMapper, app.feeCollectionKeeper))
	// app.SetAnteHandler(auth.NewAnteHandler(app.reputeAccountMapper, app.feeCollectionKeeper))
	app.SetAnteHandlers(auth.NewAnteHandler(app.accountMapper, app.feeCollectionKeeper), auth.NewAnteHandler(app.reputeAccountMapper, app.feeCollectionKeeper))
	app.MountStoresIAVL(app.keyMain, app.keyAccount, app.keyIBC, app.keyStake, app.keySlashing, app.keyGov, app.keyFeeCollection, app.keyContrib, app.keyRepute)

	// set AnteHandler
	// var ahs [2]sdk.AnteHandler
	// ahs[0] = auth.NewAnteHandler(app.accountMapper, app.feeCollectionKeeper)
	// ahs[1] = auth.NewAnteHandler(app.reputeAccountMapper, app.feeCollectionKeeper)
	// app.anteHandler = ahs

	err := app.LoadLatestVersion(app.keyMain)
	if err != nil {
		cmn.Exit(err.Error())
	}

	return app
}

// Custom tx codec
func MakeCodec() *wire.Codec {
	var cdc = wire.NewCodec()
	ibc.RegisterWire(cdc)
	bank.RegisterWire(cdc)
	stake.RegisterWire(cdc)
	slashing.RegisterWire(cdc)
	gov.RegisterWire(cdc)
	sdk.RegisterWire(cdc)
	wire.RegisterCrypto(cdc)
	contrib.RegisterWire(cdc)
	// auth.RegisterWire(cdc) //?needed?

	// register custom AppAccount
	cdc.RegisterInterface((*auth.Account)(nil), nil)
	cdc.RegisterConcrete(&auth.BaseAccount{}, "auth/Account", nil)
	cdc.RegisterConcrete(&types.ReputeAccount{}, "forbole/Repute", nil)
	cdc.Seal()
	return cdc
}

// application updates every end block
func (app *ForboleApp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	tags := slashing.BeginBlocker(ctx, req, app.slashingKeeper)

	return abci.ResponseBeginBlock{
		Tags: tags.ToKVPairs(),
	}
}

// application updates every end block
func (app *ForboleApp) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	validatorUpdates := stake.EndBlocker(ctx, app.stakeKeeper)

	tags, _ := gov.EndBlocker(ctx, app.govKeeper)

	return abci.ResponseEndBlock{
		ValidatorUpdates: validatorUpdates,
		Tags:             tags,
	}
}

// Custom logic for forbole initialization
func (app *ForboleApp) initChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	stateJSON := req.AppStateBytes

	var genesisState GenesisState
	// genesisState := new(GenesisState)
	err := app.cdc.UnmarshalJSON(stateJSON, &genesisState)
	if err != nil {
		panic(err) // TODO https://github.com/cosmos/cosmos-sdk/issues/468
		// return sdk.ErrGenesisParse("").TraceCause(err, "")
	}

	// load the accounts
	for _, facc := range genesisState.Accounts {
		acc := facc.ToAccount()
		acc.AccountNumber = app.accountMapper.GetNextAccountNumber(ctx)
		app.accountMapper.SetAccount(ctx, acc)
	}

	for _, admin := range genesisState.Admins {
		acc := admin.ToReputeAccount()
		acc.AccountNumber = app.reputeAccountMapper.GetNextAccountNumber(ctx)
		app.reputeAccountMapper.SetAccount(ctx, acc)
	}

	// load the initial stake information
	err = stake.InitGenesis(ctx, app.stakeKeeper, genesisState.StakeData)
	if err != nil {
		panic(err) // TODO https://github.com/cosmos/cosmos-sdk/issues/468
		// return sdk.ErrGenesisParse("").TraceCause(err, "")
	}

	gov.InitGenesis(ctx, app.govKeeper, gov.DefaultGenesisState())

	return abci.ResponseInitChain{}
}

// // Custom logic for state export
// func (app *ForboleApp) ExportAppStateJSON() (appState json.RawMessage, err error) {
// 	ctx := app.NewContext(true, abci.Header{})

// 	// iterate to get the accounts
// 	accounts := []GenesisAccount{}
// 	appendAccount := func(acc auth.Account) (stop bool) {
// 		account := NewGenesisAccountI(acc)
// 		accounts = append(accounts, account)
// 		return false
// 	}
// 	app.accountMapper.IterateAccounts(ctx, appendAccount)

// 	// iterate to get the admins
// 	admins := []GenesisAdmin{}
// 	appendAdmin := func(acc auth.Account) (stop bool) {
// 		role := acc.(*types.ReputeAccount).GetRole()
// 		if role == "Admin" {
// 			admin := GenesisAdmin{
// 				Address: acc.GetAddress(),
// 				Role:    role,
// 			}
// 			admins = append(admins, admin)
// 		}
// 		return false
// 	}
// 	app.reputeAccountMapper.IterateAccounts(ctx, appendAdmin)

// 	genState := GenesisState{
// 		Accounts:  accounts,
// 		Admins:    admins,
// 		StakeData: stake.WriteGenesis(ctx, app.stakeKeeper),
// 	}
// 	return wire.MarshalJSONIndent(app.cdc, genState)
// }

// export the state of gaia for a genesis file
func (app *ForboleApp) ExportAppStateAndValidators() (appState json.RawMessage, validators []tmtypes.GenesisValidator, err error) {
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
	appState, err = wire.MarshalJSONIndent(app.cdc, genState)
	if err != nil {
		return nil, nil, err
	}
	validators = stake.WriteValidators(ctx, app.stakeKeeper)
	return appState, validators, nil
}
