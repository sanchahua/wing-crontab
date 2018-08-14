package utility

import (
	"testing"
)

func TestPathPattern(t *testing.T) {

	for _, c := range []struct {
		input   string
		depth   int
		level   []string
		format0 string
		format1 string
	}{
		{"", 0, []string{}, "", ""},
		{"test.server", 1, []string{"test.server"}, "/test.server", ""},
		{"/test.server", 1, []string{"test.server"}, "/test.server", ""},
		{"//test.server", 1, []string{"test.server"}, "/test.server", ""},
		{"/test.server/", 1, []string{"test.server"}, "/test.server", ""},
		{"//test.server///", 1, []string{"test.server"}, "/test.server", ""},
		{"/test.server/api1", 2, []string{"test.server", "api1"}, "/test.server/api1", "/api1"},
		{"///test.server///api1/", 2, []string{"test.server", "api1"}, "/test.server/api1", "/api1"},
		{"/test.server/api1//call1", 3, []string{"test.server", "api1", "call1"}, "/test.server/api1/call1", "/api1/call1"},
	} {
		p := NewPathPattern(c.input)
		if p.Depth() != c.depth {
			t.Errorf("Depth error. Input '%v', '%v'(Result)!='%v'(Wanted)", c.input, p.Depth(), c.depth)
			continue
		}

		for i, v := range c.level {
			if v != p.LevelName(i) {
				t.Errorf("Level mismatchedi. Input '%v'. Index %v. '%v'(Result)!='%v'(Wanted)", c.input, i, p.LevelName(i), v)
			}
		}

		if p.LevelName(p.Depth()) != "" {
			t.Errorf("Outbound level should be empty. Input '%v'. Index %v. '%v'!=''(Wanted)", c.input, p.Depth(), p.LevelName(p.Depth()))
		}

		if p.Format(0) != c.format0 {
			t.Errorf("Format0 unexpected. Input '%v'. '%v'(Result)!='%v'(Wanted)", c.input, p.Format(0), c.format0)
		}
		if p.Format(1) != c.format1 {
			t.Errorf("Format1 unexpected. Input '%v'. '%v'(Result)!='%v'(Wanted)", c.input, p.Format(1), c.format1)
		}
	}
}
