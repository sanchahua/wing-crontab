package jwt

import "testing"

func TestConfig(t *testing.T) {

	conf := NewJwtConfig()
	err := conf.LoadFromFile("./config.json")

	if err == nil {
		t.Log("sign:", conf.Sign([]string{"target"}))
		if conf.ServiceId != "service_id" {
			t.Fatal("ServiceId")
		}
		if conf.ServiceName != "service_name" {
			t.Fatal("ServiceName")
		}
		if conf.ClientId != "client_id" {
			t.Fatal("ClientId")
		}
		if conf.ClientSecret != "client_secret" {
			t.Fatal("ClientSecret")
		}
		if conf.AuthServiceName != "xlsoa.core.auth.Auth" {
			t.Fatal("AuthServiceName")
		}
		if conf.KeySyncServiceName != "xlsoa.core.key_sync.KeySync" {
			t.Fatal("ClientSecret")
		}
	}else {
		t.Fatal("can't read ./jwt.json")
	}
}
