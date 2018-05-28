//nolint
package contrib

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// Contrib errors reserve 900 ~ 999.
const (
	DefaultCodespace sdk.CodespaceType = 2

	CodeInvalidInput   sdk.CodeType = 901
	CodeInvalidOutput  sdk.CodeType = 902
	CodeInvalidContrib sdk.CodeType = 903
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
