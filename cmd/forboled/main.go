package main

import (
	"encoding/json"

	"github.com/spf13/cobra"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/tmlibs/cli"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	"github.com/cosmos/cosmos-sdk/server"
	"github.com/forbole/forboled/app"
)

func main() {
	cdc := app.MakeCodec()
	ctx := server.NewDefaultContext()
	rootCmd := &cobra.Command{
		Use:               "forboled",
		Short:             "Forbole Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}

	server.AddCommands(ctx, cdc, rootCmd, app.ForboleAppInit(),
		server.ConstructAppCreator(newApp, "forbole"),
		server.ConstructAppExporter(exportAppState, "forbole"))

	// prepare and add flags
	executor := cli.PrepareBaseCmd(rootCmd, "FB", app.DefaultNodeHome)
	executor.Execute()
}

func newApp(logger log.Logger, db dbm.DB) abci.Application {
	return app.NewForboleApp(logger, db)
}

func exportAppState(logger log.Logger, db dbm.DB) (json.RawMessage, error) {
	fapp := app.NewForboleApp(logger, db)
	return fapp.ExportAppStateJSON()
}
