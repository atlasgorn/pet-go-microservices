package core

import "errors"

var (
	ErrBadArguments   = errors.New("arguments are not acceptable")
	ErrPhraseTooLarge = errors.New("phrase cannot be larger than 4KiB")
)
