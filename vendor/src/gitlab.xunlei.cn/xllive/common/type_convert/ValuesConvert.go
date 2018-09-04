package type_convert

import (
	"net/url"
	"fmt"
	"strconv"
	"errors"
)

func Values2s(v url.Values, key string, defaultValue string) (string, error) {
	strValue, ok := v[key]
	if ok != true {
		return defaultValue, errors.New(fmt.Sprintf("key=[%s] not exist", key))
	}

	if strValue == nil || len(strValue) == 0 {
		return defaultValue, errors.New(fmt.Sprintf("key=[%s] strValue empty", key))
	}

	return strValue[0], nil
}

func Values2i64(v url.Values, key string, defaultValue int64) (int64, error) {

	strValue, ok := v[key]
	if ok != true {
		return defaultValue, errors.New(fmt.Sprintf("key=[%s] not exist", key))
	}

	int64Value, err := strconv.ParseInt(strValue[0], 10, 64)
	if err != nil {
		return defaultValue, err
	}

	return int64Value, nil
}

