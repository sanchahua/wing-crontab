package config

import (
	"testing"
)

func TestFilter(t *testing.T) {

	for i, c := range []struct {
		filterKey   string
		filterValue string
		properties  map[string]string
		expectMatch bool
	}{
		{"dc", "dc1", map[string]string{}, false},
		{"dc", "dc1", map[string]string{"dc_not_match": "dc1"}, false},
		{"dc", "dc1", map[string]string{"dc": "dc1_not_match"}, false},
		{"dc", "dc1", map[string]string{"dc": "dc1"}, true},
		{"dc", "dc1", map[string]string{"dc": "dc1", "node": "node1"}, true},
	} {

		filter := newFilter(c.filterKey, c.filterValue)
		match := filter.match(c.properties)
		if match != c.expectMatch {
			t.Errorf("[%v case] Match '%v'(got)!='%v'(expect)\n", i, match, c.expectMatch)
		}
	}
}
