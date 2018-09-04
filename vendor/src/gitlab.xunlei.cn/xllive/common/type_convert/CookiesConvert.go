package type_convert

import (
	"net/http"
	"fmt"
	"strconv"
)

func Cookies2i64(r *http.Request, key string, defaultValue int64) (int64, error) {

	cookie, err := r.Cookie(key)
	if err != nil {
		return defaultValue, err
	}

	if cookie == nil {
		return defaultValue, fmt.Errorf("%v key not found", key)
	}

	intValue, err := strconv.ParseInt(cookie.Value, 10, 64)
	if err != nil {
		return defaultValue, err
	}

	return intValue, nil
}

func Cookies2s(r *http.Request, key string, defaultValue string) (string, error) {

	cookie, err := r.Cookie(key)
	if err != nil {
		return defaultValue, err
	}

	if cookie == nil {
		return defaultValue, fmt.Errorf("%v key not found", key)
	}

	return cookie.Value, nil
}