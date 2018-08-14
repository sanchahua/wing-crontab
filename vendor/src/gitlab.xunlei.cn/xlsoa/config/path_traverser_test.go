package config

import (
	"reflect"
	"testing"
)

func TestPathTraverser(t *testing.T) {

	type testEntry struct {
		key   string
		value interface{}
	}
	for i, c := range []struct {
		entries []testEntry
		expect  map[string]interface{}
	}{

		{
			[]testEntry{},
			map[string]interface{}{},
		},
		{
			[]testEntry{
				{"name", "supergui"},
			},
			map[string]interface{}{
				"name": "supergui",
			},
		},
		{
			[]testEntry{
				{"name", "supergui"},
				{"name", "latter will be used"},
			},
			map[string]interface{}{
				"name": "latter will be used",
			},
		},
		{
			[]testEntry{
				{"name", "supergui"},
				{"mysql/user", "root"},
				{"mysql/port", 3306},
			},
			map[string]interface{}{
				"name": "supergui",
				"mysql": map[string]interface{}{
					"user": "root",
					"port": 3306,
				},
			},
		},
		{
			[]testEntry{
				{"name", "supergui"},
				{"mysql/user", "root"},
				{"mysql/port", 3306},
				{"token/create/on", true},
				{"token/create/ttl", 3600},
			},
			map[string]interface{}{
				"name": "supergui",
				"mysql": map[string]interface{}{
					"user": "root",
					"port": 3306,
				},
				"token": map[string]interface{}{
					"create": map[string]interface{}{
						"on":  true,
						"ttl": 3600,
					},
				},
			},
		},
		{
			[]testEntry{
				{"mysql", "Shorter entry will be ignored"},
				{"name", "supergui"},
				{"mysql/user", "root"},
				{"mysql/port", 3306},
			},
			map[string]interface{}{
				"name": "supergui",
				"mysql": map[string]interface{}{
					"user": "root",
					"port": 3306,
				},
			},
		},
		{
			[]testEntry{
				{"name", "supergui"},
				{"mysql/user", "root"},
				{"mysql/port", 3306},
				{"mysql", "Shorter entry will be ignored"},
			},
			map[string]interface{}{
				"name": "supergui",
				"mysql": map[string]interface{}{
					"user": "root",
					"port": 3306,
				},
			},
		},
	} {

		traverser := newPathTraverser()
		for _, entry := range c.entries {
			traverser.digest(entry.key, entry.value)
		}

		if !reflect.DeepEqual(traverser.get(), c.expect) {
			t.Errorf("[case %v] '%v'(got)!='%v'(expect)", i, traverser.get(), c.expect)
		}
	}
}
