package contrib

import (
	"github.com/cosmos/cosmos-sdk/wire"
)

// var cdc = wire.NewCodec()

// Register concrete types on wire codec
func RegisterWire(cdc *wire.Codec) {
	cdc.RegisterInterface((*Contrib)(nil), nil)
	cdc.RegisterConcrete(&Invite{}, "contrib/Invite", nil)
	cdc.RegisterConcrete(&Recommend{}, "contrib/Recommend", nil)
	cdc.RegisterInterface((*Status)(nil), nil)
	cdc.RegisterConcrete(&InviteStatus{}, "contrib/InviteStatus", nil)
	cdc.RegisterConcrete(&RecommendStatus{}, "contrib/RecommendStatus", nil)
	cdc.RegisterConcrete(MsgContrib{}, "forbole/ContribMsg", nil)
}
