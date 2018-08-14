package service

import (
	"github.com/pkg/errors"
	"gitlab.xunlei.cn/xlsoa/service/log"
	"golang.org/x/net/context"
)

type oAuthCallCredentials struct {
	c       *ClientContext
	jwtSign string
	ts      *oAuthTokenSource
}

func NewOauthCallCredentials(c *ClientContext) (*oAuthCallCredentials, error) {
	s := &oAuthCallCredentials{c: c}

	s.jwtSign = c.env.GetJwtSign([]string{c.GetServiceName()})
	if s.jwtSign == "" {
		log.Printf("[xlsoa] [OauthCallCredentials] [Error] Get JwtSign fail, no access token will be acquired. Check 'creds.json' in environment.\n")
		return nil, errors.New("Get JwtSign fail")
	}
	log.Printf("[xlsoa] [OauthCallCredentials] Using jwtSign: '%v'\n", s.jwtSign)

	return s, nil
}

func (s *oAuthCallCredentials) Close() {
	if s.ts != nil {
		s.ts.Close()
		s.ts = nil
	}
}

func (s *oAuthCallCredentials) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	var err error
	var nopMeta = map[string]string{
		"authorization": "",
	}

	if s.ts == nil {
		if s.jwtSign == "" {
			return nopMeta, nil
		}
		log.Printf("[xlsoa] [OauthCallCredentials] Create token source with jwtSign: '%v'\n", s.jwtSign)

		if s.ts, err = NewOauthTokenSource(s.c, s.jwtSign); err != nil {
			return nopMeta, nil
		}

		log.Println("[xlsoa] [OauthCallCredentials] Create token source success")
	}

	token, err := s.ts.Token()
	if err != nil {
		log.Printf("[xlsoa] [OauthCallCredentials] [Error] Token source get Token() error: %v\n", err)
		return nopMeta, nil
	}

	return map[string]string{
		"authorization": token.Type() + " " + token.AccessToken,
	}, nil
}
