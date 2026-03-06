package utils

import (
	"fmt"
	"strconv"
	"strings"
)

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

func IsLuna(oid string) error {
	nums := make([]int, 0, len(oid))
	for _, l := range strings.Split(oid, "") {
		i, err := strconv.Atoi(l)
		if err != nil {
			return NewErrBadOrderID(err)
		}

		nums = append(nums, i)
	}

	acc1 := 0
	var divider int

	switch len(nums) % 2 {
	case 0:
		divider = 0
	case 1:
		divider = 1
	}

	for i, n := range nums {
		if i%2 == divider {
			if n*2/10 == 1 {
				acc1 += n*2 - 9
			} else {
				acc1 += n * 2
			}
		} else {
			acc1 += n
		}
	}

	if acc1%10 == 0 {
		return nil
	} else {
		return NewErrWrongLuna(nil)
	}
}
