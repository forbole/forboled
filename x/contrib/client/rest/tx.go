package rest

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/tendermint/go-crypto/keys"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/wire"

	"github.com/forbole/forboled/x/contrib"
	"github.com/forbole/forboled/x/contrib/client"
)

// RegisterRoutes - Central function to define routes that get registered by the main application
func RegisterRoutes(ctx context.CoreContext, r *mux.Router, cdc *wire.Codec, kb keys.Keybase) {
	r.HandleFunc("/accounts/contrib", ContribRequestHandlerFn(cdc, kb, ctx)).Methods("POST")
}

type contribBody struct {
	// fees is not used currently
	// Fees             sdk.Coin  `json="fees"`
	Contrib          contrib.Contrib `json:"contrib"`
	LocalAccountName string          `json:"name"`
	Password         string          `json:"password"`
	ChainID          string          `json:"chain_id"`
	Sequence         int64           `json:"sequence"`
}

// ContribRequestHandlerFn - http request handler to send contrib
func ContribRequestHandlerFn(cdc *wire.Codec, kb keys.Keybase, ctx context.CoreContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// collect data
		// vars := mux.Vars(r)
		// address := vars["address"]

		var m contribBody
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		err = json.Unmarshal(body, &m)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		// info, err := kb.Get(m.LocalAccountName)
		// if err != nil {
		// 	w.WriteHeader(http.StatusUnauthorized)
		// 	w.Write([]byte(err.Error()))
		// 	return
		// }

		// bz, err := hex.DecodeString(address)
		// if err != nil {
		// 	w.WriteHeader(http.StatusBadRequest)
		// 	w.Write([]byte(err.Error()))
		// 	return
		// }
		// to := sdk.Address(bz)

		// build message
		msg := client.BuildContribMsg(m.Contrib)
		if err != nil { // XXX rechecking same error ?
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		// sign
		ctx = ctx.WithSequence(m.Sequence)
		txBytes, err := ctx.SignAndBuild(m.LocalAccountName, m.Password, msg, cdc)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		}

		// send
		res, err := ctx.BroadcastTx(txBytes)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		output, err := json.MarshalIndent(res, "", "  ")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		w.Write(output)
	}
}
