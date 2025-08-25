package errors

import "errors"

var ErrorUserExists = errors.New("user already exists")
var ErrorBadUserPassword = errors.New("bad user password pair")
