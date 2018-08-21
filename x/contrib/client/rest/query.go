package rest

import (
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	"github.com/cosmos/cosmos-sdk/x/auth"
	// "github.com/cosmos/cosmos-sdk/x/stake"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"

	"github.com/forbole/forboled/x/contrib"
	// "github.com/forbole/forboled/x/contrib/client"
)

func registerQueryRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *wire.Codec) {
	r.HandleFunc(
		"/contrib/{key}",
		contribHandlerFn(cliCtx, "contrib", cdc),
	).Methods("GET")
	r.HandleFunc(
		"/reputeaccount/{address}",
		reputeAccountHandlerFn(cliCtx, "repute", authcmd.GetAccountDecoder(cdc), cdc),
	).Methods("GET")
	r.HandleFunc(
		"/contrib/{key}/score",
		contribScoreHandlerFn(cliCtx, "contrib", cdc),
	).Methods("GET")
}

// http request handler to query delegator bonding status
func contribHandlerFn(cliCtx context.CLIContext, storeName string, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// read parameters
		vars := mux.Vars(r)
		// address := vars["address"]
		k := vars["key"]

		// bz, err := sdk.AccAddressFromBech32(address)
		// if err != nil {
		// 	w.WriteHeader(http.StatusBadRequest)
		// 	w.Write([]byte(err.Error()))
		// 	return
		// }
		// ctbAddr := sdk.Address(bz)

		// decode the key to look up the contrib

		key, err := hex.DecodeString(k)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		res, err := cliCtx.QueryStore(key, storeName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Couldn't query contribution. Error: %s", err.Error())))
			return
		}

		// the query will return empty if there is no data for this bond
		if len(res) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		var ctb contrib.Status
		err = cdc.UnmarshalBinaryBare(res, &ctb)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Couldn't decode contribution. Error: %s", err.Error())))
			return
		}

		output, err := cdc.MarshalJSON(ctb)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(output)
	}
}

func reputeAccountHandlerFn(cliCtx context.CLIContext, storeName string, decoder auth.AccountDecoder, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		addr := vars["address"]

		key, err := sdk.AccAddressFromBech32(addr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		res, err := cliCtx.QueryStore(auth.AddressStoreKey(key), storeName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Could't query account. Error: %s", err.Error())))
			return
		}

		// the query will return empty if there is no data for this account
		if len(res) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// decode the value
		account, err := decoder(res)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Could't parse query result. Result: %s. Error: %s", res, err.Error())))
			return
		}

		// print out whole account
		output, err := cdc.MarshalJSON(account)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Could't marshall query result. Error: %s", err.Error())))
			return
		}

		w.Write(output)
	}
}

func contribScoreHandlerFn(cliCtx context.CLIContext, storeName string, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// read parameters
		vars := mux.Vars(r)
		// address := vars["address"]
		k := vars["key"]

		// bz, err := sdk.AccAddressFromBech32(address)
		// if err != nil {
		// 	w.WriteHeader(http.StatusBadRequest)
		// 	w.Write([]byte(err.Error()))
		// 	return
		// }
		// ctbAddr := sdk.Address(bz)

		// decode the key to look up the contrib

		key, err := hex.DecodeString(k)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		res, err := cliCtx.QueryStore(key, storeName)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Couldn't query contribution. Error: %s", err.Error())))
			return
		}

		// the query will return empty if there is no data for this bond
		if len(res) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		var ctb contrib.Status
		err = cdc.UnmarshalBinaryBare(res, &ctb)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Couldn't decode contribution. Error: %s", err.Error())))
			return
		}

		output, err := cdc.MarshalJSON(ctb.GetScore())
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("Could't marshall query result. Error: %s", err.Error())))
			return
		}

		w.Write(output)
	}
}
