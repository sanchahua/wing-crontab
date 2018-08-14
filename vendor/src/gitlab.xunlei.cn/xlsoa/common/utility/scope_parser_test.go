package utility

import (
	"testing"
)

func TestScopeParser(t *testing.T) {

	for _, c := range []struct {
		input        string
		serverWanted string
		methodWanted string
		format       string
	}{
		{"", "", "", "/"},
		{"test.server", "test.server", "", "/test.server"},
		{"/test.server", "test.server", "", "/test.server"},
		{"/test.server/", "test.server", "", "/test.server"},
		{"/test.server/api1", "test.server", "/api1", "/test.server/api1"},
		{"/test.server/api1/api2", "test.server", "/api1/api2", "/test.server/api1/api2"},
		{"//test.server//api1//api2//", "test.server", "/api1/api2", "/test.server/api1/api2"},
	} {
		parser := NewScopeParser(c.input)
		if parser.ServerName != c.serverWanted {
			t.Errorf("ServerName not wanted. Input '%v'. '%v'(Result)!='%v'(Wanted)", c.input, parser.ServerName, c.serverWanted)
		}
		if parser.MethodName != c.methodWanted {
			t.Errorf("Methodname not wanted. Input '%v'. '%v'(Result)!='%v'(Wanted)", c.input, parser.MethodName, c.methodWanted)
		}

		if parser.Format() != c.format {
			t.Errorf("Format not wanted. Input '%v'. '%v'(Result)!='%v'(Wanted)", c.input, parser.Format(), c.format)
		}
	}
}
