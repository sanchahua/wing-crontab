package xllive

import (
	"testing"
	"encoding/json"
	"fmt"
	"encoding/base64"
	"os"
)

type St struct{
	A string `json:"a"`
	B int `json:"b"`
}
// go test -v -test.run TestParseBody
func TestParseBody(t *testing.T) {
	p := &St{
		A: "hello",
		B: 123,
	}
	d, err := json.Marshal(p)
	if err != nil {
		t.Error(err)
		return
	}

	tt := base64.StdEncoding.EncodeToString(d)

	var out = &St{A:"rtgergwerg", B:978}
	err = ParseBody(tt, &out)
	if err != nil {
		t.Error(err)
		return
	}
	if out.A != "hello" || out.B != 123 {
		t.Error("ParseBody fail")
		return
	}
	fmt.Println("json test: ", out)

	var out2 = &St{A:"werfwer", B:6768}
	rawStr := base64.StdEncoding.EncodeToString([]byte("a=word&b=456"))
	err = ParseBody(rawStr, &out2)
	if err != nil {
		t.Error(err)
		return
	}
	if out2.A != "word" || out2.B != 456 {
		t.Error("ParseBody fail")
		return
	}
	fmt.Println("http query test: ", out2)

	rawStr = base64.StdEncoding.EncodeToString([]byte("--------------------------7799a081e2bea540\r\n" +
	"Content-Disposition: form-data; name=\"a\"\r\n"+
	"\r\n"+
	"tst设置\r\n"+
	"--------------------------7799a081e2bea540\r\n"+
	"Content-Disposition: form-data; name=\"b\"\r\n"+
	"\r\n"+
	"10\r\n"+
	"--------------------------7799a081e2bea540\r\n"+
	"Content-Disposition: form-data; name=\"c\"\r\n"+
	"\r\n"+
	"2\r\n"+
	"--------------------------7799a081e2bea540\r\n"))
	var out3 = &St{A:"werfwer", B:6768}
	err = ParseBody(rawStr, &out3)
	if err != nil {
		t.Error(err)
		return
	}
	fmt.Println(out3)
	if out3.A != "tst设置" || out3.B != 10 {
		t.Error("ParseBody fail")
		return
	}

}
type UpdateReq struct{
	Nickname string `json:"nickname"`
	Sign string `json:"sign"`
	Sex int32 `json:"sex"`
	IsAcceptPush int32 `json:"is_accept_push"`
	Body string `json:"body"`
}

// go test -v -test.run TestParseBody2
func TestParseBody2(t *testing.T) {
	raw := `LS1XdE1xVC1Jdk9fMGRHUzNDb3JhX3hMUU9uT24wSVENCkNvbnRlbnQtRGlzcG9zaXRpb246IGZvcm0tZGF0YTsgbmFtZT0ic2V4Ig0KDQoxDQotLVd0TXFULUl2T18wZEdTM0NvcmFfeExRT25PbjBJUQ0KQ29udGVudC1EaXNwb3NpdGlvbjogZm9ybS1kYXRhOyBuYW1lPSJuaWNrbmFtZSINCg0K5Y+l5ouS57udbW9tDQotLVd0TXFULUl2T18wZEdTM0NvcmFfeExRT25PbjBJUQ0KQ29udGVudC1EaXNwb3NpdGlvbjogZm9ybS1kYXRhOyBuYW1lPSJzaWduIg0KDQpKeGp4angNCi0tV3RNcVQtSXZPXzBkR1MzQ29yYV94TFFPbk9uMElRDQo=`
	var out UpdateReq
	err := ParseBody(raw, &out)
	fmt.Println(err)
	fmt.Fprintf(os.Stderr, "%+v", out)
}

type CreateRoomReq struct {
	PingAvg int32 `protobuf:"varint,1,opt,name=ping_avg" json:"ping_avg"`
	From    int32 `protobuf:"varint,2,opt,name=from" json:"from"`
	// string os = 3;// isset($this->client['os']) ? $this->client['os'] : '';
	// string appver = 4;// = isset($this->client['appver']) ? $this->client['appver'] : '';
	// string model = 5;//isset($this->client['model']) ? $this->client['model'] : ''; // 元勋要求新增字段 20170613
	// string osver = 6;//isset($this->client['osver']) ? $this->client['osver'] : ''; //  元勋要求新增字段 20170613
	Isvcaccept int32 `protobuf:"varint,7,opt,name=isvcaccept" json:"isvcaccept"`
	// int64 userid = 8;
	Title string `protobuf:"bytes,9,opt,name=title" json:"title"`
	// int32 create_from = 10 [json_name="create_from"];
	// int32 app_code = 11;
	Body string `protobuf:"bytes,10,opt,name=body" json:"body"`
}
// go test -v -test.run TestParseBody3
func TestParseBody3(t *testing.T) {
	raw := `LS1OQm1TOTFhbERtbV9ZZERSb0FuZm1HUElrS1RUZjcNCkNvbnRlbnQtRGlzcG9zaXRpb246IGZvcm0tZGF0YTsgbmFtZT0idGl0bGUiDQoNCvCfmJPwn5iN8J+Yk/CfmI0NCi0tTkJtUzkxYWxEbW1fWWREUm9BbmZtR1BJa0tUVGY3DQo=`
	var out CreateRoomReq
	err := ParseBody(raw, &out)
	fmt.Println(err)
	fmt.Fprintf(os.Stderr, "%+v", out)
}
