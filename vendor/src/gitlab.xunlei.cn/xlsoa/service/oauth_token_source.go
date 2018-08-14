package service

import (
	"fmt"
	"github.com/pkg/errors"
	"gitlab.xunlei.cn/xlsoa/service/log"
	xlsoa_core "gitlab.xunlei.cn/xlsoa/service/proto_gen/xlsoa/core"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/jws"
	"google.golang.org/grpc"
	"sync"
	"time"
)

type oAuthTokenSource struct {
	c       *ClientContext
	jwtSign string
	conn    *grpc.ClientConn
	mu      sync.Mutex // guards token
	token   *oauth2.Token
}

func NewOauthTokenSource(c *ClientContext, jwtSign string) (*oAuthTokenSource, error) {
	s := &oAuthTokenSource{
		c:       c,
		jwtSign: jwtSign,
	}

	return s, nil
}

func (s *oAuthTokenSource) Close() {
	// Nop
}

func (s *oAuthTokenSource) Token() (*oauth2.Token, error) {

	s.mu.Lock()
	defer s.mu.Unlock()

	var err error

	// Refresh new token
	// Best efford
	if s.token == nil || !s.token.Valid() {
		var token *oauth2.Token
		if token, err = s.fetch(); err == nil {
			log.Printf("[xlsoa] [OauthTokenSource] New token fetched: %v\n", token)
			s.token = token
		} else {
			log.Printf("[xlsoa] [OauthTokenSource] [Error] Fetch token error: %v\n", err)
		}
	}

	if s.token == nil {
		return nil, errors.New("Token not ready")
	}

	return s.token, nil
}

func (s *oAuthTokenSource) fetch() (*oauth2.Token, error) {
	var err error
	var resp *xlsoa_core.AuthorizeResponse

	if s.conn == nil {
		if s.conn, err = grpc.Dial(
			CERTIFICATE_AUTHORITY_SERVICE_NAME,
			grpc.WithInsecure(),
			grpc.WithDialer(s.c.GetEnv().GrpcDialer()),
		); err != nil {
			return nil, errors.Wrap(err, "grpc dial fail")
		}
	}

	c := xlsoa_core.NewCertificateClient(s.conn)
	req := &xlsoa_core.AuthorizeRequest{
		GrantType: "urn:ietf:params:oauth:grant-type:jwt-bearer",
		Assertion: s.jwtSign,
	}

	resp, err = c.Authorize(context.Background(), req)
	if err != nil {
		return nil, errors.Wrap(err, "Authenticate rpc error")
	}

	if resp.Result != xlsoa_core.CertificateResult_OK {
		return nil, errors.New(fmt.Sprintf("AuthenticateResp error. Result: %v, Message: '%v'", resp.Result, resp.Message))
	}

	//获取token正常
	token := &oauth2.Token{
		AccessToken: resp.AccessToken,
		TokenType:   resp.TokenType,
	}

	if secs := resp.ExpiresIn; secs > 0 {
		// 10 seconds beyond
		token.Expiry = time.Now().Add(time.Duration(secs-10) * time.Second)
	}

	// TODO: with IdToken for?
	if v := resp.IdToken; v != "" {
		// decode returned id token to get expiry
		claimSet, err := jws.Decode(v)
		if err != nil {
			return nil, fmt.Errorf("oauth2: error decoding JWT token: %v", err)
		}
		token.Expiry = time.Unix(claimSet.Exp, 0)
	}

	return token, nil
}
