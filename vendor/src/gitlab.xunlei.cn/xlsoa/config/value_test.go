package config

import (
	"gopkg.in/yaml.v2"
	//"log"
	"reflect"
	"testing"
)

// string
var testValueSampleString = `
supergui
`

// int || string
var testValueSampleInt = `
123
`

// string
var testValueSampleQuotedInt = `
"123"
`

// float || string
var testValueSampleFloat = `
123.123
`

// bool || string
var testValueSampleBool = `
True
`

// Array
var testValueSampleArray = `
- supergui
- guigui
`

// Object
var testValueSampleObject = `
name: supergui
mysql:
    user: root
    host: localhost
    port: 3306
    rate: 123.123
token:
    create:
        ttl: 3600
        open: true
`

func createYamlTree(t *testing.T, data []byte) *yamlTree {
	var out interface{}
	err := yaml.Unmarshal(data, &out)
	if err != nil {
		t.Fatalf("createYamlTree error: %v\n", err)
	}

	node := newYamlNode("root", out)
	tree := newYamlTreeFromYamlNode(node)
	tree.init()
	return tree

}

func assertValueString(t *testing.T, v *Value, expect string) {
	s, ok := v.TryAsString()
	if !ok {
		t.Error("TryAsString fail")
	} else if s != expect {
		t.Errorf("TryAsString value '%v'(got)!='%v'(expect)\n", s, expect)
	}
}
func assertValueNotString(t *testing.T, v *Value) {
	_, ok := v.TryAsString()
	if ok {
		t.Errorf("TryAsString success, expecting fail. Value '%v'\n", v)
	}
}

func assertValueInt(t *testing.T, v *Value, expect int) {
	i, ok := v.TryAsInt()
	if !ok {
		t.Error("TryAsInt fail")
	} else if i != expect {
		t.Errorf("TryAsInt value '%v'(got)!='%v'(expect)\n", i, expect)
	}
}
func assertValueNotInt(t *testing.T, v *Value) {

	_, ok := v.TryAsInt()
	if ok {
		t.Errorf("TryAsInt success, expecting fail. Value '%v'\n", v)
	}
}

func assertValueFloat(t *testing.T, v *Value, expect float64) {
	f, ok := v.TryAsFloat()
	if !ok {
		t.Error("TryAsFloat fail")
	} else if f != expect {
		t.Errorf("TryAsFloat value '%v'(got)!='%v'(expect)\n", f, expect)
	}
}
func assertValueNotFloat(t *testing.T, v *Value) {

	_, ok := v.TryAsFloat()
	if ok {
		t.Errorf("TryAsFloat success, expecting fail. Value '%v'\n", v)
	}
}

func assertValueBool(t *testing.T, v *Value, expect bool) {
	b, ok := v.TryAsBool()
	if !ok {
		t.Error("TryAsBool fail")
	} else if b != expect {
		t.Errorf("TryAsBool value '%v'(got)!='%v'(expect)\n", b, expect)
	}
}
func assertValueNotBool(t *testing.T, v *Value) {

	_, ok := v.TryAsBool()
	if ok {
		t.Errorf("TryAsBool success, expecting fail. Value '%v'\n", v)
	}
}

func TestValueSampleString(t *testing.T) {

	v := newValue(createYamlTree(t, []byte(testValueSampleString)))
	assertValueNotInt(t, v)
	assertValueString(t, v, "supergui")
	assertValueNotFloat(t, v)
	assertValueNotBool(t, v)

}
func TestValueSampleInt(t *testing.T) {

	v := newValue(createYamlTree(t, []byte(testValueSampleInt)))
	assertValueInt(t, v, 123)
	assertValueString(t, v, "123")
	assertValueFloat(t, v, 123)
	assertValueNotBool(t, v)

}
func TestValueSampleQuotedInt(t *testing.T) {

	v := newValue(createYamlTree(t, []byte(testValueSampleQuotedInt)))
	assertValueNotInt(t, v)
	assertValueString(t, v, "123")
	assertValueNotFloat(t, v)
	assertValueNotBool(t, v)

}
func TestValueSampleFloat(t *testing.T) {

	v := newValue(createYamlTree(t, []byte(testValueSampleFloat)))
	assertValueInt(t, v, 123)
	assertValueString(t, v, "123.123")
	assertValueFloat(t, v, 123.123)
	assertValueNotBool(t, v)

}
func TestValueSampleBool(t *testing.T) {

	v := newValue(createYamlTree(t, []byte(testValueSampleBool)))
	assertValueNotInt(t, v)
	assertValueString(t, v, "true")
	assertValueNotFloat(t, v)
	assertValueBool(t, v, true)

}

func TestValueSampleArray(t *testing.T) {

	v := newValue(createYamlTree(t, []byte(testValueSampleArray)))
	assertValueNotInt(t, v)
	assertValueNotString(t, v)
	assertValueNotFloat(t, v)
	assertValueNotBool(t, v)

	var out []string
	var expect = []string{"supergui", "guigui"}

	err := v.Populate(&out)
	if err != nil {
		t.Fatal("Populate error")
	}

	if !reflect.DeepEqual(out, expect) {
		t.Errorf("Value '%v'(out)!='%v'(expect)\n", out, expect)
	}
}

func TestValueSampleObject(t *testing.T) {

	v := newValue(createYamlTree(t, []byte(testValueSampleObject)))
	assertValueNotInt(t, v)
	assertValueNotString(t, v)
	assertValueNotFloat(t, v)
	assertValueNotBool(t, v)

	type testObj struct {
		Name  string `yaml:"name"`
		Mysql struct {
			Host string  `yaml:"host"`
			User string  `yaml:"user"`
			Port int     `yaml:"port"`
			Rate float64 `yaml:"rate"`
		} `yaml:"mysql"`
		Token struct {
			Create struct {
				Ttl  int  `yaml:"ttl"`
				Open bool `yaml:"open"`
			} `yaml:"create"`
		} `yaml:"token"`
	}
	var out = testObj{}

	var expect = testObj{
		"supergui",
		struct {
			Host string  `yaml:"host"`
			User string  `yaml:"user"`
			Port int     `yaml:"port"`
			Rate float64 `yaml:"rate"`
		}{
			"localhost",
			"root",
			3306,
			123.123,
		},
		struct {
			Create struct {
				Ttl  int  `yaml:"ttl"`
				Open bool `yaml:"open"`
			} `yaml:"create"`
		}{
			struct {
				Ttl  int  `yaml:"ttl"`
				Open bool `yaml:"open"`
			}{
				3600,
				true,
			},
		},
	}

	err := v.Populate(&out)
	if err != nil {
		t.Fatal("Populate error")
	}

	if !reflect.DeepEqual(out, expect) {
		t.Errorf("Value '%v'(out)!='%v'(expect)\n", out, expect)
	}
}
