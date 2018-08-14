package gateway

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"time"

	encoding "gitlab.xunlei.cn/xllive/common/library/gateway/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

func CbinRequest(host string, port int, data map[string]interface{}, timeout time.Duration) (map[string]string, error) {
	encoder := &encoding.Encoder{}
	var str []byte
	str, err := encoder.Encode(data)
	if err != nil {
		return nil, err
	}
	addr := fmt.Sprintf("%s:%d", host, port)
	netWithTimeout := net.Dialer{Timeout: timeout}
	conn, err := netWithTimeout.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	_, err = conn.Write(str)
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(conn)
	decoder := &encoding.Decoder{}
	var result map[string]string
	result, err = decoder.Decode(reader)
	if err != nil {
		return nil, err
	}
	for k := range result {
		if k != "unickname" {
			s, err := ioutil.ReadAll(transform.NewReader(bytes.NewReader([]byte(result[k])), simplifiedchinese.GBK.NewDecoder()))
			if err == nil {
				result[k] = string(s)
			}
		}
	}
	return result, nil
}
