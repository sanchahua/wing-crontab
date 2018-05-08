package consul

import (
	"github.com/hashicorp/consul/api"
	log "github.com/sirupsen/logrus"
)

type Kv struct {
	kv *api.KV
}
func NewKv(kv *api.KV) *Kv {
	return &Kv{kv:kv}
}

func (k *Kv) Set(key string, value string) {
	log.Debugf("write %s=%s", key, value)
	kv := &api.KVPair{
		Key:key,
		Value:[]byte(value),
	}
	k.kv.Put(kv, nil)
}
func (k *Kv) Get(key string) string {
	kv, _, e := k.kv.Get(key, nil)
	if e != nil {
		log.Errorf("%+v", e)
		return ""
	}
	return string(kv.Value)
}
func (k *Kv) Delete(key string) {
	_, e := k.kv.Delete(key, nil)
	if e != nil {
		log.Errorf("%+v", e)
	}
}