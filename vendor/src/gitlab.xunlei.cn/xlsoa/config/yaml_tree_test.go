package config

import (
	"reflect"
	"testing"
)

// string
var testYamlTreeSampleString = `
supergui
`

// int || string
var testYamlTreeSampleInt = `
123
`

// string
var testYamlTreeSampleQuotedInt = `
"123"
`

// float
var testYamlTreeSampleFloat = `
123.123
`

// bool
var testYamlTreeSampleBool = `
True
`

// Array
var testYamlTreeSampleArray = `
- supergui
- guigui
`

// Object
var testYamlTreeSampleObject = `
name: supergui
mysql:
    user: root
    host: localhost
    port: 3306
    rate: 123.123
token:
    create:
        ttl: 3600
        open: True
`

// Object with array
var testYamlTreeSampleObjectWithArray = `
name: 
    - supergui
    - guigui
`

func TestYamlTreeSampleString(t *testing.T) {
	var err error
	var v *Value

	tree := newYamlTreeFromByte([]byte(testYamlTreeSampleString))
	if err = tree.init(); err != nil {
		t.Fatalf("Init tree error: %v\n", err)
	}

	// Get ROOT, should success
	if v, err = tree.get(ROOT); err != nil {
		t.Errorf("Get 'ROOT' error: %v\n", err)
	} else if v == nil {
		t.Errorf("Get 'ROOT' not exists\n")
	}

	assertValueNotInt(t, v)
	assertValueString(t, v, "supergui")
	assertValueNotFloat(t, v)
	assertValueNotBool(t, v)

	// Get anything key, should not exists
	if v, err = tree.get("anything"); err != nil {
		t.Errorf("Get 'anything' error: %v\n", err)
	} else if v != nil {
		t.Errorf("Get 'anything' success, expecting not exists")
	}

}
func TestYamlTreeSampleInt(t *testing.T) {
	var err error
	var v *Value

	tree := newYamlTreeFromByte([]byte(testYamlTreeSampleInt))
	if err = tree.init(); err != nil {
		t.Fatalf("Init tree error: %v\n", err)
	}

	// Get ROOT, should success
	if v, err = tree.get(ROOT); err != nil {
		t.Errorf("Get 'ROOT' error: %v\n", err)
	} else if v == nil {
		t.Errorf("Get 'ROOT' not exists\n")
	}

	assertValueInt(t, v, 123)
	assertValueString(t, v, "123")
	assertValueFloat(t, v, 123)
	assertValueNotBool(t, v)

	// Get anything key, should not exists
	if v, err = tree.get("anything"); err != nil {
		t.Errorf("Get 'anything' error: %v\n", err)
	} else if v != nil {
		t.Errorf("Get 'anything' success, expecting not exists")
	}

}
func TestYamlTreeSampleQuotedInt(t *testing.T) {
	var err error
	var v *Value

	tree := newYamlTreeFromByte([]byte(testYamlTreeSampleQuotedInt))
	if err = tree.init(); err != nil {
		t.Fatalf("Init tree error: %v\n", err)
	}

	// Get ROOT, should success
	if v, err = tree.get(ROOT); err != nil {
		t.Errorf("Get 'ROOT' error: %v\n", err)
	} else if v == nil {
		t.Errorf("Get 'ROOT' not exists\n")
	}

	assertValueNotInt(t, v)
	assertValueString(t, v, "123")
	assertValueNotFloat(t, v)
	assertValueNotBool(t, v)

	// Get anything key, should not exists
	if v, err = tree.get("anything"); err != nil {
		t.Errorf("Get 'anything' error: %v\n", err)
	} else if v != nil {
		t.Errorf("Get 'anything' success, expecting not exists")
	}

}

func TestYamlTreeSampleFloat(t *testing.T) {
	var err error
	var v *Value

	tree := newYamlTreeFromByte([]byte(testYamlTreeSampleFloat))
	if err = tree.init(); err != nil {
		t.Fatalf("Init tree error: %v\n", err)
	}

	// Get ROOT, should success
	if v, err = tree.get(ROOT); err != nil {
		t.Errorf("Get 'ROOT' error: %v\n", err)
	} else if v == nil {
		t.Errorf("Get 'ROOT' not exists\n")
	}

	assertValueInt(t, v, 123)
	assertValueString(t, v, "123.123")
	assertValueFloat(t, v, 123.123)
	assertValueNotBool(t, v)

	// Get anything key, should not exists
	if v, err = tree.get("anything"); err != nil {
		t.Errorf("Get 'anything' error: %v\n", err)
	} else if v != nil {
		t.Errorf("Get 'anything' success, expecting not exists")
	}

}
func TestYamlTreeSampleBool(t *testing.T) {
	var err error
	var v *Value

	tree := newYamlTreeFromByte([]byte(testYamlTreeSampleBool))
	if err = tree.init(); err != nil {
		t.Fatalf("Init tree error: %v\n", err)
	}

	// Get ROOT, should success
	if v, err = tree.get(ROOT); err != nil {
		t.Errorf("Get 'ROOT' error: %v\n", err)
	} else if v == nil {
		t.Errorf("Get 'ROOT' not exists\n")
	}

	assertValueNotInt(t, v)
	assertValueString(t, v, "true")
	assertValueNotFloat(t, v)
	assertValueBool(t, v, true)

	// Get anything key, should not exists
	if v, err = tree.get("anything"); err != nil {
		t.Errorf("Get 'anything' error: %v\n", err)
	} else if v != nil {
		t.Errorf("Get 'anything' success, expecting not exists")
	}

}
func TestYamlTreeSampleArray(t *testing.T) {
	var err error
	var v *Value

	tree := newYamlTreeFromByte([]byte(testYamlTreeSampleArray))
	if err = tree.init(); err != nil {
		t.Fatalf("Init tree error: %v\n", err)
	}

	// Get ROOT, should success
	if v, err = tree.get(ROOT); err != nil {
		t.Errorf("Get 'ROOT' error: %v\n", err)
	} else if v == nil {
		t.Errorf("Get 'ROOT' not exists\n")
	}

	assertValueNotInt(t, v)
	assertValueNotString(t, v)
	assertValueNotFloat(t, v)
	assertValueNotBool(t, v)

	var out []string
	var expect = []string{"supergui", "guigui"}
	err = v.Populate(&out)
	if err != nil {
		t.Error("Populate error")
	} else if !reflect.DeepEqual(out, expect) {
		t.Errorf("Value '%v'(out)!='%v'(expect)\n", out, expect)
	}

	// Get anything key, should not exists
	if v, err = tree.get("anything"); err != nil {
		t.Errorf("Get 'anything' error: %v\n", err)
	} else if v != nil {
		t.Errorf("Get 'anything' success, expecting not exists")
	}

}
func TestYamlTreeSampleObject(t *testing.T) {
	var err error
	var v *Value

	tree := newYamlTreeFromByte([]byte(testYamlTreeSampleObject))
	if err = tree.init(); err != nil {
		t.Fatalf("Init tree error: %v\n", err)
	}

	// Get ROOT, should success
	{
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

		if v, err = tree.get(ROOT); err != nil {
			t.Errorf("Get 'ROOT' error: %v\n", err)
		} else if v == nil {
			t.Errorf("Get 'ROOT' not exists\n")
		} else {

			assertValueNotInt(t, v)
			assertValueNotString(t, v)
			assertValueNotFloat(t, v)
			assertValueNotBool(t, v)

			if err = v.Populate(&out); err != nil {
				t.Error("Populate error")
			} else if !reflect.DeepEqual(out, expect) {
				t.Errorf("Value '%v'(out)!='%v'(expect)\n", out, expect)
			}
		}

	}

	// Get sub-tree
	{
		type testCreateObj struct {
			Ttl  int  `yaml:"ttl"`
			Open bool `yaml:"open"`
		}
		var out = testCreateObj{}
		var expect = testCreateObj{
			3600,
			true,
		}

		if v, err = tree.get("token.create"); err != nil {
			t.Errorf("Get 'token.create' error: %v\n", err)
		} else if v == nil {
			t.Errorf("Get 'token.create' not exists\n")
		} else {

			assertValueNotInt(t, v)
			assertValueNotString(t, v)
			assertValueNotFloat(t, v)
			assertValueNotBool(t, v)

			if err = v.Populate(&out); err != nil {
				t.Error("Populate error")
			} else if !reflect.DeepEqual(out, expect) {
				t.Errorf("Value '%v'(out)!='%v'(expect)\n", out, expect)
			}
		}

	}

	// Get leaf node string
	{
		if v, err = tree.get("mysql.user"); err != nil {
			t.Errorf("Get 'mysql.user' error: %v\n", err)
		} else if v == nil {
			t.Errorf("Get 'mysql.user' not exists\n")
		} else {

			assertValueNotInt(t, v)
			assertValueString(t, v, "root")
			assertValueNotFloat(t, v)
			assertValueNotBool(t, v)
		}
	}

	// Get leaf node int
	{
		if v, err = tree.get("mysql.port"); err != nil {
			t.Errorf("Get 'mysql.port' error: %v\n", err)
		} else if v == nil {
			t.Errorf("Get 'mysql.port' not exists\n")
		} else {

			assertValueInt(t, v, 3306)
			assertValueString(t, v, "3306")
			assertValueFloat(t, v, 3306)
			assertValueNotBool(t, v)
		}
	}

	// Get leaf node float64
	{
		if v, err = tree.get("mysql.rate"); err != nil {
			t.Errorf("Get 'mysql.rate' error: %v\n", err)
		} else if v == nil {
			t.Errorf("Get 'mysql.rate' not exists\n")
		} else {

			assertValueInt(t, v, 123)
			assertValueString(t, v, "123.123")
			assertValueFloat(t, v, 123.123)
			assertValueNotBool(t, v)
		}
	}

	// Get leaf node bool
	{
		if v, err = tree.get("token.create.open"); err != nil {
			t.Errorf("Get 'token.create.open' error: %v\n", err)
		} else if v == nil {
			t.Errorf("Get 'token.create.open' not exists\n")
		} else {

			assertValueNotInt(t, v)
			assertValueString(t, v, "true")
			assertValueNotFloat(t, v)
			assertValueBool(t, v, true)
		}
	}

	// Get none-exisit key, should not exists
	if v, err = tree.get("anything.anything1"); err != nil {
		t.Errorf("Get 'anything' error: %v\n", err)
	} else if v != nil {
		t.Errorf("Get 'anything' success, expecting not exists")
	}

}

func TestYamlTreeSampleObjectWithArray(t *testing.T) {
	var err error
	var v *Value

	tree := newYamlTreeFromByte([]byte(testYamlTreeSampleObjectWithArray))
	if err = tree.init(); err != nil {
		t.Fatalf("Init tree error: %v\n", err)
	}

	// Get root
	{
		type testObj struct {
			Name []string `yaml:"name"`
		}

		var out = testObj{}
		var expect = testObj{
			[]string{"supergui", "guigui"},
		}

		if v, err = tree.get(ROOT); err != nil {
			t.Errorf("Get 'ROOT' error: %v\n", err)
		} else if v == nil {
			t.Errorf("Get 'ROOT' not exists\n")
		} else {

			assertValueNotInt(t, v)
			assertValueNotString(t, v)
			assertValueNotFloat(t, v)
			assertValueNotBool(t, v)

			if err = v.Populate(&out); err != nil {
				t.Error("Populate error")
			} else if !reflect.DeepEqual(out, expect) {
				t.Errorf("Value '%v'(out)!='%v'(expect)\n", out, expect)
			}
		}
	}

	// Get 'name'
	{

		var out []string
		var expect = []string{"supergui", "guigui"}
		if v, err = tree.get("name"); err != nil {
			t.Errorf("Get 'name' error: %v\n", err)
		} else if v == nil {
			t.Errorf("Get 'name' not exists\n")
		} else {

			assertValueNotInt(t, v)
			assertValueNotString(t, v)
			assertValueNotFloat(t, v)
			assertValueNotBool(t, v)

			if err = v.Populate(&out); err != nil {
				t.Error("Populate error")
			} else if !reflect.DeepEqual(out, expect) {
				t.Errorf("Value '%v'(out)!='%v'(expect)\n", out, expect)
			}
		}
	}
}
