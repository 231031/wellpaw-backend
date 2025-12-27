package utils

import "errors"

var (
	ErrNoRowsUpdated  = errors.New("no rows affected when updating")
	FailedToGetMsg    = "failed to get "
	NotFoundMsg       = " not found"
	FailedToCreateMsg = "failed to create "
	FailedToUpdateMsg = "failed to update "

	ErrUnauth       = errors.New("the token is invalid")
	ErrUnauthHeader = errors.New("the user is unauthorization")
	ErrFailToGet    = errors.New("failed to get data")
)
