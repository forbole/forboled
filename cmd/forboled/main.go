package main

import (
	"encoding/json"
	"io"

	"github.com/forbole/cosmos-sdk/baseapp"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/cli"
	dbm "github.com/tendermint/tendermint/libs/db"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/cosmos/cosmos-sdk/server"
	version2 "github.com/cosmos/cosmos-sdk/version"
	"github.com/forbole/forboled/app"
	version "github.com/forbole/forboled/version"
)

func main() {
	cdc := app.MakeCodec()
	ctx := server.NewDefaultContext()
	cobra.EnableCommandSorting = false
	rootCmd := &cobra.Command{
		Use:               "forboled",
		Short:             "Forbole Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}

	server.AddCommands(ctx, cdc, rootCmd, app.ForboleAppInit(),
		server.ConstructAppCreator(newApp, "forbole"),
		server.ConstructAppExporter(exportAppStateAndTMValidators, "forbole"))

	rootCmd.RemoveCommand(version2.VersionCmd)
	rootCmd.AddCommand(version.VersionCmd)
	// prepare and add flags
	executor := cli.PrepareBaseCmd(rootCmd, "FB", app.DefaultNodeHome)
	err := executor.Execute()
	if err != nil {
		// handle with #870
		panic(err)
	}
}

func newApp(logger log.Logger, db dbm.DB, traceStore io.Writer) abci.Application {
	return app.NewForboleApp(logger, db, traceStore, baseapp.SetPruning(viper.GetString("pruning")))
}

func exportAppStateAndTMValidators(logger log.Logger, db dbm.DB, traceStore io.Writer) (json.RawMessage, []tmtypes.GenesisValidator, error) {
	fApp := app.NewForboleApp(logger, db, traceStore)
	return fApp.ExportAppStateAndValidators()
}
