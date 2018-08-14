// Note:
//     Section-unspecfied items will be adopted by 'GLOBAL' section.
//
package configure

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

const (
	KB = 1024
	MB = 1024 * KB
	GB = 1024 * MB
)

type sectionContainer map[string]string

type Configure struct {
	path string
	m    map[string]*sectionContainer
}

func New(path string) *Configure {
	c := &Configure{path: path}
	/*
		err := c.Reload()
		if err != nil {
			return nil, err
		}*/
	return c
}

func (c *Configure) Load() error {

	f, err := os.Open(c.path)
	if err != nil {
		return err
	}
	defer f.Close()

	//Global section
	c.m = make(map[string]*sectionContainer)
	c.m["GLOBAL"] = &sectionContainer{}

	// Scan lines
	var curSection *sectionContainer = c.m["GLOBAL"]
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		var line string
		var pos int
		var sectionName string

		line = strings.Trim(scanner.Text(), "\r\n ")

		//Comment
		pos = strings.Index(line, "#")
		if pos >= 0 {
			line = line[0:pos]
		}

		//Empty line
		if len(line) == 0 {
			continue
		}

		//Section
		if line[0] == '[' {
			pos = strings.Index(line, "]")
			if pos > 1 {
				sectionName = line[1:pos]

				//Create and switch
				_, ok := c.m[sectionName]
				if ok == false {
					c.m[sectionName] = &sectionContainer{}
				}
				curSection = c.m[sectionName]
			}
		}

		//Key-value
		kv := strings.Split(line, "=")
		if len(kv) != 2 {
			continue
		}
		key := strings.Trim(kv[0], "\t\n ")
		value := strings.Trim(kv[1], "\t\n ")

		(*curSection)[key] = value
	}
	return nil
}

// Check configuration key of section 'section'.
// The second return will be true if exists, the first return holds the string value.
// Else return false, the first return value should be meaningless.
func (c *Configure) Check(section string, key string) (string, bool) {
	var ok bool
	var cur *sectionContainer
	var v string

	cur, ok = c.m[section]
	if ok == false {
		return "", false
	}

	v, ok = (*cur)[key]
	if ok == false {
		return "", false
	}
	return v, true
}

func (c *Configure) GetString(section string, key string, vdefault string) string {
	v, ok := c.Check(section, key)
	if ok == false {
		return vdefault
	}
	return v
}

func (c *Configure) GetInt32(section string, key string, vdefault int32) int32 {
	v, ok := c.Check(section, key)
	if ok == false {
		return vdefault
	}
	i, err := strconv.ParseInt(v, 10, 32)
	if err != nil {
		return vdefault
	}
	return int32(i)
}

func (c *Configure) GetInt64(section string, key string, vdefault int64) int64 {
	v, ok := c.Check(section, key)
	if ok == false {
		return vdefault
	}
	i, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return vdefault
	}
	return i
}
