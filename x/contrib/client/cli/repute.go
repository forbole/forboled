package cli

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
)

// GetReputeCmd returns a query account that will display the
// state of the account at a given address
func GetReputeCmd(storeName string, cdc *wire.Codec, decoder auth.AccountDecoder) *cobra.Command {
	cmdr := commander{
		storeName,
		cdc,
		decoder,
	}
	return &cobra.Command{
		Use:   "repute <address>",
		Short: "Query account repute",
		RunE:  cmdr.getReputeCmd,
	}
}

type commander struct {
	storeName string
	cdc       *wire.Codec
	decoder   auth.AccountDecoder
}

func (c commander) getReputeCmd(cmd *cobra.Command, args []string) error {
	if len(args) != 1 || len(args[0]) == 0 {
		return errors.New("You must provide an account name")
	}

	// find the key to look up the account
	addr := args[0]
	key, err := sdk.GetAccAddressBech32(addr)
	if err != nil {
		return err
	}

	ctx := context.NewCoreContextFromViper()

	res, err := ctx.Query(auth.AddressStoreKey(key), c.storeName)
	if err != nil {
		return err
	}

	// decode the value
	account, err := c.decoder(res)
	if err != nil {
		return err
	}

	// print out whole account
	output, err := json.MarshalIndent(account, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(output))

	return nil
}
