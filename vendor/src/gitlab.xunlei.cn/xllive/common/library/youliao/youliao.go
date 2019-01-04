package youliao

import (
	"fmt"
	"crypto/md5"
	"encoding/json"
	"time"
	"net/http"
	"io/ioutil"
	"net/url"
	"strings"
	"sort"
	"encoding/hex"
	"strconv"
	"github.com/parnurzeal/gorequest"
	"github.com/rs/xid"
	"os"
)

var appid = "37"
var appSecret = "06ee163bd31e34fd99d0fe67fbd63a01"
var server = map[string]string{
	"dev":"http://t03-api.xlmc.xunlei.com/api/",
	"test":"http://t03-api.xlmc.xunlei.com/api/",
	"pre":"http://pre.api.tw06.xlmc.sandai.net/api/",
	"online":"http://api.tw06.xlmc.sandai.net/api/",
}

var slPushServer = map[string] string {
	"dev":"http://test.api-shoulei-ssl.xunlei.com/",
	"test":"http://api-shoulei-ssl.xunlei.com/",
	"pre":"http://test.api-shoulei-ssl.xunlei.com/",
	"online":"http://api-shoulei-ssl.xunlei.com/",
}
var ttl = int64(60000000)
var channel = 0
var badge = 0
var deviceId = ""

func buildSign(param map[string]string )(string ){
	str :=""
	names := make([]string, 0, len(param))
	for name,value := range param {
		if len(value) <1{
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names) //sort by key
	for _, name := range names {
		str = fmt.Sprintf("%s%s=%s",str,name,param[name])
	}
	str = str+appSecret

	h := md5.New()
	h.Write([]byte(str))
	x := h.Sum(nil)
	y := hex.EncodeToString(x)
	ret := fmt.Sprintf("%s",y)
	return ret
}
func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func SetTimeout(timeout int64) int64{
	ttl = timeout
	return timeout
}

func SetChannel(chnl int) int{
	channel = chnl
	return channel
}

func SetBadge(bg int) int{
	badge = bg
	return badge
}

func SetDeviceId(id string) string {
	deviceId = id
	return deviceId
}
func queryYouliao(api string,param map[string]string,env string)(string){
	params :=map[string]string{
		"v":"1.0",
		"appId":appid,
		"callId":strconv.FormatInt(makeTimestamp(),10),
	}
	for key,value :=range param{
		if len(value) <1 {
			continue
		}
		params[key]=value
	}
	sign := buildSign(params)
	paramStr := HttpBuildQuery(params)
	requestUrl := server[env]+api+"?"+paramStr+"&sig="+sign
	ret,err :=httpGet(requestUrl)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Fprintf(os.Stderr,"$$$$YouliaoPush%+v\r\n", requestUrl)
	fmt.Fprintf(os.Stderr,"$$$$YouliaoPush%+v\r\n", ret)
	return ret

}
func HttpBuildQuery(params map[string]string) (param_str string) {
	u := url.Values{}
	for k, v := range params {
		u.Set(k,v)
	}
	return u.Encode()
}

func httpGet( url string) (string,error){
	resp, err := http.Get(url)
	if err != nil {
		// handle error
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return string(body),err
	}

	return string(body),nil
}
func SendMiPush(t int,title string,description string,topic string,numId []int64,extra map[string]interface{},env string )(string,error){
	//_,err := SendSdkAndroidPush(t, title, description, topic, numId, extra, env)
	//if err != nil{
	//}
	return SendMiPushV2(t, title, description, topic, numId, extra, env)
}

func SendMiPushV2(t int,title string,description string,topic string,numId []int64,extra map[string]interface{},env string )(string,error){
	api :="admin/mipush/send"
	extraString,err := json.Marshal(extra)
	if err != nil{
		return "",err
	}
	newNumId := []string{}
	for _,id :=range numId {
		newNumId = append(newNumId,strconv.FormatInt(id,10))
	}
	queryParams :=map[string]string{
		"channel":strconv.Itoa(channel),
		"type":strconv.Itoa(t),
		"title":title,
		"topic":topic,
		"userIds":strings.Join(newNumId,","),
		"description":description,
		"extra":string(extraString),
		"ttl":strconv.FormatInt(ttl,10),
		"badge":strconv.Itoa(badge),
		"deviceId":deviceId,
	}
	ret := queryYouliao(api,queryParams,env)
	return ret,nil
}

func SendSdkAndroidPush(t int,title string,description string,topic string,numId []int64,extra map[string]interface{},env string )(string,error) {
	api := "push_services/admin/push/singlecast/"

	if t<146 || t>151 {
		return "", nil
	}

	if len(numId) <=0 {
		return "", nil
	}

	if env != "online" {
		env="test"
	}
	var sdkPush sdkPushStruct

	guid := xid.New()
	androidExtras := sdkPushAndroidExtras{guid.String(), 23, "live_pw", title, description, extra, 1005}

	pushDataMsg := sdkPushDataMsg{title, description, androidExtras}


	sdkPush = sdkPushStruct{"uid", "Android", numId, 0, 15, pushDataMsg, "0", 0}

	var requestUrl = slPushServer[env] + api
	sdkPushBytes, _ := json.Marshal(sdkPush)
	sdkPushStr := string(sdkPushBytes[:])
	request := gorequest.New()
	_, body, errs := request.Post(requestUrl).
		Set("content-type","application/json").
		Send(string(sdkPushStr[:])).
		End()
	fmt.Fprintf(os.Stderr,"$$$$SendSdkAndroidPush%+v\r\n", requestUrl)
	fmt.Fprintf(os.Stderr,"$$$$SendSdkAndroidPush%+v\r\n", sdkPushStr)
	fmt.Fprintf(os.Stderr,"$$$$SendSdkAndroidPush%+v\r\n", body)
	if len(errs)>0 {
		return body,errs[0]
	}	else {
		return body, nil
	}
}


