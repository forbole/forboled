package client

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/forbole/forboled/x/contrib"
)

// build the contribTx msg
func BuildContribMsg(ctb contrib.Contrib) sdk.Msg {
	msg := contrib.NewMsgContrib(contrib.Contribs{ctb})
	return msg
}
