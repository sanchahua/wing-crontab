package config

import (
	"errors"
	"fmt"
)

type Value struct {
	//node *yamlNode
	tree *yamlTree
}

func newValue(tree *yamlTree) *Value {
	return &Value{
		tree: tree,
	}
}

func (v *Value) String() string {
	return fmt.Sprintf("Value{ tree: %v }", v.tree)
}

func (v *Value) Populate(out interface{}) error {
	return v.tree.populate(out)
}

func (v *Value) ToYamlByte() ([]byte, error) {
	return v.tree.toYamlByte()
}

// Panic if can't convert
func (v *Value) AsString() string {
	s, ok := v.TryAsString()
	if !ok {
		panic(errors.New(fmt.Sprintf("Can't convert string: '%v'", v)))
	}
	return s
}

func (v *Value) TryAsString() (string, bool) {
	var s string
	err := v.Populate(&s)
	if err != nil {
		return "", false
	}
	return s, true
}

// Panic if can't convert
func (v *Value) AsInt() int {
	i, ok := v.TryAsInt()
	if !ok {
		panic(errors.New(fmt.Sprintf("Can't convert int: '%v'", v)))
	}
	return i
}
func (v *Value) TryAsInt() (int, bool) {
	var i int
	err := v.Populate(&i)
	if err != nil {
		return 0, false
	}
	return i, true
}

// Panic if can't convert
func (v *Value) AsFloat() float64 {
	f, ok := v.TryAsFloat()
	if !ok {
		panic(errors.New(fmt.Sprintf("Can't convert float64: '%v'", v)))
	}
	return f
}
func (v *Value) TryAsFloat() (float64, bool) {
	var f float64
	err := v.Populate(&f)
	if err != nil {
		return 0, false
	}
	return f, true
}

// Panic if can't convert
func (v *Value) AsBool() bool {
	b, ok := v.TryAsBool()
	if !ok {
		panic(errors.New(fmt.Sprintf("Can't convert bool: '%v'", v)))
	}
	return b
}
func (v *Value) TryAsBool() (bool, bool) {
	var b bool
	err := v.Populate(&b)
	if err != nil {
		return false, false
	}
	return b, true
}
