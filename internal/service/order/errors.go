package order

import "fmt"

type ErrWrongLuna struct {
	err error
}

func (ewl *ErrWrongLuna) Error() string {
	return ewl.err.Error()
}

func (ewl *ErrWrongLuna) Unwrap() error {
	return ewl.err
}

func (ewl *ErrWrongLuna) Is(target error) bool {
	_, ok := target.(*ErrWrongLuna)
	return ok
}

func NewErrWrongLuna(err error) *ErrWrongLuna {
	if err == nil {
		err = fmt.Errorf("wrong number, luna algorithm failed")
	}

	return &ErrWrongLuna{
		err: fmt.Errorf("%w", err),
	}
}

type ErrBadOrderID struct {
	err error
}

func (ebo *ErrBadOrderID) Error() string {
	return ebo.err.Error()
}

func (ebo *ErrBadOrderID) Unwrap() error {
	return ebo.err
}

func (ebo *ErrBadOrderID) Is(target error) bool {
	_, ok := target.(*ErrBadOrderID)
	return ok
}

func NewErrBadOrderID(err error) *ErrBadOrderID {
	if err == nil {
		err = fmt.Errorf("got wrong string")
	}

	return &ErrBadOrderID{
		err: fmt.Errorf("%w", err),
	}
}

type ErrWrongUser struct {
	err error
}

func (ewu *ErrWrongUser) Error() string {
	return ewu.err.Error()
}

func (ewu *ErrWrongUser) Unwrap() error {
	return ewu.err
}

func (ewu *ErrWrongUser) Is(target error) bool {
	_, ok := target.(*ErrWrongUser)
	return ok
}

func NewErrWrongUser(err error) *ErrWrongUser {
	if err == nil {
		err = fmt.Errorf("this order has owner")
	}

	return &ErrWrongUser{
		err: fmt.Errorf("%w", err),
	}
}

type ErrOrderExists struct {
	err error
}

func (eoe *ErrOrderExists) Error() string {
	return eoe.err.Error()
}

func (eoe *ErrOrderExists) Unwrap() error {
	return eoe.err
}

func (eoe *ErrOrderExists) Is(target error) bool {
	_, ok := target.(*ErrOrderExists)
	return ok
}

func NewErrOrderExists(err error) *ErrOrderExists {
	if err == nil {
		err = fmt.Errorf("this order has owner")
	}

	return &ErrOrderExists{
		err: fmt.Errorf("%w", err),
	}
}
