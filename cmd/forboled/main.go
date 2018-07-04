package main

import (
	"encoding/json"

	"github.com/spf13/cobra"

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

func newApp(logger log.Logger, db dbm.DB) abci.Application {
	return app.NewForboleApp(logger, db)
}

func exportAppStateAndTMValidators(logger log.Logger, db dbm.DB) (json.RawMessage, []tmtypes.GenesisValidator, error) {
	fapp := app.NewForboleApp(logger, db)
	return fapp.ExportAppStateAndValidators()
}
