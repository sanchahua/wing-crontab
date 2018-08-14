package config

import (
	"reflect"
	"testing"
)

func TestParseProperties(t *testing.T) {

	for i, c := range []struct {
		in          string
		expectName  string
		expectProps map[string]string
	}{
		{"", "", map[string]string{}},
		{"[]", "", map[string]string{}},
		{"name", "name", map[string]string{}},
		{"name.subname", "name.subname", map[string]string{}},
		{"  name.subname  ", "name.subname", map[string]string{}},
		{"[dc=tw06]", "", map[string]string{"dc": "tw06"}},
		{"[dc=tw06]name", "name", map[string]string{"dc": "tw06"}},
		{"[dc=tw06name", "[dc=tw06name", map[string]string{}},
		{"]dc=tw06name[", "]dc=tw06name[", map[string]string{}},
		{"[dc=tw06]name.subname", "name.subname", map[string]string{"dc": "tw06"}},
		{"  [dc=tw06]  name  ", "name", map[string]string{"dc": "tw06"}},
		{"  [dc=tw06]  name.subname  ", "name.subname", map[string]string{"dc": "tw06"}},
		{"  [  dc  =  tw06  ]  name  ", "name", map[string]string{"dc": "tw06"}},
		{"[dc=tw06,node=tw06249]  name", "name", map[string]string{"dc": "tw06", "node": "tw06249"}},
		{"  [dc=tw06,node=tw06249]name  ", "name", map[string]string{"dc": "tw06", "node": "tw06249"}},
		{"  [  dc  =  tw06,node=tw06249] name  ", "name", map[string]string{"dc": "tw06", "node": "tw06249"}},
		{"  [  dc  =  tw06,  node  =  tw06249  ]name  ", "name", map[string]string{"dc": "tw06", "node": "tw06249"}},
		{"[dc=tw06,wrong prop]name", "name", map[string]string{"dc": "tw06"}},
		{"[dc=tw06,wrong prop,  ,, node = tw06249]name", "name", map[string]string{"dc": "tw06", "node": "tw06249"}},
	} {

		name, props := parseProperties(c.in)
		if name != c.expectName {
			t.Errorf("[%v] Name '%v'(got)!='%v'(expect)", i, name, c.expectName)
		}
		if !reflect.DeepEqual(props, c.expectProps) {
			t.Errorf("[%v] Props %v(got)!=%v(expect)", i, props, c.expectProps)
		}
	}
}

func TestConvertValue(t *testing.T) {

	for i, c := range []struct {
		s      string
		expect interface{}
	}{
		{"123", int64(123)},
		{"-123", int64(-123)},
		{"123.123", 123.123},
		{"-123.123", -123.123},

		{"true", true},
		{"True", true},
		{"TRUE", true},
		{"TRue", "TRue"}, //Invalid bool string
		{"TrUe", "TrUe"}, // Invalid bool string

		{"false", false},
		{"False", false},
		{"False", false},
		{"FAlse", "FAlse"}, //Invalid bool string
		{"FaLse", "FaLse"}, // Invalid bool string

		{"a123", "a123"},
		{"123a", "123a"},
		{"supergui", "supergui"},
	} {
		v := convertValue(c.s)
		if v != c.expect {
			t.Errorf("[case %v] '%v'(type %v)(got)!='%v'(type %v)(expect)\n", i, v, reflect.TypeOf(v), c.expect, reflect.TypeOf(c.expect))
		}
	}
}
