package config

import (
	"fmt"
	consul_api "github.com/hashicorp/consul/api"
	"gitlab.xunlei.cn/xlsoa/common/utility"
	//"strings"
)

type entry struct {
	key     string
	filters []*filter
	value   []byte

	filterMatchCount int
}

func newEntry() *entry {
	return &entry{
		filters: make([]*filter, 0),
	}
}

func (e *entry) decodeDepth(name string) {

	realName, props := parseProperties(name)
	e.key += fmt.Sprintf("/%v", realName)

	for k, v := range props {
		// value为空的property，跟没有指定是一个意思
		if k == "" {
			continue
		}

		e.filters = append(e.filters, newFilter(k, v))
	}

}

func (e *entry) decode(kv *consul_api.KVPair) {

	e.value = kv.Value

	pp := utility.NewPathPattern(kv.Key)
	for i := 0; i < pp.Depth(); i++ {
		e.decodeDepth(pp.LevelName(i))
	}
}

// bool: Filter ok
func (e *entry) match(properties map[string]string) bool {

	for _, filter := range e.filters {
		if !filter.match(properties) {
			return false
		}
		e.filterMatchCount++
	}
	return true
}
