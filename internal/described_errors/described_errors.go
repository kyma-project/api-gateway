package described_errors

import "github.com/pkg/errors"

type Level int

const (
	Error   Level = 0
	Warning Level = 1
)

// DescribedError wraps standard golang error with additional description to be set on Istio CR Status
type DescribedError interface {
	Description() string
	Error() string
	Level() Level
}

type DefaultDescribedError struct {
	err         error
	description string
	wrapError   bool
	level       Level
}

func NewDescribedError(err error, description string) DefaultDescribedError {
	return DefaultDescribedError{
		err:         err,
		description: description,
		wrapError:   true,
		level:       Error,
	}
}

func (d DefaultDescribedError) DisableErrorWrap() DefaultDescribedError {
	d.wrapError = false
	return d
}

func (d DefaultDescribedError) SetWarning() DefaultDescribedError {
	d.level = Warning
	return d
}

func (d DefaultDescribedError) Description() string {
	if d.wrapError {
		return errors.Wrap(d.err, d.description).Error()
	} else {
		return d.description
	}
}

func (d DefaultDescribedError) Error() string {
	return d.err.Error()
}

func (d DefaultDescribedError) Level() Level {
	return d.level
}
