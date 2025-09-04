package errors

import "errors"

var (
	ErrorUserExists           = errors.New("user already exists")
	ErrorDoesNotUserExist     = errors.New("user does not exist")
	ErrorBadUserPassword      = errors.New("bad user password pair")
	ErrorOrderIdAlreadyExists = errors.New("order id already exists")
)
