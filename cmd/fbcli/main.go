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
	slashingcmd "github.com/cosmos/cosmos-sdk/x/slashing/client/cli"
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

	// add standard rpc commands
	rpc.AddCommands(rootCmd)
	// rootCmd.AddCommand(client.LineBreak)
	// tx.AddCommands(rootCmd, cdc)
	// rootCmd.AddCommand(client.LineBreak)

	tendermintCmd := &cobra.Command{
		Use:   "tendermint",
		Short: "Tendermint state querying subcommands",
	}
	tendermintCmd.AddCommand(
		rpc.BlockCommand(),
		rpc.ValidatorCommand(),
	)
	tx.AddCommands(tendermintCmd, cdc)

	//Add IBC commands
	ibcCmd := &cobra.Command{
		Use:   "ibc",
		Short: "Inter-Blockchain Communication subcommands",
	}
	ibcCmd.AddCommand(
		client.PostCommands(
			ibccmd.IBCTransferCmd(cdc),
			ibccmd.IBCRelayCmd(cdc),
		)...)

	advancedCmd := &cobra.Command{
		Use:   "advanced",
		Short: "Advanced subcommands",
	}

	advancedCmd.AddCommand(
		tendermintCmd,
		ibcCmd,
		lcd.ServeCommand(cdc),
	)

	rootCmd.AddCommand(
		advancedCmd,
		client.LineBreak,
	)

	// // add query/post commands (custom to binary)
	// rootCmd.AddCommand(
	// 	client.GetCommands(
	// 		authcmd.GetAccountCmd("acc", cdc, authcmd.GetAccountDecoder(cdc)),
	// 		stakecmd.GetCmdQueryValidator("stake", cdc),
	// 		stakecmd.GetCmdQueryValidators("stake", cdc),
	// 		stakecmd.GetCmdQueryDelegation("stake", cdc),
	// 		stakecmd.GetCmdQueryDelegations("stake", cdc),
	// 		ctbcmd.GetReputeCmd("repute", cdc, types.GetReputeAccountDecoder(cdc)),
	// 		ctbcmd.GetContribCmd("contrib", cdc),
	// 	)...)

	//Add stake commands
	stakeCmd := &cobra.Command{
		Use:   "stake",
		Short: "Stake and validation subcommands",
	}
	stakeCmd.AddCommand(
		client.GetCommands(
			stakecmd.GetCmdQueryValidator("stake", cdc),
			stakecmd.GetCmdQueryValidators("stake", cdc),
			stakecmd.GetCmdQueryDelegation("stake", cdc),
			stakecmd.GetCmdQueryDelegations("stake", cdc),
			slashingcmd.GetCmdQuerySigningInfo("slashing", cdc),
		)...)
	stakeCmd.AddCommand(
		client.PostCommands(
			stakecmd.GetCmdCreateValidator(cdc),
			stakecmd.GetCmdEditValidator(cdc),
			stakecmd.GetCmdDelegate(cdc),
			stakecmd.GetCmdUnbond(cdc),
			slashingcmd.GetCmdUnrevoke(cdc),
		)...)
	rootCmd.AddCommand(
		stakeCmd,
	)

	//Add auth and bank commands
	rootCmd.AddCommand(
		client.GetCommands(
			authcmd.GetAccountCmd("acc", cdc, authcmd.GetAccountDecoder(cdc)),
			ctbcmd.GetReputeCmd("repute", cdc, types.GetReputeAccountDecoder(cdc)),
			ctbcmd.GetContribCmd("contrib", cdc),
		)...)
	rootCmd.AddCommand(
		client.PostCommands(
			bankcmd.SendTxCmd(cdc),
			ctbcmd.ContribTxCmd(cdc),
		)...)

	// rootCmd.AddCommand(
	// 	client.PostCommands(
	// 		bankcmd.SendTxCmd(cdc),
	// 		ibccmd.IBCTransferCmd(cdc),
	// 		ibccmd.IBCRelayCmd(cdc),
	// 		stakecmd.GetCmdCreateValidator(cdc),
	// 		stakecmd.GetCmdEditValidator(cdc),
	// 		stakecmd.GetCmdDelegate(cdc),
	// 		stakecmd.GetCmdUnbond(cdc),
	// 		ctbcmd.ContribTxCmd(cdc),
	// 	)...)

	// add proxy, version and key info
	rootCmd.AddCommand(
		// client.LineBreak,
		// lcd.ServeCommand(cdc),
		keys.Commands(),
		client.LineBreak,
		version.VersionCmd,
	)

	// prepare and add flags
	executor := cli.PrepareMainCmd(rootCmd, "FB", app.DefaultCLIHome)
	executor.Execute()
}
