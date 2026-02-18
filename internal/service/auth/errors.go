package authservice

import (
	"errors"
	"fmt"
)

var ErrInvalidToken = errors.New("invalid token")

type ErrUserExists struct {
	err error
}

func (eu *ErrUserExists) Error() string {
	return eu.err.Error()
}

func (eu *ErrUserExists) Unwrap() error {
	return eu.err
}

func (eu *ErrUserExists) Is(target error) bool {
	_, ok := target.(*ErrUserExists)
	return ok
}

func NewErrUserExists(err error) *ErrUserExists {
	if err == nil {
		err = fmt.Errorf("user exists")
	}

	return &ErrUserExists{
		err: fmt.Errorf("%w", err),
	}
}
