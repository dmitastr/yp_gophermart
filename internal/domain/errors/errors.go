package errors

import "errors"

var ErrorUserExists = errors.New("user already exists")
var ErrorDoesNotUserExist = errors.New("user does not exist")
var ErrorBadUserPassword = errors.New("bad user password pair")
