package configure

import (
	"os"
	"testing"
)

const CONTENT_1 = `
id=1234
name = supergui
 email = supergui@live.cn
#sex=male
hobby = football #comment
ts = 147418950393

[SECTION1]
id=1234
name = supergui
 email = supergui@live.cn
#sex=male
hobby = football #comment

[SECTION2]
social = facebook
coding = github

[GLOBAL]
height = 183

#[SECTION3]
weight = 65kg  # GLOBAL item
`

const CONTENT_2 = `
[SECTION]
name = supergui
`

const CONTENT_3 = `
[SECTION]
name = supergui123

[SECTION1]
name = guigui
`

func writeContent(path string, content string) {
	f, err := os.OpenFile(".tmp", os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}
	f.Write([]byte(content))
	f.Close()
}

func cleanFile(path string) {
	os.Remove(path)
}

func TestCheck(t *testing.T) {
	writeContent(".tmp", CONTENT_1)

	c := New(".tmp")
	err := c.Load()
	if err != nil {
		t.Error(err)
	}

	for _, st := range []struct {
		section, key, expected string
		check                  bool
	}{
		// Implicit 'GLOBAL'
		{section: "GLOBAL", key: "id", expected: "1234", check: true},
		{section: "GLOBAL", key: "name", expected: "supergui", check: true},
		{section: "GLOBAL", key: "email", expected: "supergui@live.cn", check: true},
		{section: "GLOBAL", key: "sex", expected: "", check: false},
		{section: "GLOBAL", key: "hobby", expected: "football", check: true},

		// Explicit  GLOBAL'
		{section: "GLOBAL", key: "height", expected: "183", check: true},
		{section: "GLOBAL", key: "weight", expected: "65kg", check: true},

		// SECTION1
		{section: "SECTION1", key: "id", expected: "1234", check: true},
		{section: "SECTION1", key: "name", expected: "supergui", check: true},
		{section: "SECTION1", key: "email", expected: "supergui@live.cn", check: true},
		{section: "SECTION1", key: "sex", expected: "", check: false},
		{section: "SECTION1", key: "hobby", expected: "football", check: true},

		// SECTION2
		{section: "SECTION2", key: "social", expected: "facebook", check: true},
		{section: "SECTION2", key: "coding", expected: "github", check: true},
	} {
		v, ok := c.Check(st.section, st.key)
		if st.check == true {
			if ok == false {
				t.Fatalf("Check fail, not exists. Section '%v', key '%v'.", st.section, st.key)
			}
			if v != st.expected {
				t.Fatalf("Check fail, expected unexpected. Section '%v', key '%v'. '%v'!=(expected)'%v'", st.section, st.key, v, st.expected)
			}
		} else {
			if ok == true {
				t.Fatalf("Check fail, commented key exists. Section '%v', key '%v'", st.section, st.key)
			}
		}
	}

	cleanFile(".tmp")
}

func TestGet(t *testing.T) {
	writeContent(".tmp", CONTENT_1)

	c := New(".tmp")
	err := c.Load()
	if err != nil {
		t.Error(err)
	}

	for _, st := range []struct {
		section, key         string
		expected, useDefault interface{}
		t                    string
	}{
		{section: "GLOBAL", key: "id", expected: int32(1234), useDefault: int32(0), t: "int32"},
		{section: "GLOBAL", key: "id1", expected: int32(99), useDefault: int32(99), t: "int32"}, //Default int32
		{section: "GLOBAL", key: "name", expected: "supergui", useDefault: "", t: "string"},
		{section: "GLOBAL", key: "email", expected: "supergui@live.cn", useDefault: "", t: "string"},
		{section: "GLOBAL", key: "sex", expected: "DEFAULT", useDefault: "DEFAULT", t: "string"}, //Default string
		{section: "GLOBAL", key: "hobby", expected: "football", useDefault: "", t: "string"},
		{section: "GLOBAL", key: "ts", expected: int64(147418950393), useDefault: int64(0), t: "int64"},

		// SECTION1
		{section: "SECTION1", key: "id", expected: int32(1234), useDefault: int32(0), t: "int32"},
		{section: "SECTION1", key: "id1", expected: int32(99), useDefault: int32(99), t: "int32"}, //Default int32
		{section: "SECTION1", key: "name", expected: "supergui", useDefault: "", t: "string"},
		{section: "SECTION1", key: "email", expected: "supergui@live.cn", useDefault: "", t: "string"},
		{section: "SECTION1", key: "sex", expected: "DEFAULT", useDefault: "DEFAULT", t: "string"}, //Default string
		{section: "SECTION1", key: "hobby", expected: "football", useDefault: "", t: "string"},
	} {
		if st.t == "string" {
			v := c.GetString(st.section, st.key, st.useDefault.(string))
			if v != st.expected {
				t.Fatalf("Check fail, expected unexpected. Section '%v', key '%v'. '%v'!=(expected)'%v'", st.section, st.key, v, st.expected)
			}

		} else if st.t == "int32" {
			v := c.GetInt32(st.section, st.key, st.useDefault.(int32))
			if v != st.expected {
				t.Fatalf("Check fail, expected unexpected. Section '%v', key '%v'. '%v'!=(expected)'%v'", st.section, st.key, v, st.expected)
			}
		} else if st.t == "int64" {
			v := c.GetInt64(st.section, st.key, st.useDefault.(int64))
			if v != st.expected {
				t.Fatalf("Check fail, expected unexpected. Section '%v', key '%v'. '%v'!=(expected)'%v'", st.section, st.key, v, st.expected)
			}
		}
	}
	cleanFile(".tmp")
}

func TestReload(t *testing.T) {
	writeContent(".tmp", CONTENT_2)

	c := New(".tmp")
	err := c.Load()
	if err != nil {
		t.Error(err)
	}

	for _, st := range []struct {
		section, key string
		expected     string
	}{
		{section: "SECTION", key: "name", expected: "supergui"},
	} {
		v := c.GetString(st.section, st.key, "")
		if v != st.expected {
			t.Fatalf("Check fail, expected unexpected. Section '%v', key '%v'. '%v'!=(expected)'%v'", st.section, st.key, v, st.expected)
		}
	}

	// Flush && Reload
	writeContent(".tmp", CONTENT_3)
	err = c.Load()
	if err != nil {
		t.Fatal(err)
	}

	for _, st := range []struct {
		section, key string
		expected     string
	}{
		{section: "SECTION", key: "name", expected: "supergui123"},
		{section: "SECTION1", key: "name", expected: "guigui"},
	} {
		v := c.GetString(st.section, st.key, "")
		if v != st.expected {
			t.Fatalf("Check fail, expected unexpected. Section '%v', key '%v'. '%v'!=(expected)'%v'", st.section, st.key, v, st.expected)
		}
	}
	cleanFile(".tmp")
}
