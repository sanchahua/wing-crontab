package service

import (
	"encoding/json"
	"github.com/pkg/errors"
	xlsoa_jws "gitlab.xunlei.cn/xlsoa/common/jws"
	"gitlab.xunlei.cn/xlsoa/common/utility"
	"gitlab.xunlei.cn/xlsoa/service/log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"strings"
)

type oAuthCertification struct {
	serviceId   string
	serviceName string
}

type oAuthGuardian struct {
	c  *ServerContext
	ks *oAuthKeySource
}

func NewOauthGuardian(c *ServerContext) (*oAuthGuardian, error) {
	var err error
	g := &oAuthGuardian{c: c}

	if err = g.init(); err != nil {
		return nil, errors.Wrap(err, "init fail")
	}

	return g, nil
}

func (g *oAuthGuardian) init() error {
	var err error

	jwtSign := g.c.GetEnv().GetJwtSign([]string{"/"})
	if jwtSign == "" {
		log.Printf("[xlsoa] [OauthGuardian] [Error] Get JwtSign fail, no oauth secure will startup. Check 'creds.json' in environment.\n")
		return errors.New("Get JwtSign fail")
	}

	log.Printf("[xlsoa] [OauthGuardian] Using jwtSign: '%v'\n", jwtSign)
	if g.ks, err = NewOauthKeySource(g.c, jwtSign); err != nil {
		return errors.Wrap(err, "NewOauthKeySource fail")
	}

	return nil

}

func (g *oAuthGuardian) Close() {
	if g.ks != nil {
		g.ks.Close()
		g.ks = nil
	}
}

func (g *oAuthGuardian) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var err error
		if ctx, err = g.interceptorImpl(ctx, info.FullMethod); err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

func (g *oAuthGuardian) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		var err error
		if _, err = g.interceptorImpl(stream.Context(), info.FullMethod); err != nil {
			return err
		}

		return handler(srv, stream)
	}
}

func (g *oAuthGuardian) interceptorImpl(ctx context.Context, fullMethodName string) (context.Context, error) {

	var err error
	var cert *oAuthCertification
	if cert, err = g.verify(ctx, fullMethodName); err != nil {
		if e, ok := err.(oAuthError); ok {
			ctx = context.WithValue(ctx, "soa_verify_status", int(e.code))
		} else {
			ctx = context.WithValue(ctx, "soa_verify_status", 100)
		}
	} else {
		ctx = context.WithValue(ctx, "soa_verify_status", 0)
	}

	// Set certification
	if cert != nil {
		ctx = context.WithValue(ctx, "soa_service_id", cert.serviceId)
		ctx = context.WithValue(ctx, "soa_service_name", cert.serviceName)
	}

	if err != nil && g.mustVerify() {
		log.Printf("[xlsoa] [OauthGuardian] [Error] oauth verify fail, and must verify, will block request. error: %v\n", err)
		return ctx, err
	}

	return ctx, nil

}

func (g *oAuthGuardian) verify(ctx context.Context, fullMethodName string) (*oAuthCertification, error) {

	var err error

	// Get token
	token := g.retrieveAccessToken(ctx)
	if token == "" {
		return nil, newOauthError(oAuthErrorAccessTokenEmpty, "Token not exisits")
	}

	// Decode JWS
	header, claim, _, claimBytes := xlsoa_jws.Decode(token)
	if header == nil || claim == nil {
		return nil, newOauthError(oAuthErrorAccessTokenDecode, "AccessToken illegal: JWS decode fail")
	}

	if xlsoa_jws.IsClaimSetExpired(claim) {
		return nil, newOauthError(oAuthErrorAccessTokenExpired, "AccessToken expires")
	}

	// Verify JWS
	key := g.ks.Get(header.KeyID)
	if key == nil {
		return nil, newOauthError(oAuthErrorAccessTokenIllegal, "AccessToken illegal: KeyID not exisits.")
	}

	if err = xlsoa_jws.Verify(token, key.secret); err != nil {
		return nil, newOauthError(oAuthErrorAccessTokenIllegal, "AccessToken illegal: JWT verify fail.")
	}

	// Retrive service info
	cert := &oAuthCertification{}
	privateClaim := map[string]interface{}{}
	if err = json.Unmarshal(claimBytes, &privateClaim); err != nil {
	} else {
		cert.serviceId = privateClaim["service_id"].(string)
		cert.serviceName = privateClaim["service_name"].(string)
	}

	// Verify scope
	if !g.verifyScope(fullMethodName, claim.Scope) { //期望的访问域不在授权范围内
		return cert, newOauthError(oAuthErrorAccessTokenBeyondScope, "AccessToken illegal: Beyond scope.")
	}

	return cert, nil
}

func (g *oAuthGuardian) retrieveAccessToken(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	auths, ok := md["authorization"]
	if !ok || len(auths) == 0 {
		return ""
	}

	// sp[0]: Token type
	// sp[1]: AccessToken
	// We take the first, if mutiple 'authorization' are provided.
	sp := strings.SplitN(auths[0], " ", 2)
	if len(sp) < 2 {
		return ""
	}

	return sp[1]
}

func (interceptor *oAuthGuardian) verifyScope(fullMethodName, scope string) bool {

	// Empty scope is for granting all methods
	if scope == "" {
		return true
	}

	methodName := utility.NewScopeParser(fullMethodName).MethodName
	methodName = strings.ToLower(methodName)

	sp := strings.Split(scope, " ")
	for _, s := range sp {
		s = strings.ToLower(s)
		if s == methodName {
			return true
		}
	}

	return false
}

func (g *oAuthGuardian) mustVerify() bool {
	if g.c.GetOauthSecureLevel() == OauthSecureLevelRigorous {
		// Always verify
		log.Println("[xlsoa] [OauthGuardian] Always verify with secure level 'OauthSecureLevelRigorous'")
		return true
	} else if g.ks.Status() != KeySourceSyncStatusOk {
		// If something wrong
		log.Println("[xlsoa] [OauthGuardian] Something error with KeySource, don't have to verify")
		return false
	} else if !g.ks.HasValid() {
		log.Println("[xlsoa] [OauthGuardian] [Warning] no valid key, dont't have to verify")
		return false
	}

	return true
}
