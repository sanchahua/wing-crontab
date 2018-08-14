package jws

import (
	"testing"
	"strings"
	"encoding/json"
	"reflect"
	"golang.org/x/oauth2/jws"
)

func Test(t *testing.T) {
	secert := "123456789987654321abcdefghijklmnopkrstuvwxyz"
	errSecert := "asbsfdsdfsdaffd"
	header := &jws.Header{
		KeyID:  "my_client_id",
	}
	claimSet := &jws.ClaimSet{
		Scope:      strings.Join([]string{"s1", "s2", "s3"}, " "),
		Aud:        "http://localhost:8500",
		PrivateClaims:  map[string]interface{}{"a":100, "b" : "assd"},
	}
	token, err := Encode(header, claimSet, secert)
	if err != nil {
		t.Fatal("Encode", err)
	}
	h1, c1, _, claimSetBytes := Decode(token)
	if ( h1 == nil || c1 == nil ){
		t.Fatal("header or claimset decode err", h1, c1)
	}else {
		if h1.KeyID != header.KeyID {
			t.Fatal("decode header err")
		}
		if c1.Aud != claimSet.Aud {
			t.Fatal("decode claimset err")
		}
	}

	t.Log(token)

	var privateClaims = map[string]interface{}{}
	json.Unmarshal(claimSetBytes, &privateClaims)
	value, ok := privateClaims["a"]
	if !ok {
		t.Fatal("private claims err, can't find a")
	}
	realValue, ok := value.(float64)
	t.Log(reflect.TypeOf(value))
	if realValue != float64(100){
		t.Fatal("private claims err, value err", realValue, ok, value)
	}


	if Verify(token, secert) != nil {
		t.Fatal("verify should pass :")
	}

	if Verify("sdfjsdfsaf.safdaer.wesdf", errSecert) == nil {
		t.Fatal("verify should failed :")
	}

	if Verify(token, errSecert) == nil {
		t.Fatal("verify should failed :")
	}

	if Verify("eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCIsImtpZCI6ImNsaWVudGlkIn0.eyJpc3MiOiIiLCJhdWQiOiJodHRwOi8vbG9jYWxob3N0Ojg1MDAiLCJleHAiOjE0ODMwODI5ODQsImlhdCI6MTQ4MzA3OTM4NH0.yi10UNbT9lyLAb2bBF7sLXW1Otb+I1i2y2lKx2BReEU=1", secert) == nil {
		t.Fatal("verify should failed :", )
	}
}
