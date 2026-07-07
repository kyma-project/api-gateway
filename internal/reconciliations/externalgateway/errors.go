package externalgateway

import (
	"errors"
	"fmt"
)

// ReasonedError is an error that carries a status-condition reason string.
// The reason is meant to be propagated to the ExternalGateway Ready condition
// so consumers can distinguish failure classes without parsing the message.
type ReasonedError struct {
	Reason  string
	Message string
}

func (e *ReasonedError) Error() string {
	return e.Message
}

// NewReasonedError returns a ReasonedError with the given reason and formatted message.
func NewReasonedError(reason, format string, args ...any) *ReasonedError {
	return &ReasonedError{
		Reason:  reason,
		Message: fmt.Sprintf(format, args...),
	}
}

// ErrorReason unwraps err looking for a ReasonedError and returns its reason.
// It returns ("", false) if err does not contain a ReasonedError.
func ErrorReason(err error) (string, bool) {
	if target, ok := errors.AsType[*ReasonedError](err); ok {
		return target.Reason, true
	}
	return "", false
}
