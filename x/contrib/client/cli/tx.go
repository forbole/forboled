package cli

import (
	"encoding/hex"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/client/utils"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	authctx "github.com/cosmos/cosmos-sdk/x/auth/client/context"

	"github.com/forbole/forboled/types"
	"github.com/forbole/forboled/x/contrib"
	"github.com/forbole/forboled/x/contrib/client"
)

const (
	flagTo      = "to"
	flagKey     = "key"
	flagType    = "type"
	flagContent = "content"
	flagVotes   = "votes"
	flagTime    = "time"
	// flagRole = "role"
	// flagAsync  = "async"
)

// ContribTxCommand will create a contrib tx and sign it with the given key
func ContribTxCmd(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contrib",
		Short: "Create and sign a contrib tx",
		RunE: func(cmd *cobra.Command, args []string) error {
			// ctx := context.NewCoreContextFromViper().WithAccountStore("repute").WithDecoder(types.GetReputeAccountDecoder(cdc))

			// cliCtx := context.NewCLIContext().WithAccountStore("repute").WithAccountDecoder(types.GetReputeAccountDecoder(cdc))
			txCtx := authctx.NewTxContextFromCLI().WithCodec(cdc)
			cliCtx := context.NewCLIContext().
				WithCodec(cdc).
				WithLogger(os.Stdout).
				WithAccountStore("repute").
				WithAccountDecoder(types.GetReputeAccountDecoder(cdc))

			// get the from address
			from, err := cliCtx.GetFromAddress()
			if err != nil {
				return err
			}

			ctbKey, err := hex.DecodeString(viper.GetString(flagKey))
			if err != nil {
				return err
			}

			ctbTime, err := time.Parse(time.RFC3339, viper.GetString(flagTime))
			if err != nil {
				return err
			}

			ctbContent, err := hex.DecodeString(viper.GetString(flagContent))
			if err != nil {
				return err
			}

			// parse destination address
			dest := viper.GetString(flagTo)
			to, err := sdk.AccAddressFromBech32(dest)
			if err != nil {
				return err
			}

			ctbType := viper.GetString(flagType)
			var ctb contrib.Contrib
			switch ctbType {
			case "Invite", "Recommend", "Post":

				switch ctbType {
				case "Invite":
					ctb = contrib.Invite{contrib.BaseContrib2{contrib.BaseContrib{ctbKey, from, ctbTime}, to}, ctbContent}
				case "Post":
					ctb = contrib.Post{contrib.BaseContrib2{contrib.BaseContrib{ctbKey, from, ctbTime}, to}, ctbContent}
				case "Recommend":
					ctb = contrib.Recommend{contrib.BaseContrib2{contrib.BaseContrib{ctbKey, from, ctbTime}, to}, ctbContent}
				}
			case "Vote":

				// get the vote flag to see what kind of vote
				ctbVote := viper.GetString(flagVotes)
				if ctbVote == "Upvote" {
					ctb = contrib.Vote{contrib.BaseContrib3{contrib.BaseContrib{ctbKey, from, ctbTime}, to, int64(1)}, ctbContent}
				} else if ctbVote == "Downvote" {
					ctb = contrib.Vote{contrib.BaseContrib3{contrib.BaseContrib{ctbKey, from, ctbTime}, to, int64(-1)}, ctbContent}
				} else {
					return errors.New("Invalid Vote Type")
				}
			default:
				return errors.New("Invalid Contrib Type")
			}

			// votes := viper.GetString(flagVotes)
			// ctbVotes, err := strconv.ParseInt(votes, 10, 64)
			// if err != nil {
			// 	return err
			// }

			// build and sign the transaction, then broadcast to Tendermint
			msg := client.BuildContribMsg(ctb)

			// Add async tx ??
			// if viper.GetBool(flagAsync) {
			// 	res, err := ctx.EnsureSignBuildBroadcastAsync(ctx.FromAddressName, []sdk.Msg{msg}, cdc)
			// 	if err != nil {
			// 		return err
			// 	}
			// 	fmt.Println("Async contrib tx sent. tx hash: ", res.Hash.String())
			// 	return nil
			// }

			// err = ctx.EnsureSignBuildBroadcast(ctx.FromAddressName, []sdk.Msg{msg}, cdc)
			// if err != nil {
			// 	return err
			// }
			// // fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
			// return nil
			return utils.SendTx(txCtx, cliCtx, []sdk.Msg{msg})
		},
	}

	cmd.Flags().String(flagTo, "", "Address to contrib")
	cmd.Flags().String(flagKey, "", "Key of the contrib")
	cmd.Flags().String(flagType, "", "Type of the contrib")
	cmd.Flags().String(flagContent, "", "Content of the contrib")
	cmd.Flags().String(flagVotes, "", "Votes of the contrib")
	cmd.Flags().String(flagTime, "", "Time of the contrib")
	// cmd.Flags().Bool(flagAsync, false, "Pass the async flag to send a tx without waiting for the tx to be included in a block")
	return cmd
}
