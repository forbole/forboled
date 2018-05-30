package cli

import (
	"encoding/hex"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"

	"github.com/forbole/forboled/types"
	"github.com/forbole/forboled/x/contrib"
	"github.com/forbole/forboled/x/contrib/client"
)

const (
	flagTo      = "to"
	flagKey     = "key"
	flagType    = "type"
	flagContent = "content"
	// flagVotes = "votes"
	flagTime = "time"
	// flagRole = "role"
)

// ContribTxCommand will create a contrib tx and sign it with the given key
func ContribTxCmd(cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "contrib",
		Short: "Create and sign a contrib tx",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.NewCoreContextFromViper().WithAccountStore("repute").WithDecoder(types.GetReputeAccountDecoder(cdc))

			// get the from address
			from, err := ctx.GetFromAddress()
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

			ctbType := viper.GetString(flagType)
			var ctb contrib.Contrib
			switch ctbType {
			case "Invite", "Recommend":
				ctbContent, err := hex.DecodeString(viper.GetString(flagContent))
				if err != nil {
					return err
				}

				// parse destination address
				dest := viper.GetString(flagTo)
				bz, err := hex.DecodeString(dest)
				if err != nil {
					return err
				}
				to := sdk.Address(bz)

				switch ctbType {
				case "Invite":
					ctb = contrib.Invite{contrib.BaseContrib2{contrib.BaseContrib{ctbKey, from, ctbTime}, to}, ctbContent}
				case "Recommend":
					ctb = contrib.Recommend{contrib.BaseContrib2{contrib.BaseContrib{ctbKey, from, ctbTime}, to}, ctbContent}
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
			res, err := ctx.EnsureSignBuildBroadcast(ctx.FromAddressName, msg, cdc)
			if err != nil {
				return err
			}
			fmt.Printf("Committed at block %d. Hash: %s\n", res.Height, res.Hash.String())
			return nil
		},
	}

	cmd.Flags().String(flagTo, "", "Address to contrib")
	cmd.Flags().String(flagKey, "", "Key of the contrib")
	cmd.Flags().String(flagType, "", "Type of the contrib")
	cmd.Flags().String(flagContent, "", "Content of the contrib")
	// cmd.Flags().String(flagVotes, "", "Votes of the contrib")
	cmd.Flags().String(flagTime, "", "Time of the contrib")
	return cmd
}
