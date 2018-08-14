package encoding

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"reflect"
	"testing"
)

var encoder *Encoder
var decoder *Decoder

var base64Encoder *base64.Encoding

func init() {
	encoder = &Encoder{}
	decoder = &Decoder{}

	base64Encoder = &base64.Encoding{}

}

//编码方式total_length | [item_length | key_length | key | value_length | value ]...
func TestEncode(t *testing.T) {
	tests := [](map[string]interface{}){
		map[string]interface{}{
			"a": 1,
			"b": "2",
		},
		map[string]interface{}{
			"abc":  1,
			"ehjd": "2",
			"rrj":  1.9,
		},
	}
	results := []string{"AAAAJAAAABwAAAAKAAAAAWEAAAABMQAAAAoAAAABYgAAAAEy",
		"AAAAOwAAADMAAAAMAAAAA2FiYwAAAAExAAAADQAAAARlaGpkAAAAATIAAAAOAAAAA3JyagAAAAMxLjk="}
	for idx := range tests {
		if str, err := encoder.Encode(tests[idx]); err != nil || base64.StdEncoding.EncodeToString(str) != results[idx] {
			fmt.Print(str)
			t.Error(fmt.Sprintf("encode err:%v，%v,%v", err, results[idx], base64.StdEncoding.EncodeToString(str)))
		}
	}
}
func TestDecode(t *testing.T) {
	tests := []string{"AAAAJAAAABwAAAAKAAAAAWEAAAABMQAAAAoAAAABYgAAAAEy",
		"AAAAOwAAADMAAAAMAAAAA2FiYwAAAAExAAAADQAAAARlaGpkAAAAATIAAAAOAAAAA3JyagAAAAMxLjk="}
	results := [](map[string]string){
		map[string]string{
			"a": "1",
			"b": "2",
		},
		map[string]string{
			"abc":  "1",
			"ehjd": "2",
			"rrj":  "1.9",
		},
	}
	for idx := range tests {
		str, err := base64.StdEncoding.DecodeString(tests[idx])
		if err != nil {
			t.Error(fmt.Sprintf("base64 decode err:%v", err))
		}
		if m, err := decoder.Decode(bytes.NewReader(str)); err != nil || !reflect.DeepEqual(m, results[idx]) {
			t.Error(fmt.Sprintf("decode err:%v %v %v", err, m, results[idx]))
		}
	}
}
