package balance

import (
	"errors"
	"fmt"
)

var ErrEmptyUsername = errors.New("empty username")
var ErrEmptyInputData = errors.New("empty input data")

type ErrNoMoney struct {
	err error
}

func (eoe *ErrNoMoney) Error() string {
	return eoe.err.Error()
}

func (eoe *ErrNoMoney) Unwrap() error {
	return eoe.err
}

func (eoe *ErrNoMoney) Is(target error) bool {
	_, ok := target.(*ErrNoMoney)
	return ok
}

func NewErrNoMoney(err error) *ErrNoMoney {
	if err == nil {
		err = fmt.Errorf("too little bonuses")
	}

	return &ErrNoMoney{
		err: fmt.Errorf("%w", err),
	}
}
