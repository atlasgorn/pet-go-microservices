package core

import "errors"

var (
	ErrBadArguments   = errors.New("arguments are not acceptable")
	ErrAlreadyExists  = errors.New("resource or task already exists")
	ErrPhraseTooLarge = errors.New("phrase cannot be larger than 4KiB")
)
