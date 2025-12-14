package utils

import "errors"

var (
	ErrNoRowsUpdated  = errors.New("no rows affected when updating")
	FailedToGetMsg    = "failed to get "
	NotFoundMsg       = " not found"
	FailedToCreateMsg = "failed to create "
	FailedToUpdateMsg = "failed to update "
)
