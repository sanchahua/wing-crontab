package config

import (
	consul_api "github.com/hashicorp/consul/api"
	"testing"
)

func TestEntryDecode(t *testing.T) {

	for i, c := range []struct {
		keyOrigin string
		keyExpect string
	}{
		{"config/serviceA/name", "/config/serviceA/name"},
		{"config/serviceA/name.subname", "/config/serviceA/name.subname"},
		{"config/serviceA/[dc=dc1]name", "/config/serviceA/name"},
		{"config/serviceA/[dc=dc1,node=node1]name", "/config/serviceA/name"},
		{"config/serviceA/[dc=dc1,node=node1]name.subname", "/config/serviceA/name.subname"},

		{"config/serviceA/[dc=dc1]mysql/[node=node1]name", "/config/serviceA/mysql/name"},
		{"config/serviceA/[dc=dc1,node=node2]mysql/[node=node1]name", "/config/serviceA/mysql/name"},
	} {
		kv := &consul_api.KVPair{
			Key: c.keyOrigin,
		}
		e := newEntry()
		e.decode(kv)
		if e.key != c.keyExpect {
			t.Errorf("[%v] Key '%v'(got)!='%v'(expect)", i, e.key, c.keyExpect)
		}
	}
}

func TestEntryFilter(t *testing.T) {
	for i, c := range []struct {
		properties       map[string]string
		keyOrigin        string
		resultExpect     bool
		matchCountExpect int
	}{
		{map[string]string{}, "config/serviceA/name", true, 0},
		{map[string]string{}, "config/serviceA/[dc=dc1]name", false, 0},

		{map[string]string{"dc": "dc1"}, "config/serviceA/name", true, 0},
		{map[string]string{"dc": "dc1"}, "config/serviceA/[dc=dc1]name", true, 1},
		{map[string]string{"dc": "dc1"}, "config/[dc=dc1]serviceA/name", true, 1},
		{map[string]string{"dc": "dc1"}, "config/serviceA/[dc=dc2]name", false, 0},
		{map[string]string{"dc": "dc1"}, "config/[dc=dc1]serviceA/[dc=dc2]name", false, 0},

		{map[string]string{"dc": "dc1", "node": "node1"}, "config/serviceA/name", true, 0},
		{map[string]string{"dc": "dc1", "node": "node1"}, "config/serviceA/[dc=dc1,node=node1]name", true, 2},
		{map[string]string{"dc": "dc1", "node": "node1"}, "config/serviceA/[dc=dc1]name", true, 1},
		{map[string]string{"dc": "dc1", "node": "node1"}, "config/serviceA/[node=node1]name", true, 1},
		{map[string]string{"dc": "dc1", "node": "node1"}, "config/serviceA/[dc=dc2]name", false, 0},
		{map[string]string{"dc": "dc1", "node": "node1"}, "config/serviceA/[node=node2]name", false, 0},
		{map[string]string{"dc": "dc1", "node": "node1"}, "config/serviceA/[dc=dc2,node=node1]name", false, 0},
		{map[string]string{"dc": "dc1", "node": "node1"}, "config/serviceA/[dc=dc1,node=node2]name", false, 0},

		{map[string]string{"node": "node1"}, "config/serviceA/[node=node1]name", true, 1},
		{map[string]string{"node": "node1"}, "config/serviceA/[dc=dc1]name", false, 0},
		{map[string]string{"node": "node1"}, "config/serviceA/[dc=dc1,node=node1]name", false, 0},
	} {
		kv := &consul_api.KVPair{
			Key: c.keyOrigin,
		}
		e := newEntry()
		e.decode(kv)

		result := e.match(c.properties)
		if result != c.resultExpect {
			t.Errorf("[%v] Result '%v'(got)!='%v'(expect)", i, result, c.resultExpect)
		}

		if c.resultExpect == true {
			if e.filterMatchCount != c.matchCountExpect {
				t.Errorf("[%v] MatchCount '%v'(got)!='%v'(expect)", i, e.filterMatchCount, c.matchCountExpect)
			}
		}
	}
}
