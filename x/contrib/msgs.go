package contrib

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// MsgContrib - high level transaction of the contrib module
type MsgContrib struct {
	Contribs Contribs `json:"contribs"`
}

var _ sdk.Msg = MsgContrib{}

// NewMsgContrib - construct arbitrary multi-in, multi-out contrib msg.
func NewMsgContrib(ctb []Contrib) MsgContrib {
	return MsgContrib{Contribs: ctb}
}

// Implements Msg.
func (msg MsgContrib) Type() string { return "contrib" } // TODO: "contrib/contrib"

// Implements Msg.
func (msg MsgContrib) ValidateBasic() sdk.Error {
	// this just makes sure all the contribs are properly formatted
	if len(msg.Contribs) == 0 {
		return ErrNoContribs(DefaultCodespace).Trace("")
	}

	// make sure all contribs are individually valid
	err := msg.Contribs.ValidateBasic()
	if err != nil {
		return err.Trace("")
	}

	return nil
}

// Implements Msg.
func (msg MsgContrib) GetSignBytes() []byte {
	b, err := json.Marshal(msg) // XXX: ensure some canonical form
	if err != nil {
		panic(err)
	}
	return b
}

// Implements Msg.
func (msg MsgContrib) GetSigners() []sdk.Address {
	m := make(map[string]struct{})
	addrs := make([]sdk.Address, 0, len(msg.Contribs))
	for _, ctb := range msg.Contribs {
		contributor := ctb.GetContributor()
		key := contributor.String()
		_, found := m[key]
		if !found {
			addrs = append(addrs, contributor)
			m[key] = struct{}{}
		}
	}
	return addrs
}
