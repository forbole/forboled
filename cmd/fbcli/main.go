package main

import (
	"github.com/spf13/cobra"

	"github.com/tendermint/tmlibs/cli"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/tx"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	bankcmd "github.com/cosmos/cosmos-sdk/x/bank/client/cli"
	ibccmd "github.com/cosmos/cosmos-sdk/x/ibc/client/cli"
	stakecmd "github.com/cosmos/cosmos-sdk/x/stake/client/cli"
	"github.com/forbole/forboled/client/lcd"
	"github.com/forbole/forboled/version"

	"github.com/forbole/forboled/app"
	"github.com/forbole/forboled/types"
	ctbcmd "github.com/forbole/forboled/x/contrib/client/cli"
)

// rootCmd is the entry point for this binary
var (
	rootCmd = &cobra.Command{
		Use:   "fbcli",
		Short: "Forbole light-client",
	}
)

func main() {
	// disable sorting
	cobra.EnableCommandSorting = false

	// get the codec
	cdc := app.MakeCodec()

	// TODO: setup keybase, viper object, etc. to be passed into
	// the below functions and eliminate global vars, like we do
	// with the cdc

	// add standard rpc, and tx commands
	rpc.AddCommands(rootCmd)
	rootCmd.AddCommand(client.LineBreak)
	tx.AddCommands(rootCmd, cdc)
	rootCmd.AddCommand(client.LineBreak)

	// add query/post commands (custom to binary)
	rootCmd.AddCommand(
		client.GetCommands(
			authcmd.GetAccountCmd("acc", cdc, authcmd.GetAccountDecoder(cdc)),
			stakecmd.GetCmdQueryValidator("stake", cdc),
			stakecmd.GetCmdQueryValidators("stake", cdc),
			stakecmd.GetCmdQueryDelegation("stake", cdc),
			stakecmd.GetCmdQueryDelegations("stake", cdc),
			ctbcmd.GetReputeCmd("repute", cdc, types.GetReputeAccountDecoder(cdc)),
			ctbcmd.GetContribCmd("contrib", cdc),
		)...)

	rootCmd.AddCommand(
		client.PostCommands(
			bankcmd.SendTxCmd(cdc),
			ibccmd.IBCTransferCmd(cdc),
			ibccmd.IBCRelayCmd(cdc),
			// stakecmd.GetCmdDeclareCandidacy(cdc),
			// stakecmd.GetCmdEditCandidacy(cdc),
			stakecmd.GetCmdCreateValidator(cdc),
			stakecmd.GetCmdEditValidator(cdc),
			stakecmd.GetCmdDelegate(cdc),
			stakecmd.GetCmdUnbond(cdc),
			// ctbcmd.ContribTxCmd(cdc),
		)...)

	// add proxy, version and key info
	rootCmd.AddCommand(
		client.LineBreak,
		lcd.ServeCommand(cdc),
		keys.Commands(),
		client.LineBreak,
		version.VersionCmd,
	)

	// prepare and add flags
	executor := cli.PrepareMainCmd(rootCmd, "FB", app.DefaultCLIHome)
	executor.Execute()
}
