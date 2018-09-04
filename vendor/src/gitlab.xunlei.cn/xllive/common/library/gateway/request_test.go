package gateway

import (
	"fmt"
	"testing"
	"time"
)

func TestCbinRequest(t *testing.T) {
	var data map[string]interface{} = map[string]interface{}{
		"request":  "getuserinfo_base",
		"userid":   252655038,
		"usertype": 2,
	}
	d, err := CbinRequest("gateway1.reg", 8636, data, 2*time.Second)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v", d)
	data = map[string]interface{}{
		"request":  "name2id",
		"userid":   100987,
		"usertype": -1,
	}
	d, err = CbinRequest("gateway1.reg", 8636, data, 2*time.Second)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%v", d)
}
