package errors

import "errors"

var (
	ErrUserExists           = errors.New("user already exists")
	ErrDoesNotUserExist     = errors.New("user does not exist")
	ErrBadUserPassword      = errors.New("bad user password pair")
	ErrOrderIDAlreadyExists = errors.New("order id already exists")
)
