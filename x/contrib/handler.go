package contrib

import (
	"reflect"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewHandler returns a handler for "contrib" type messages.
func NewHandler(k Keeper) sdk.Handler {
	return func(ctx sdk.Context, msg sdk.Msg) sdk.Result {
		switch msg := msg.(type) {
		case MsgContrib:
			return handleMsgContrib(ctx, k, msg)
		default:
			errMsg := "Unrecognized contrib Msg type: " + reflect.TypeOf(msg).Name()
			return sdk.ErrUnknownRequest(errMsg).Result()
		}
	}
}

// Handle MsgContrib.
func handleMsgContrib(ctx sdk.Context, k Keeper, msg MsgContrib) sdk.Result {
	tags := sdk.EmptyTags()

	for _, ctb := range msg.Contribs {
		err := k.UpdateContrib(ctx, ctb, &tags)
		if err != nil {
			return err.Result()
		}
	}

	return sdk.Result{
		Tags: tags,
	}
}
