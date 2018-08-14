package service

import (
	"fmt"
)

// Error code
type oAuthErrorCode int

const (
	oAuthErrorAccessTokenEmpty       = 1
	oAuthErrorAccessTokenExpired     = 2
	oAuthErrorAccessTokenDecode      = 3
	oAuthErrorAccessTokenIllegal     = 4
	oAuthErrorAccessTokenBeyondScope = 5
	oAuthErrorAccessTokenNoKeySource = 6
)

var oAuthErrorCodeString = map[oAuthErrorCode]string{
	oAuthErrorAccessTokenEmpty:       "oAuthErrorAccessTokenEmpty",
	oAuthErrorAccessTokenExpired:     "oAuthErrorAccessTokenExpired",
	oAuthErrorAccessTokenDecode:      "oAuthErrorAccessTokenDecode",
	oAuthErrorAccessTokenIllegal:     "oAuthErrorAccessTokenIllegal",
	oAuthErrorAccessTokenBeyondScope: "oAuthErrorAccessTokenBeyondScope",
	oAuthErrorAccessTokenNoKeySource: "oAuthErrorAccessTokenNoKeySource",
}

func (c oAuthErrorCode) String() string {
	return oAuthErrorCodeString[c]
}

// Error
type oAuthError struct {
	code    oAuthErrorCode
	message string
}

func newOauthError(code oAuthErrorCode, message string) *oAuthError {
	return &oAuthError{
		code:    code,
		message: message,
	}
}

func (e oAuthError) Error() string {
	return fmt.Sprintf("oAuthError{ code: %v, message: %v}", e.code, e.message)
}
