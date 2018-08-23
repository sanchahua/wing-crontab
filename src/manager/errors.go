package manager

import "errors"

const (
	HttpSuccess = 2000
	HttpErrorParamCronSet  = 4001
	HttpErrorParamCommand  = 4002
	HttpErrorIdInvalid     = 4003

	HttpErrorCronModelAddFail = 5001
	HttpErrorCronModelStopFalseFail = 5002
	HttpErrorCronModelDeleteFail = 5003
	HttpErrorCronModelUpdateFail = 5004

	HttpErrorCronControllerStopFalseFail = 6001
	HttpErrorCronControllerGetListJsonFail = 6002
)

var (
	ErrSendNotComplete = errors.New("http send not complete")
)
