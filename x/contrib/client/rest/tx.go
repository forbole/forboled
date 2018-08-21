package rest

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gorilla/mux"
	// "github.com/tendermint/go-crypto/keys"
	"github.com/cosmos/cosmos-sdk/crypto/keys"

	"github.com/cosmos/cosmos-sdk/client/context"
	"github.com/cosmos/cosmos-sdk/wire"
	authctx "github.com/cosmos/cosmos-sdk/x/auth/client/context"

	"github.com/forbole/forboled/x/contrib"
	"github.com/forbole/forboled/x/contrib/client"
)

var msgCdc = wire.NewCodec()

func init() {
	contrib.RegisterWire(msgCdc)
}

// RegisterRoutes - Central function to define routes that get registered by the main application
func registerTxRoutes(cliCtx context.CLIContext, r *mux.Router, cdc *wire.Codec, kb keys.Keybase) {
	r.HandleFunc("/contrib/{address}/{ctbtype}", ContribRequestHandlerFn(cdc, kb, cliCtx)).Methods("POST")
}

type contribBody struct {
	// fees is not used currently
	// Fees             sdk.Coin  `json="fees"`
	LocalAccountName string `json:"name"`
	Password         string `json:"password"`
	ChainID          string `json:"chain_id"`
	Sequence         int64  `json:"sequence"`
	AccountNumber    int64  `json:"account_number"`
	Gas              int64  `json:"gas"`
	Content          string `json:"content"`
	Key              string `json:"key"`
	Time             string `json:"time"`
	VoteType         string `json:"votetype"` // must provide if doing vote contrib. can ignore it if not vote
}

// ContribRequestHandlerFn - http request handler to send contrib.
func ContribRequestHandlerFn(cdc *wire.Codec, kb keys.Keybase, cliCtx context.CLIContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// collect data
		vars := mux.Vars(r)
		bech32addr := vars["address"]
		ctbtype := vars["ctbtype"]

		to, err := sdk.AccAddressFromBech32(bech32addr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		var m contribBody
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}
		err = msgCdc.UnmarshalJSON(body, &m)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		info, err := kb.Get(m.LocalAccountName)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		}

		// to, err := sdk.AccAddressFromHex(address.String())
		// if err != nil {
		// 	w.WriteHeader(http.StatusBadRequest)
		// 	w.Write([]byte(err.Error()))
		// 	return
		// }

		ctbTime, err := time.Parse(time.RFC3339, m.Time)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		ctbContent, err := hex.DecodeString(m.Content)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		key, err := hex.DecodeString(m.Key)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		var ctb contrib.Contrib
		switch ctbtype {
		case "invite":
			ctb = contrib.Invite{contrib.BaseContrib2{contrib.BaseContrib{key, sdk.AccAddress(info.GetPubKey().Address()), ctbTime}, to}, ctbContent}
		case "recommend":
			ctb = contrib.Recommend{contrib.BaseContrib2{contrib.BaseContrib{key, sdk.AccAddress(info.GetPubKey().Address()), ctbTime}, to}, ctbContent}
		case "post":
			ctb = contrib.Post{contrib.BaseContrib2{contrib.BaseContrib{key, sdk.AccAddress(info.GetPubKey().Address()), ctbTime}, to}, ctbContent}
		case "vote":
			switch m.VoteType {
			case "upvote":
				ctb = contrib.Vote{contrib.BaseContrib3{contrib.BaseContrib{key, sdk.AccAddress(info.GetPubKey().Address()), ctbTime}, to, 1}, ctbContent}
			case "downvote":
				ctb = contrib.Vote{contrib.BaseContrib3{contrib.BaseContrib{key, sdk.AccAddress(info.GetPubKey().Address()), ctbTime}, to, -1}, ctbContent}
			default:
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(err.Error()))
				return
			}
		default:
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

		// build message
		msg := client.BuildContribMsg(ctb)
		if err != nil { // XXX rechecking same error ?
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}

		txCtx := authctx.TxContext{
			Codec:         cdc,
			ChainID:       m.ChainID,
			AccountNumber: m.AccountNumber,
			Sequence:      m.Sequence,
			Gas:           m.Gas,
		}

		txBytes, err := txCtx.BuildAndSign(m.LocalAccountName, m.Password, []sdk.Msg{msg})
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(err.Error()))
			return
		}

		// send
		res, err := cliCtx.BroadcastTx(txBytes)
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
