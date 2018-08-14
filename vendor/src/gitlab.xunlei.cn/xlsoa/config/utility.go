package config

import (
	"gitlab.xunlei.cn/xlsoa/common/utility"
	"strconv"
	"strings"
)

// case1: "name"
// case2: "[dc=xxx, node=xxx, instance=xxx]name"
func parseProperties(in string) (string, map[string]string) {

	in = strings.Trim(in, " ")

	name := in
	props := make(map[string]string)

	// Catch up properties
	if len(in) > 0 && in[0] == '[' {
		pos := strings.IndexByte(in, ']')
		if pos >= 0 {

			name = strings.Trim(in[pos+1:], " ")
			items := strings.Split(in[1:pos], ",")
			for _, item := range items {
				item = strings.Trim(item, " ")
				kv := strings.Split(item, "=")
				if len(kv) != 2 {
					continue
				}

				key := strings.Trim(kv[0], " ")
				value := strings.Trim(kv[1], " ")
				props[key] = value
			}

		} // Endof if pos>=0 ...
	} // Endof if len(in) ...

	return name, props
}

func traversePathToMap(path string, value interface{}) map[string]interface{} {

	sp := utility.NewPathPattern(path)

	m := make(map[string]interface{})
	for i := 0; i < sp.Depth(); i++ {

		s := sp.LevelName(i)

		_, ok := m[s]
		if !ok {

			if i == sp.Depth()-1 {
				// Assign value at leaf
				m[s] = value
			} else {
				// Create child
				m[s] = make(map[string]interface{})
			}

		}

		// Traverse to child
		tmp, ok := m[s].(map[string]interface{})
		if ok {
			m = tmp
		}

	}

	return m

}

// If ParseInt() ok, it is an int64.
// Else if ParseFloat(), it is an float64.
// Else if ParseBool(), it is an bool.
// Else, it is a string.
func convertValue(s string) interface{} {

	// Int
	{
		v, err := strconv.ParseInt(s, 10, 64)
		if err == nil {
			return v
		}
	}

	// Float
	{
		v, err := strconv.ParseFloat(s, 64)
		if err == nil {
			return v
		}
	}

	// Bool
	{
		v, err := strconv.ParseBool(s)
		if err == nil {
			return v
		}
	}

	return s
}
