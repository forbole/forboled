//nolint
package contrib

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Contrib errors reserve 900 ~ 999.
const (
	DefaultCodespace sdk.CodespaceType = 2

	CodeNotValidator     sdk.CodeType = 1101
	CodeAlreadyProcessed sdk.CodeType = 1102
	CodeAlreadySigned    sdk.CodeType = 1103
	CodeUnknownRequest   sdk.CodeType = sdk.CodeUnknownRequest
	CodeInvalidInput     sdk.CodeType = 901
	CodeInvalidOutput    sdk.CodeType = 902
	CodeInvalidContrib   sdk.CodeType = 903
)

// NOTE: Don't stringer this, we'll put better messages in later.
func codeToDefaultMsg(code sdk.CodeType) string {
	switch code {
	case CodeInvalidInput:
		return "Invalid input coins"
	case CodeInvalidOutput:
		return "Invalid output coins"
	case CodeInvalidContrib:
		return "Invalid contrib"
	default:
		return sdk.CodeToDefaultMsg(code)
	}
}

//----------------------------------------
// Error constructors

// ErrNotValidator called when the signer of a Msg is not a validator
func ErrNotValidator(codespace sdk.CodespaceType, address sdk.AccAddress) sdk.Error {
	return sdk.NewError(codespace, CodeNotValidator, address.String())
}

// ErrAlreadyProcessed called when a payload is already processed
func ErrAlreadyProcessed(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeAlreadyProcessed, "")
}

// ErrAlreadySigned called when the signer is trying to double signing
func ErrAlreadySigned(codespace sdk.CodespaceType) sdk.Error {
	return sdk.NewError(codespace, CodeAlreadySigned, "")
}

func ErrInvalidInput(codespace sdk.CodespaceType, msg string) sdk.Error {
	return newError(codespace, CodeInvalidInput, msg)
}

func ErrNoInputs(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeInvalidInput, "")
}

func ErrInvalidOutput(codespace sdk.CodespaceType, msg string) sdk.Error {
	return newError(codespace, CodeInvalidOutput, msg)
}

func ErrNoOutputs(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeInvalidOutput, "")
}

func ErrInvalidContrib(codespace sdk.CodespaceType, msg string) sdk.Error {
	return newError(codespace, CodeInvalidContrib, msg)
}

func ErrNoContribs(codespace sdk.CodespaceType) sdk.Error {
	return newError(codespace, CodeInvalidContrib, "")
}

//----------------------------------------

func msgOrDefaultMsg(msg string, code sdk.CodeType) string {
	if msg != "" {
		return msg
	}
	return codeToDefaultMsg(code)
}

func newError(codespace sdk.CodespaceType, code sdk.CodeType, msg string) sdk.Error {
	msg = msgOrDefaultMsg(msg, code)
	return sdk.NewError(codespace, code, msg)
}
