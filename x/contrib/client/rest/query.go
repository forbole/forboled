package rest

import (
	"encoding/hex"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"

	"github.com/cosmos/cosmos-sdk/client/context"
	// sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/wire"
	// "github.com/cosmos/cosmos-sdk/x/stake"

	"github.com/forbole/forboled/x/contrib"
	// "github.com/forbole/forboled/x/contrib/client"
)

func registerQueryRoutes(ctx context.CoreContext, r *mux.Router, cdc *wire.Codec) {
	r.HandleFunc(
		"/contrib/{key}",
		contribHandlerFn(ctx, "contrib", cdc),
	).Methods("GET")
}

// http request handler to query delegator bonding status
func contribHandlerFn(ctx context.CoreContext, storeName string, cdc *wire.Codec) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		// read parameters
		vars := mux.Vars(r)
		// address := vars["address"]
		k := vars["key"]

		// bz, err := sdk.GetAccAddressBech32(address)
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

		res, err := ctx.Query(key, storeName)
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
