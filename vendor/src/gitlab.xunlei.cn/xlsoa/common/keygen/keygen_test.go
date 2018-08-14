package keygen

import (
	"fmt"
	"testing"
)

func TestAppId(t *testing.T) {
	g := New()

	id, err := g.AppId()

	if err != nil {
		t.Fatalf("AppId() fail")
	}
	if len(id) == 0 {
		t.Fatalf("length of ID is 0")
	}

	fmt.Println("AppId ", id)
}

func TestClientId(t *testing.T) {

	g := New()

	id, secret, err := g.ClientId()
	if err != nil {
		t.Fatalf("ClientId() fail")
	}

	if len(id) == 0 {
		t.Fatalf("Length of ID is 0")
	}
	if len(secret) == 0 {
		t.Fatalf("Length of secret is 0")
	}
	fmt.Println("Id ", id, ", Secret", secret)
}
