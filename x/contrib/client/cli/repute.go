package cli

import (
	"encoding/json"
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/context"
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
	key, err := sdk.AccAddressFromBech32(addr)
	if err != nil {
		return err
	}

	ctx := context.NewCoreContextFromViper()

	res, err := ctx.QueryStore(auth.AddressStoreKey(key), c.storeName)
	// res, err := ctx.Query(c.storeName)
	if err != nil {
		return sdk.ErrUnknownAddress("No repute account with address " + addr +
			" was found in the state.\nAre you sure there has been a transaction involving it?")
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
