package net

import (
	"net/http"
	"net"
	"time"
	"fmt"
	"bytes"
	"io/ioutil"
	"errors"
	log "github.com/sirupsen/logrus"
)


type Http struct {
	url string
}

func NewHttp(url string) *Http {
	request := &Http{
		url : url,
	}
	return request
}

var defaultHttpClient = http.Client {
	Transport: &http.Transport {
		MaxIdleConnsPerHost : 8,
		Dial: func(netw, addr string) (net.Conn, error) {
			//deadline := time.Now().Add(HTTP_POST_TIMEOUT * time.Second)
			dial := net.Dialer{
				Timeout:   3 * time.Second,
				KeepAlive: 86400 * time.Second, //一天之内有效
			}
			conn, err := dial.Dial(netw, addr)
			if err != nil {
				return conn, err
			}
			//conn.SetDeadline(deadline)
			return conn, nil
		},
	},
}


func request(method string, url string, data []byte)  ([]byte, error) {
	reader := bytes.NewReader(data)
	req, err := http.NewRequest(method, url, reader)
	if err != nil {
		log.Errorf("syshttp request error-1:%+v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "keep-alive")

	//req.Header.Set("AppKey", APP_KEY);  // 秘钥
	//req.Header.Set("Nonce", nonce);     // 随机数
	//req.Header.Set("CurTime", curTime);
	//req.Header.Set("CheckSum", checkSum);
	//req.Header.Set("Content-Type", "application/x-www-form-urlencoded;charset=utf-8");

	resp, err := defaultHttpClient.Do(req)
	if err != nil {
		log.Errorf("http request error-3:%+v", err)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Errorf("http request error错误的状态码:%+v", resp.StatusCode)
		return nil, errors.New(fmt.Sprintf("错误的状态码：%d", resp.StatusCode))
	}
	res, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("http request error-4:%+v", err)
		return nil, err
	}
	return res, nil
}

func (req *Http) Put(data []byte) ([]byte, error) {
	return request("PUT", req.url, data)
}

func (req *Http) Post(data []byte) ([]byte, error) {
	log.Debugf("post: %s", string(data))
	return request("POST", req.url, data)
}

func (req *Http) Get() ([]byte, error) {
	return request("GET", req.url, nil)
}

func (req *Http) Delete() ([]byte, error) {
	return request("DELETE", req.url, nil)
}
