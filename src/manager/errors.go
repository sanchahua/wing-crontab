package manager

import "errors"

const (
	HttpSuccess = 2000
	HttpErrorParamCronSet  = 4001
	HttpErrorParamCommand  = 4002
	HttpErrorCronModelFail = 5001
)

var (
	ErrSendNotComplete = errors.New("http send not complete")
)
