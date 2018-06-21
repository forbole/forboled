package contrib

import (
	"bytes"
	"fmt"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/auth"
)

//----------------------------------------
// Contrib

type Contrib interface {
	GetKey() []byte
	GetContributor() sdk.Address
	GetTime() time.Time
	AppendTags(*sdk.Tags)
	NewStatus() Status
	ValidateBasic() sdk.Error
	ValidateAccounts(sdk.Context, auth.AccountMapper) (auth.Account, sdk.Error)
	String() string
}

type Contribs []Contrib

// ValidateBasic - validate transaction contribs
func (contribs Contribs) ValidateBasic() sdk.Error {
	// m := make(map[string]struct{})
	for _, ctb := range contribs {
		err := ctb.ValidateBasic()
		if err != nil {
			return err
		}
		// _, found := m[string(ctb.Key)]
		// if found {
		// 	return ErrInvalidContrib(DefaultCodespace, "duplicate key")
		// }
		// m[string(ctb.Key)] = struct{}{}
	}

	return nil
}

type BaseContrib struct {
	Key         []byte      `json:"key"`
	Contributor sdk.Address `json:"contributor"`
	Time        time.Time   `json:"time"`
}

// Implements Contrib
func (ctb BaseContrib) GetKey() []byte {
	return ctb.Key
}

func (ctb BaseContrib) GetContributor() sdk.Address {
	return ctb.Contributor
}

func (ctb BaseContrib) GetTime() time.Time {
	return ctb.Time
}

func (ctb BaseContrib) AppendTags(tags *sdk.Tags) {
	*tags = append(*tags, sdk.MakeTag("contributor", ctb.Contributor.Bytes()))
}

func (ctb BaseContrib) NewStatus() Status {
	return &BaseStatus{Score: 1, Contributor: ctb.Contributor, Time: ctb.Time}
}

func (ctb BaseContrib) ValidateBasic() sdk.Error {
	if len(ctb.Key) == 0 {
		return ErrInvalidContrib(DefaultCodespace, ctb.String())
	}
	if len(ctb.Contributor) == 0 {
		return sdk.ErrInvalidAddress(ctb.Contributor.String())
	}
	return nil
}

func (ctb BaseContrib) ValidateAccounts(ctx sdk.Context, am auth.AccountMapper) (auth.Account, sdk.Error) {
	acc := am.GetAccount(ctx, ctb.Contributor)
	if acc == nil {
		return nil, sdk.ErrUnknownAddress(ctb.Contributor.String())
	}
	return acc, nil
}

func (ctb BaseContrib) String() string {
	return fmt.Sprintf("%v", ctb)
}

type BaseContrib2 struct {
	BaseContrib
	Recipient sdk.Address `json:"recipient"`
}

func (ctb BaseContrib2) AppendTags(tags *sdk.Tags) {
	*tags = append(*tags, sdk.MakeTag("contributor", ctb.Contributor.Bytes()), sdk.MakeTag("recipient", ctb.Recipient.Bytes()))
}

func (ctb BaseContrib2) NewStatus() Status {
	return &BaseStatus2{BaseStatus: BaseStatus{Score: 1, Contributor: ctb.Contributor, Time: ctb.Time}, Recipient: ctb.Recipient}
}

func (ctb BaseContrib2) ValidateBasic() sdk.Error {
	if len(ctb.Key) == 0 {
		return ErrInvalidContrib(DefaultCodespace, ctb.String())
	}
	if len(ctb.Contributor) == 0 {
		return sdk.ErrInvalidAddress(ctb.Contributor.String())
	}
	if len(ctb.Recipient) == 0 {
		return sdk.ErrInvalidAddress(ctb.Recipient.String())
	}
	return nil
}

func (ctb BaseContrib2) ValidateAccounts(ctx sdk.Context, am auth.AccountMapper) (auth.Account, sdk.Error) {
	acc := am.GetAccount(ctx, ctb.Contributor)
	if acc == nil {
		return nil, sdk.ErrUnknownAddress(ctb.Contributor.String())
	}

	if am.GetAccount(ctx, ctb.Recipient) == nil {
		return nil, sdk.ErrUnknownAddress(ctb.Recipient.String())
	}

	return acc, nil
}

type Invite struct {
	BaseContrib2
	Content []byte `json:"content"`
}

func (ctb Invite) NewStatus() Status {
	return &InviteStatus{BaseStatus: BaseStatus{Score: 1, Contributor: ctb.Contributor, Time: ctb.Time}, Recipient: ctb.Recipient}
}

func (ctb Invite) ValidateAccounts(ctx sdk.Context, am auth.AccountMapper) (auth.Account, sdk.Error) {
	acc := am.GetAccount(ctx, ctb.Contributor)
	if acc == nil {
		return nil, sdk.ErrUnknownAddress(ctb.Contributor.String())
	}
	if am.GetAccount(ctx, ctb.Recipient) != nil {
		return nil, sdk.ErrUnknownAddress(ctb.Recipient.String())
	}

	am.SetAccount(ctx, am.NewAccountWithAddress(ctx, ctb.Recipient))

	return acc, nil
}

type Recommend struct {
	BaseContrib2
	Content []byte `json:"content"`
}

func (ctb Recommend) NewStatus() Status {
	return &RecommendStatus{BaseStatus: BaseStatus{Score: 1, Contributor: ctb.Contributor, Time: ctb.Time}, Recipient: ctb.Recipient}
}

func (ctb Recommend) ValidateAccounts(ctx sdk.Context, am auth.AccountMapper) (auth.Account, sdk.Error) {
	acc := am.GetAccount(ctx, ctb.Contributor)
	if acc == nil {
		return nil, sdk.ErrUnknownAddress(ctb.Contributor.String())
	}
	if am.GetAccount(ctx, ctb.Recipient) == nil {
		return nil, sdk.ErrUnknownAddress(ctb.Recipient.String())
	}
	// cannot recommend his/herself
	if acc == am.GetAccount(ctx, ctb.Recipient) {
		return nil, sdk.ErrInvalidAddress(ctb.Recipient.String())
	}

	return acc, nil
}

type Post struct {
	BaseContrib2
	Content []byte `json:"content"`
}

func (ctb Post) NewStatus() Status {
	return &PostStatus{BaseStatus: BaseStatus{Score: 1, Contributor: ctb.Contributor, Time: ctb.Time}, Recipient: ctb.Recipient}
}

type BaseContrib3 struct {
	BaseContrib
	Recipient sdk.Address `json:"recipient"`
	Vote      int64       `json:"vote"`
}

func (ctb BaseContrib3) AppendTags(tags *sdk.Tags) {
	*tags = append(*tags, sdk.MakeTag("contributor", ctb.Contributor.Bytes()), sdk.MakeTag("recipient", ctb.Recipient.Bytes()))
}

func (ctb BaseContrib3) NewStatus() Status {
	return &BaseStatus3{BaseStatus: BaseStatus{Score: 1, Contributor: ctb.Contributor, Time: ctb.Time}, Recipient: ctb.Recipient, Vote: ctb.Vote}
}

func (ctb BaseContrib3) ValidateBasic() sdk.Error {
	if len(ctb.Key) == 0 {
		return ErrInvalidContrib(DefaultCodespace, ctb.String())
	}
	if len(ctb.Contributor) == 0 {
		return sdk.ErrInvalidAddress(ctb.Contributor.String())
	}
	if len(ctb.Recipient) == 0 {
		return sdk.ErrInvalidAddress(ctb.Recipient.String())
	}
	return nil
}

func (ctb BaseContrib3) ValidateAccounts(ctx sdk.Context, am auth.AccountMapper) (auth.Account, sdk.Error) {
	acc := am.GetAccount(ctx, ctb.Contributor)
	if acc == nil {
		return nil, sdk.ErrUnknownAddress(ctb.Contributor.String())
	}

	if am.GetAccount(ctx, ctb.Recipient) == nil {
		return nil, sdk.ErrUnknownAddress(ctb.Recipient.String())
	}

	return acc, nil
}

type Vote struct {
	BaseContrib3
	Content []byte `json:"content"`
}

//new status of vote always start from 1, showing difference between up and down will be in update()
func (ctb Vote) NewStatus() Status {
	return &VoteStatus{BaseStatus: BaseStatus{Score: 1, Contributor: ctb.Contributor, Time: ctb.Time}, Recipient: ctb.Recipient, Vote: ctb.Vote}
}

func (ctb Vote) GetVote() int64 {
	return ctb.Vote
}

// Status - contrib status
type Status interface {
	GetScore() int64
	Update(Contrib) sdk.Error
}

type BaseStatus struct {
	Score       int64       `json:"score"`
	Contributor sdk.Address `json:"contributor"`
	Time        time.Time   `json:"time"`
}

func (status BaseStatus) GetScore() int64 {
	return status.Score
}

func (status *BaseStatus) Update(ctb Contrib) sdk.Error {
	// check if addr is the contributor
	if !bytes.Equal(ctb.GetContributor(), status.Contributor) {
		return sdk.ErrUnknownAddress("contributor error")
	}
	// check if time is valid
	if !ctb.GetTime().After(status.Time) {
		return sdk.ErrUnknownAddress("time error")
	}
	// TODO: better score calculation
	status.Score++
	status.Time = ctb.GetTime()
	return nil
}

type BaseStatus2 struct {
	BaseStatus
	Recipient sdk.Address `json:"recipient"`
}

func (status *BaseStatus2) Update(ctb Contrib) sdk.Error {
	ctb2 := ctb.(BaseContrib2)
	// check if addr is the contributor
	if !bytes.Equal(ctb2.Contributor, status.Contributor) {
		return sdk.ErrUnknownAddress("contributor error")
	}
	// check if the recipient is matched
	if !bytes.Equal(ctb2.Recipient, status.Recipient) {
		return sdk.ErrUnknownAddress("recipient error")
	}
	// check if time is valid
	if !ctb2.Time.After(status.Time) {
		return sdk.ErrUnknownAddress("time error")
	}
	// TODO: better score calculation
	status.Score++
	status.Time = ctb.GetTime()
	return nil
}

type InviteStatus BaseStatus2

func (status *InviteStatus) Update(ctb Contrib) sdk.Error {
	ctb2 := ctb.(*Invite)
	// check if addr is the contributor
	if !bytes.Equal(ctb2.Contributor, status.Contributor) {
		return sdk.ErrUnknownAddress("contributor error")
	}
	// check if the recipient is matched
	if !bytes.Equal(ctb2.Recipient, status.Recipient) {
		return sdk.ErrUnknownAddress("recipient error")
	}
	// check if time is valid
	if !ctb2.Time.After(status.Time) {
		return sdk.ErrUnknownAddress("time error")
	}

	status.Time = ctb.GetTime()
	return nil
}

type RecommendStatus BaseStatus2

type PostStatus BaseStatus2

//update() will be using the BaseStatus2's update()

type BaseStatus3 struct {
	BaseStatus
	Recipient sdk.Address `json:"recipient"`
	Vote      int64       `json:"vote"`
}

func (status *BaseStatus3) Update(ctb Contrib) sdk.Error {
	ctb2 := ctb.(BaseContrib3)
	// check if addr is the contributor
	if !bytes.Equal(ctb2.Contributor, status.Contributor) {
		return sdk.ErrUnknownAddress("contributor error")
	}
	// check if the recipient is matched
	if !bytes.Equal(ctb2.Recipient, status.Recipient) {
		return sdk.ErrUnknownAddress("recipient error")
	}
	// check if time is valid
	if !ctb2.Time.After(status.Time) {
		return sdk.ErrUnknownAddress("time error")
	}
	// TODO: better score calculation
	status.Score++
	status.Time = ctb.GetTime()
	return nil
}

type VoteStatus BaseStatus3

func (status *VoteStatus) Update(ctb Contrib) sdk.Error {
	ctb2 := ctb.(*Vote)
	// check if addr is the contributor
	if !bytes.Equal(ctb2.Contributor, status.Contributor) {
		return sdk.ErrUnknownAddress("contributor error")
	}
	// check if the recipient is matched
	if !bytes.Equal(ctb2.Recipient, status.Recipient) {
		return sdk.ErrUnknownAddress("recipient error")
	}
	// check if time is valid
	if !ctb2.GetTime().After(status.Time) {
		return sdk.ErrUnknownAddress("time error")
	}
	// not sure if status.Vote is getting from previous status, and GetVote() is the current vote
	if status.Vote == ctb2.GetVote() {
		// cancel the vote here
		// TODO: need better score here. (plus 2 to show vote is cancelled)
		status.Score += 2
		// change vote status to 0 (vote is cancelled)
		status.Vote = 0
	} else {

		// TODO: better score calculation ??????
		status.Score++
		// change vote status
		status.Vote = ctb2.GetVote()
	}

	status.Time = ctb.GetTime()
	return nil
}
