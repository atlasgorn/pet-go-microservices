package core

import "errors"

var (
	ErrBadArguments   = errors.New("arguments are not acceptable")
	ErrAlreadyExists  = errors.New("resource or task already exists")
	ErrNotFound       = errors.New("resource is not found")
	ErrPhraseTooLarge = errors.New("phrase cannot be larger than 4KiB")
)
