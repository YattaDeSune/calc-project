package errors

import "errors"

var (
	ErrUserExists    = errors.New("user already exists")
	ErrWrongLogin    = errors.New("invalid login")
	ErrWrongPassword = errors.New("invalid password")
)
