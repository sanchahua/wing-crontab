package manager

import "errors"

const (
	HttpSuccess = 2000
	HttpErrorParamCronSet  = 4001
	HttpErrorParamCommand  = 4002
	HttpErrorIdInvalid     = 4003
	HttpErrorParamInvalid  = 4004
	HttpErrorParamStartTime  = 4005
	HttpErrorParamEndTime  = 4006
	HttpErrorTimeoutInvalid     = 4007


	HttpErrorCronModelAddFail = 5001
	HttpErrorCronModelStopFalseFail = 5002
	HttpErrorCronModelDeleteFail = 5003
	HttpErrorCronModelUpdateFail = 5004
	HttpErrorCronModelMutexFalseFail = 5005


	HttpErrorCronControllerStopFalseFail = 6001
	HttpErrorCronControllerGetListJsonFail = 6002
	HttpErrorCronControllerMutexFail = 6003
	HttpErrorCronControllerGetFail = 6004
	HttpErrorCronControllerRunCommandFail = 6005

	HttpErrorParseFormFail = 7001
	HttpErrorGetUserByUserNameFail = 7002
	HttpErrorUserNotFound = 7003
	HttpErrorPasswordError = 7004
	HttpErrorAddUserFail = 7005


)

var (
	ErrSendNotComplete = errors.New("http send not complete")
)
