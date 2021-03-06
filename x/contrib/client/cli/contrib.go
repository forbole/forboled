package cli

import (
	"encoding/hex"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/wire"

	"github.com/forbole/forboled/x/contrib"
)

const (
	flagScore = "score"
)

// GetContribCmd returns a query contrib that will display the
// state of the contrib at a given key
func GetContribCmd(storeName string, cdc *wire.Codec) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query [key]",
		Short: "Query contrib status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// find the key to look up the contrib
			key, err := hex.DecodeString(args[0])
			if err != nil {
				return err
			}

			// perform query
			cliCtx := context.NewCLIContext().WithCodec(cdc)
			res, err := cliCtx.QueryStore(key, storeName)
			// res, err := ctx.Query(storeName)
			if err != nil {
				return err
			}

			// parse out the value
			var ctb contrib.Status
			err = cdc.UnmarshalBinaryBare(res, &ctb)
			if err != nil {
				fmt.Println(res)
				return err
			}

			// print out whole contrib
			output, err := json.MarshalIndent(ctb, "", "  ")
			if err != nil {
				return err
			}
			if viper.GetBool(flagScore) {
				fmt.Println(ctb.GetScore())
			} else {
				fmt.Println(string(output))
			}
			return nil
		},
	}
	cmd.Flags().Bool(flagScore, false, "bool of only showing score")
	return cmd
}
