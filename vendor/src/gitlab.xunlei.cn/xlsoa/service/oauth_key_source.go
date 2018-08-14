package service

import (
	"fmt"
	"github.com/pkg/errors"
	"gitlab.xunlei.cn/xlsoa/service/log"
	xlsoa_core "gitlab.xunlei.cn/xlsoa/service/proto_gen/xlsoa/core"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"sync"
	"time"
)

type oAuthKey struct {
	id      string
	secret  string
	version int32
	expires time.Time
}

func (k *oAuthKey) valid() bool {
	return k.expires.After(time.Now())
}

func (k *oAuthKey) String() string {
	return fmt.Sprintf("OauthKey{id:%s, secret:%s, version:%d, expireds:%v}",
		k.id,
		k.secret,
		k.version,
		k.expires,
	)
}

type oAuthKeyChan chan []*oAuthKey
type oAuthKeyWatchable interface {
	Watch() (oAuthKeyChan, error)
}

const (
	KeySourceSyncStatusOk    = 0
	KeySourceSyncStatusError = 1
)

type oAuthKeySource struct {
	c             *ServerContext
	jwtSign       string
	conn          *grpc.ClientConn
	watchers      []oAuthKeyChan
	latestVersion int32
	stopCh        chan bool

	keyMutex        sync.Mutex
	keys            map[string]*oAuthKey
	garbageInterval time.Duration
	status          int
}

func NewOauthKeySource(c *ServerContext, jwtSign string) (*oAuthKeySource, error) {
	s := &oAuthKeySource{
		c:               c,
		jwtSign:         jwtSign,
		watchers:        make([]oAuthKeyChan, 0),
		stopCh:          make(chan bool),
		keys:            make(map[string]*oAuthKey),
		garbageInterval: 1 * time.Hour,
		status:          KeySourceSyncStatusOk,
	}

	go s.run()
	go s.gc()
	return s, nil
}

func (s *oAuthKeySource) Close() {
	close(s.stopCh)
}

func (s *oAuthKeySource) HasValid() bool {
	s.keyMutex.Lock()
	defer s.keyMutex.Unlock()

	for _, v := range s.keys {
		if v.valid() {
			return true
		}
	}
	return false
}

func (s *oAuthKeySource) Status() int {
	return s.status
}

func (s *oAuthKeySource) Get(id string) *oAuthKey {
	s.keyMutex.Lock()
	defer s.keyMutex.Unlock()

	var key *oAuthKey
	var ok bool
	if key, ok = s.keys[id]; !ok || !key.valid() {
		return nil
	}

	return key
}

func (s *oAuthKeySource) run() {

	running := true
	for running {
		var err error
		var keys []*oAuthKey
		var interval int32

		// Sync updated
		if keys, interval, err = s.sync(); err != nil {
			log.Printf("[xlsoa] [OauthKeySource] [Error] sync() error: %v\n", err)
			s.status = KeySourceSyncStatusError // Set status
			time.Sleep(10 * time.Second)
			continue
		}
		log.Printf("[xlsoa] [OauthKeySource] Keys synced: '%v', interval: %v\n", keys, interval)
		s.status = KeySourceSyncStatusOk // Set status
		s.append(keys...)

		// Sleep next interval
		if interval <= 0 {
			interval = 10
		}

		select {
		case <-time.After(time.Duration(interval) * time.Second):
		case <-s.stopCh:
			running = false
		}

	}
}

func (s *oAuthKeySource) gc() {

	running := true
	for running {
		select {
		case <-time.After(s.garbageInterval):
			s.removeInvalid()
		case <-s.stopCh:
			running = false
		}
	}

}

func (s *oAuthKeySource) sync() ([]*oAuthKey, int32, error) {

	var err error
	var resp *xlsoa_core.SyncKeyResponse

	if s.conn == nil {

		if s.conn, err = grpc.Dial(
			CERTIFICATE_AUTHORITY_SERVICE_NAME,
			grpc.WithInsecure(),
			grpc.WithDialer(s.c.GetEnv().GrpcDialer()),
		); err != nil {
			return nil, 0, errors.Wrap(err, "grpc dial fail")
		}

	}

	req := &xlsoa_core.SyncKeyRequest{}
	req.LatestVersion = s.latestVersion
	req.Assertion = s.jwtSign
	c := xlsoa_core.NewCertificateClient(s.conn)

	if resp, err = c.SyncKey(context.Background(), req); err != nil {
		return nil, 0, errors.Wrap(err, "SyncKey fail")
	}

	if resp.Result != xlsoa_core.CertificateResult_OK {
		return nil, 0, errors.New(fmt.Sprintf("SyncKey resp.Result fail: %v, %v", resp.Result, resp.Message))
	}

	keys := []*oAuthKey{}
	for _, key := range resp.Keys {

		aKey := &oAuthKey{
			id:      key.Id,
			secret:  key.Secret,
			version: key.Version,
			expires: time.Now().Add(time.Duration(key.TimeToLive) * time.Second),
		}
		keys = append(keys, aKey)

	}

	// Update latest version
	if len(keys) > 0 {
		s.latestVersion = keys[len(keys)-1].version
	}

	return keys, resp.NextInterval, nil
}

func (s *oAuthKeySource) append(keys ...*oAuthKey) {
	s.keyMutex.Lock()
	defer s.keyMutex.Unlock()

	for _, key := range keys {
		s.keys[key.id] = key
	}
}

func (s *oAuthKeySource) removeInvalid() {
	s.keyMutex.Lock()
	defer s.keyMutex.Unlock()

	for k, v := range s.keys {
		if v.valid() {
			continue
		}
		delete(s.keys, k)
	}
}
