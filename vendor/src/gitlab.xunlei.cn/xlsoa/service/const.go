package service

const (
	ENV_DATACENTER_NAME = "XLSOA_DC"
	ENV_NODE_NAME       = "XLSOA_NODE"

	CERTIFICATE_AUTHORITY_SERVICE_NAME = "xlsoa.core.certificate"
)

type OauthSecureLevel int

const (
	// Security should be degraded when there is system exceptions.
	// e.g There is no local keys to verify access token.
	OauthSecureLevelDegradeWhenException OauthSecureLevel = 0

	// Always keep higest secure level.
	OauthSecureLevelRigorous OauthSecureLevel = 10
)

var oAuthSecureLevelString = map[OauthSecureLevel]string{
	OauthSecureLevelDegradeWhenException: "OauthSecureLevelDegradeWhenException",
	OauthSecureLevelRigorous:             "OauthSecureLevelRigorous",
}

var oAuthSecureLevelStringToLevel = map[string]OauthSecureLevel{
	"OauthSecureLevelDegradeWhenException": OauthSecureLevelDegradeWhenException,
	"OauthSecureLevelRigorous":             OauthSecureLevelRigorous,
}

func (l OauthSecureLevel) String() string {
	return oAuthSecureLevelString[l]
}
