package main

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/crypto"
	cmn "github.com/tendermint/tendermint/libs/common"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"

	bam "github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/bank"
	"github.com/cosmos/cosmos-sdk/x/ibc"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	"github.com/cosmos/cosmos-sdk/x/stake"

	"github.com/forbole/forboled/types"
	"github.com/forbole/forboled/x/contrib"

	forbole "github.com/forbole/forboled/app"
)

func runHackCmd(cmd *cobra.Command, args []string) error {

	if len(args) != 1 {
		return fmt.Errorf("Expected 1 arg")
	}

	// ".forboled"
	dataDir := args[0]
	dataDir = path.Join(dataDir, "data")

	// load the app
	logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
	db, err := dbm.NewGoLevelDB("forbole", dataDir)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	app := NewForboleApp(logger, db)

	// print some info
	id := app.LastCommitID()
	lastBlockHeight := app.LastBlockHeight()
	fmt.Println("ID", id)
	fmt.Println("LastBlockHeight", lastBlockHeight)

	//----------------------------------------------------
	// XXX: start hacking!
	//----------------------------------------------------
	// eg. gaia-6001 testnet bug
	// We paniced when iterating through the "bypower" keys.
	// The following powerKey was there, but the corresponding "trouble" validator did not exist.
	// So here we do a binary search on the past states to find when the powerKey first showed up ...

	// owner of the validator the bonds, gets revoked, later unbonds, and then later is still found in the bypower store
	trouble := hexToBytes("D3DC0FF59F7C3B548B7AFA365561B87FD0208AF8")
	// this is his "bypower" key
	powerKey := hexToBytes("05303030303030303030303033FFFFFFFFFFFF4C0C0000FFFED3DC0FF59F7C3B548B7AFA365561B87FD0208AF8")

	topHeight := lastBlockHeight
	bottomHeight := int64(0)
	checkHeight := topHeight
	for {
		// load the given version of the state
		err = app.LoadVersion(checkHeight, app.keyMain)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		ctx := app.NewContext(true, abci.Header{})

		// check for the powerkey and the validator from the store
		store := ctx.KVStore(app.keyStake)
		res := store.Get(powerKey)
		val, _ := app.stakeKeeper.GetValidator(ctx, trouble)
		fmt.Println("checking height", checkHeight, res, val)
		if res == nil {
			bottomHeight = checkHeight
		} else {
			topHeight = checkHeight
		}
		checkHeight = (topHeight + bottomHeight) / 2
	}
}

func base64ToPub(b64 string) crypto.PubKeyEd25519 {
	data, _ := base64.StdEncoding.DecodeString(b64)
	var pubKey crypto.PubKeyEd25519
	copy(pubKey[:], data)
	return pubKey

}

func hexToBytes(h string) []byte {
	trouble, _ := hex.DecodeString(h)
	return trouble

}

//--------------------------------------------------------------------------------
// NOTE: This is all copied from gaia/app/app.go
// so we can access internal fields!

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
	keyMain     *sdk.KVStoreKey
	keyAccount  *sdk.KVStoreKey
	keyIBC      *sdk.KVStoreKey
	keyStake    *sdk.KVStoreKey
	keySlashing *sdk.KVStoreKey
	keyContrib  *sdk.KVStoreKey
	// keyGov      *sdk.KVStoreKey
	keyRepute *sdk.KVStoreKey

	// Manage getting and setting accounts
	accountMapper       auth.AccountMapper
	feeCollectionKeeper auth.FeeCollectionKeeper
	reputeAccountMapper auth.AccountMapper
	coinKeeper          bank.Keeper
	ibcMapper           ibc.Mapper
	stakeKeeper         stake.Keeper
	slashingKeeper      slashing.Keeper
	// govKeeper           gov.Keeper
	contribKeeper contrib.Keeper
}

func NewForboleApp(logger log.Logger, db dbm.DB) *ForboleApp {

	// Create app-level codec for txs and accounts.
	var cdc = MakeCodec()

	// Create your application object.
	var app = &ForboleApp{
		BaseApp:     bam.NewBaseApp(appName, cdc, logger, db),
		cdc:         cdc,
		keyMain:     sdk.NewKVStoreKey("main"),
		keyAccount:  sdk.NewKVStoreKey("acc"),
		keyIBC:      sdk.NewKVStoreKey("ibc"),
		keyStake:    sdk.NewKVStoreKey("stake"),
		keySlashing: sdk.NewKVStoreKey("slashing"),
		keyContrib:  sdk.NewKVStoreKey("contrib"),
		// keyGov:      sdk.NewKVStoreKey("gov"),
		keyRepute: sdk.NewKVStoreKey("repute"),
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
	app.slashingKeeper = slashing.NewKeeper(app.cdc, app.keySlashing, app.stakeKeeper, app.RegisterCodespace(slashing.DefaultCodespace))
	// app.govKeeper = gov.NewKeeper(app.cdc, app.keyGov, app.coinKeeper, app.stakeKeeper, app.RegisterCodespace(gov.DefaultCodespace))
	app.contribKeeper = contrib.NewKeeper(app.cdc, app.reputeAccountMapper, app.keyContrib)
	// app.feeCollectionKeeper = auth.NewFeeCollectionKeeper(app.cdc, app.keyFeeCollection)
	app.Router().
		// AddRoute("auth", auth.NewHandler(app.accountMapper)).
		AddRoute("bank", bank.NewHandler(app.coinKeeper)).
		AddRoute("ibc", ibc.NewHandler(app.ibcMapper, app.coinKeeper)).
		AddRoute("stake", stake.NewHandler(app.stakeKeeper)).
		// AddRoute("gov", gov.NewHandler(app.govKeeper)).
		AddRoute("contrib", contrib.NewHandler(app.contribKeeper)).
		AddRoute("slashing", slashing.NewHandler(app.slashingKeeper))

	// Initialize BaseApp.
	app.SetInitChainer(app.initChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)
	app.SetAnteHandler(auth.NewAnteHandler(app.accountMapper, app.feeCollectionKeeper))
	app.SetAnteHandler(auth.NewAnteHandler(app.reputeAccountMapper, app.feeCollectionKeeper))
	// app.MountStoresIAVL(app.keyMain, app.keyAccount, app.keyIBC, app.keyStake, app.keySlashing, app.keyGov, app.keyFeeCollection, app.keyContrib, app.keyRepute)
	app.MountStoresIAVL(app.keyMain, app.keyAccount, app.keyIBC, app.keyStake, app.keySlashing, app.keyContrib, app.keyRepute)
	// app.SetAnteHandlers(auth.NewAnteHandler(app.accountMapper, app.feeCollectionKeeper), auth.NewAnteHandler(app.reputeAccountMapper, app.feeCollectionKeeper))
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
	slashing.RegisterWire(cdc)
	// gov.RegisterWire(cdc)   //later used maybe
	auth.RegisterWire(cdc)
	// auth.RegisterWire(cdc) //?needed?
	ibc.RegisterWire(cdc)
	contrib.RegisterWire(cdc)

	// register custom AppAccount
	cdc.RegisterInterface((*auth.Account)(nil), nil)
	cdc.RegisterConcrete(&auth.BaseAccount{}, "auth/Account", nil)
	cdc.RegisterConcrete(&types.ReputeAccount{}, "forbole/Repute", nil)
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

	return abci.ResponseEndBlock{
		ValidatorUpdates: validatorUpdates,
	}
}

// Custom logic for forbole initialization
func (app *ForboleApp) initChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	stateJSON := req.AppStateBytes

	var genesisState forbole.GenesisState
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
	stake.InitGenesis(ctx, app.stakeKeeper, genesisState.StakeData)

	// gov.InitGenesis(ctx, app.govKeeper, gov.DefaultGenesisState())

	return abci.ResponseInitChain{}
}

// // export the state of gaia for a genesis file
// func (app *ForboleApp) ExportAppStateAndValidators() (appState json.RawMessage, validators []tmtypes.GenesisValidator, err error) {
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
// 	appState, err = wire.MarshalJSONIndent(app.cdc, genState)
// 	if err != nil {
// 		return nil, nil, err
// 	}
// 	validators = stake.WriteValidators(ctx, app.stakeKeeper)
// 	return appState, validators, nil
// }
