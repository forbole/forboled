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

	"github.com/forbole/forboled/x/contrib"
	"github.com/forbole/forboled/x/contrib/client"
)

var msgCdc = wire.NewCodec()

func init() {
	contrib.RegisterWire(msgCdc)
}

// RegisterRoutes - Central function to define routes that get registered by the main application
func registerTxRoutes(ctx context.CoreContext, r *mux.Router, cdc *wire.Codec, kb keys.Keybase) {
	r.HandleFunc("/contrib/{address}/{ctbtype}", ContribRequestHandlerFn(cdc, kb, ctx)).Methods("POST")
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
func ContribRequestHandlerFn(cdc *wire.Codec, kb keys.Keybase, ctx context.CoreContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// collect data
		vars := mux.Vars(r)
		bech32addr := vars["address"]
		ctbtype := vars["ctbtype"]

		address, err := sdk.GetAccAddressBech32(bech32addr)
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

		to, err := sdk.GetAccAddressHex(address.String())
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			return
		}

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
			ctb = contrib.Invite{contrib.BaseContrib2{contrib.BaseContrib{key, info.GetPubKey().Address(), ctbTime}, to}, ctbContent}
		case "recommend":
			ctb = contrib.Recommend{contrib.BaseContrib2{contrib.BaseContrib{key, info.GetPubKey().Address(), ctbTime}, to}, ctbContent}
		case "post":
			ctb = contrib.Post{contrib.BaseContrib2{contrib.BaseContrib{key, info.GetPubKey().Address(), ctbTime}, to}, ctbContent}
		case "vote":
			switch m.VoteType {
			case "upvote":
				ctb = contrib.Vote{contrib.BaseContrib3{contrib.BaseContrib{key, info.GetPubKey().Address(), ctbTime}, to, 1}, ctbContent}
			case "downvote":
				ctb = contrib.Vote{contrib.BaseContrib3{contrib.BaseContrib{key, info.GetPubKey().Address(), ctbTime}, to, -1}, ctbContent}
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

		// add gas to context
		ctx = ctx.WithGas(m.Gas)
		// add chain-id to context
		ctx = ctx.WithChainID(m.ChainID)
		// sign
		ctx = ctx.WithAccountNumber(m.AccountNumber)
		ctx = ctx.WithSequence(m.Sequence)
		txBytes, err := ctx.SignAndBuild(m.LocalAccountName, m.Password, []sdk.Msg{msg}, cdc)
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
