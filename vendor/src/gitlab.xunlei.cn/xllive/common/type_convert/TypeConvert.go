package type_convert

import (
	"reflect"
	"strconv"
	"fmt"
	"errors"
)

var (
	ErrInvalid = errors.New("Invalid")
	ErrNotSupport = errors.New("not support")
	ErrNotImplement = errors.New("not implement")
)

type TypeConvert struct {
	floatFormatFmt byte   /// 详细的参数说明见 strconv.FormatFloat
	floatFormatPrec int
}

func NewTypeConvert(fmt byte, prec int) (*TypeConvert, error) {
	c := new(TypeConvert)
	c.SetFloatFormat(fmt, prec)
	return c, nil
}

func (c *TypeConvert)SetFloatFormat(fmt byte, prec int) {
	c.floatFormatFmt = fmt
	c.floatFormatPrec = prec
}

func (c *TypeConvert)ValueTypeName(i interface{}) string {
	v := reflect.ValueOf(i)
	return c.KindName(v.Kind())
}

func (c *TypeConvert)KindName(k reflect.Kind ) string {
	switch k {
	case reflect.Invalid:
		return "Invalid"
	case reflect.Bool:
		return "Bool"
	case reflect.Int:
		return "Int"
	case reflect.Int8:
		return "Int8"
	case reflect.Int16:
		return "Int16"
	case reflect.Int32:
		return "Int32"
	case reflect.Int64:
		return "Int64"
	case reflect.Uint:
		return "Uint"
	case reflect.Uint8:
		return "Uint8"
	case reflect.Uint16:
		return "Uint16"
	case reflect.Uint32:
		return "Uint32"
	case reflect.Uint64:
		return "Uint64"
	case reflect.Uintptr:
		return "Uintptr"
	case reflect.Float32:
		return "Float32"
	case reflect.Float64:
		return "Float64"
	case reflect.Complex64:
		return "Complex64"
	case reflect.Complex128:
		return "Complex128"
	case reflect.Array:
		return "Array"
	case reflect.Chan:
		return "Chan"
	case reflect.Func:
		return "Func"
	case reflect.Interface:
		return "Interface"
	case reflect.Map:
		return "Map"
	case reflect.Ptr:
		return "Ptr"
	case reflect.Slice:
		return "Slice"
	case reflect.String:
		return "String"
	case reflect.Struct:
		return "Struct"
	case reflect.UnsafePointer:
		return "UnsafePointer"
	default:
		return fmt.Sprintf("not support kind='%d'", k)
	}
}

func (c *TypeConvert)interface2Error(i interface{}) error {
	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.Invalid:
		return ErrInvalid
	case reflect.Uintptr, reflect.Ptr, reflect.UnsafePointer,
		reflect.Complex64, reflect.Complex128,
		reflect.Array, reflect.Map, reflect.Slice, reflect.Chan,
		reflect.Interface, reflect.Struct,
		reflect.Func:
		return fmt.Errorf("not implement v='%d-%s-%v'", v.Kind(), c.KindName(v.Kind()), v)
	default:
		return fmt.Errorf("not support v='%d-%s-%v'", v.Kind(), c.KindName(v.Kind()), v)
	}
}

func (c *TypeConvert)Interface2Bool(i interface{}, defaultValue bool) (bool, error) {
	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.Bool:
		return v.Bool(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if 0 == v.Int() {
			return false, nil
		} else {
			return true, nil
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if 0 == v.Uint() {
			return false, nil
		} else {
			return true, nil
		}
	case reflect.Float32, reflect.Float64:
		if 0 == v.Float() {
			return false, nil
		} else {
			return true, nil
		}
	case reflect.String:
		if v.String() == "" {
			return false, nil
		} else {
			return true, nil
		}
	default:
			return defaultValue, c.interface2Error(i)
	}
}

func (c *TypeConvert)Interface2Int64(i interface{}, defaultValue int64) (int64, error) {
	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.Bool:
		if v.Bool() == true {
			return 1, nil
		} else {
			return 0, nil
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int(), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return int64(v.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return int64(v.Float()), nil
	case reflect.String: {
		if i64, err := strconv.ParseInt(v.String(), 10, 64); err != nil {
			return defaultValue, err
		} else {
			return i64, nil
		}
	}
	default:
		return defaultValue, c.interface2Error(i)
	}
}

func (c *TypeConvert)Interface2Uint64(i interface{}, defaultValue uint64) (uint64, error) {
	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.Bool:
		if v.Bool() == true {
			return 1, nil
		} else {
			return 0, nil
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return uint64(v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint(), nil
	case reflect.Float32, reflect.Float64:
		return uint64(v.Float()), nil
	case reflect.String: {
		if u64, err := strconv.ParseUint(v.String(), 10, 64); err != nil {
			return defaultValue, err
		} else {
			return u64, nil
		}
	}
	default:
		return defaultValue, c.interface2Error(i)
	}
}

func (c *TypeConvert)Interface2Float64(i interface{}, defaultValue float64) (float64, error) {
	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.Bool:
		if v.Bool() == true {
			return 1, nil
		} else {
			return 0, nil
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(v.Uint()), nil
	case reflect.Float32, reflect.Float64:
		return v.Float(), nil
	case reflect.String: {
		if f64, err := strconv.ParseFloat(v.String(), 64); err != nil {
			return defaultValue, err
		} else {
			return f64, nil
		}
	}
	default:
		return defaultValue, c.interface2Error(i)
	}
}

func (c *TypeConvert)Interface2String(i interface{}, defaultValue string) (string, error) {
	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.Bool:
		return strconv.FormatBool(v.Bool()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10), nil
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(v.Float(), c.floatFormatFmt, c.floatFormatPrec, 64), nil
	case reflect.String:
		return v.String(), nil
	default:
		return defaultValue, c.interface2Error(i)
	}
}

