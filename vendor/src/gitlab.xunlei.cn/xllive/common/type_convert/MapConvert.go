package type_convert

import (
	"time"
	"fmt"
)

func MII2b(m map[interface{}]interface{}, key string, defaultValue bool) bool {
	if m == nil {
		return defaultValue
	}
	interfaceValue, ok := m[key]
	if ok == false {
		return defaultValue
	}
	return I2b(interfaceValue, defaultValue)
}

func Msi2b(m map[string]interface{}, key string, defaultValue bool) bool {
	if m == nil {
		return defaultValue
	}
	interfaceValue, ok := m[key]
	if ok == false {
		return defaultValue
	}
	return I2b(interfaceValue, defaultValue)
}

func Msi2f64(m map[string]interface{}, key string, defaultValue float64) float64 {
	if m == nil {
		return defaultValue
	}
	i, ok := m[key]
	if ok == false {
		return defaultValue
	}

	return I2f64(i, defaultValue)
}

func Msi2i(m map[string]interface{}, key string, defaultValue int) int {
	if m == nil {
		return defaultValue
	}
	i, ok := m[key]
	if ok == false {
		return defaultValue
	}

	return I2i(i, defaultValue)
}

func Msi2i64(m map[string]interface{}, key string, defaultValue int64) int64 {
	if m == nil {
		return defaultValue
	}
	i, ok := m[key]
	if ok == false {
		return defaultValue
	}

	return I2i64(i, defaultValue)
}

func Msi2u64(m map[string]interface{}, key string, defaultValue uint64) uint64 {
	if m == nil {
		return defaultValue
	}
	i, ok := m[key]
	if ok == false {
		return defaultValue
	}

	return I2u64(i, defaultValue)
}

func Msi2s(m map[string]interface{}, key string, defaultValue string) string {
	if m == nil {
		return defaultValue
	}
	i, ok := m[key]
	if ok == false {
		return defaultValue
	}

	return I2s(i, defaultValue)
}

func Msi2Time(m map[string]interface{}, key string, format string, defaultValue time.Time) (time.Time, error) {
	strValue := Msi2s(m, key, "")
	if strValue == "" {
		return defaultValue, fmt.Errorf("%s not found", key)
	}

	timeValue, err := time.ParseInLocation(format, strValue, time.Local)
	if err != nil {
		return defaultValue, err
	}
	return timeValue, nil
}

func Msi2Msi(m map[string]interface{}, key string, defaultValue map[string]interface{}) map[string]interface{} {
	if m == nil {
		return defaultValue
	}
	interfaceValue, success := m[key]
	if success == false || interfaceValue == nil {
		return defaultValue
	}

	mapValue, success := interfaceValue.(map[string]interface{})
	if success == false {
		return defaultValue
	}

	return mapValue
}