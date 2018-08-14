package config

import (
	"crypto/md5"
	"crypto/rand"
	"fmt"
	consul_api "github.com/hashicorp/consul/api"
	"log"
	"os"
	"testing"
)

const (
	testConfigPrefix = CONFIG_KEY_PREFIX

	TEST_INVALID_DC       string = "iNValiD_Dc"
	TEST_INVALID_NODE     string = "iNValiD_NodE"
	TEST_INVALID_INSTANCE string = "iNValiD_INstANCe"

	TEST_CACHE_LOADER_DIR = "./test_cache/"
)

var (
	testConfigServerAddr = ""
	testConsulAddr       = ""

	testConsulClient *consul_api.Client = nil
)

func testRandString() string {

	buf := make([]byte, 10)
	rand.Read(buf)
	return fmt.Sprintf("%x", md5.Sum(buf))
}

func TestMain(m *testing.M) {

	testConfigServerAddr = os.Getenv("TEST_XLSOA_CONFIG_SERVER_ADDR")
	if testConfigServerAddr == "" {
		log.Fatal("testConfigServerAddr is required. Set with environment 'TEST_XLSOA_CONFIG_SERVER_ADDR'.\n")
	}
	testConsulAddr = os.Getenv("TEST_XLSOA_CONFIG_SERVER_CONSUL_ADDR")
	if testConsulAddr == "" {
		log.Fatal("testConsulAddr is required. Set with environment 'TEST_XLSOA_CONFIG_SERVER_CONSUL_ADDR'.\n")
	}

	var err error
	testConsulClient, err = consul_api.NewClient(
		&consul_api.Config{
			Address: testConsulAddr,
		},
	)
	if err != nil {
		log.Fatalf("New consul client fail: %v\n", err)
	}

	ret := m.Run()

	// Clean cache_loader_test data
	os.RemoveAll(TEST_CACHE_LOADER_DIR)

	os.Exit(ret)
}
