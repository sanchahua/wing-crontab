package type_convert

import "reflect"

var defaultTypeConvert *TypeConvert

func init() {
	defaultTypeConvert, _ = NewTypeConvert('f', -1)
}

func SetDefaultTypeConvert(c *TypeConvert)  {
	defaultTypeConvert = c
}

func ValueTypeName(i interface{}) string {
	return defaultTypeConvert.ValueTypeName(i)
}

func KindName(k reflect.Kind ) string {
	return defaultTypeConvert.KindName(k)
}

func Interface2Bool(i interface{}, defaultValue bool) (bool, error) {
	return defaultTypeConvert.Interface2Bool(i, defaultValue)
}

func Interface2Int64(i interface{}, defaultValue int64) (int64, error) {
	return defaultTypeConvert.Interface2Int64(i, defaultValue)
}

func Interface2Uint64(i interface{}, defaultValue uint64) (uint64, error) {
	return defaultTypeConvert.Interface2Uint64(i, defaultValue)
}

func Interface2Float64(i interface{}, defaultValue float64) (float64, error) {
	return defaultTypeConvert.Interface2Float64(i, defaultValue)
}

func Interface2String(i interface{}, defaultValue string) (string, error) {
	return defaultTypeConvert.Interface2String(i, defaultValue)
}

func I2b(i interface{}, defaultValue bool) bool {
	v, _ := defaultTypeConvert.Interface2Bool(i, defaultValue)
	return v
}

func I2i(i interface{}, defaultValue int) int {
	v, _ := defaultTypeConvert.Interface2Int64(i, int64(defaultValue))
	return int(v)
}

func I2i64(i interface{}, defaultValue int64) int64 {
	v, _ := defaultTypeConvert.Interface2Int64(i, defaultValue)
	return v
}

func I2u64(i interface{}, defaultValue uint64) uint64 {
	v, _ := defaultTypeConvert.Interface2Uint64(i, defaultValue)
	return v
}

func I2f64(i interface{}, defaultValue float64) float64 {
	v, _ := defaultTypeConvert.Interface2Float64(i, defaultValue)
	return v
}

func I2s(i interface{}, defaultValue string) string {
	v, _ := defaultTypeConvert.Interface2String(i, defaultValue)
	return v
}