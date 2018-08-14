package config

import (
	"errors"
	"fmt"
	consul_api "github.com/hashicorp/consul/api"
	"log"
	"reflect"
	"testing"
	"time"
)

func testCheckLoaderUpdate(loader Loader, ch chan bool) (*Value, error) {
	select {
	case <-ch:
		return loader.Get(ROOT)
	case <-time.After(2 * time.Second):
		return nil, errors.New("No update event")
	}
}

func testAddConfigEntry(t *testing.T, serviceName string, key string, value string) {
	var err error
	realKey := fmt.Sprintf("%v/%v/%v", testConfigPrefix, serviceName, key)

	kv := testConsulClient.KV()
	pair := &consul_api.KVPair{Key: realKey, Value: []byte(value)}
	if _, err = kv.Put(pair, nil); err != nil {
		t.Fatalf("Failed to add key %v, value %v, error: %v\n", realKey, value, err)
	}
}

func testDelConfigEntry(t *testing.T, serviceName string, key string) {

	realKey := fmt.Sprintf("%v/%v/%v", testConfigPrefix, serviceName, key)
	kv := testConsulClient.KV()
	if _, err := kv.Delete(realKey, nil); err != nil {
		t.Fatalf("Failed to delete key %v, error: %v\n", realKey, err)
	}

}

//delete all keys under the given key prefix
//note the given key will be prefixed with the configured test prefix
func testDeletePrefix(t *testing.T, serviceName string) {
	prefix := fmt.Sprintf("%v/%v/", testConfigPrefix, serviceName)

	kv := testConsulClient.KV()
	if _, err := kv.DeleteTree(prefix, nil); err != nil {
		log.Fatalf("Failed to delete prefix %v, error: %v\n", prefix, err)
	}
}

func TestConfigCenterLoaderLoad(t *testing.T) {
	var err error
	serviceName := fmt.Sprintf("test-service-%v", testRandString())

	// Fake data
	pairs := []struct {
		key   string
		value string
	}{
		{"id", "1234"},
		{"name", "supergui"},
		{"mysql/port", "3306"},
		{"mysql/host", "localhost"},
		{"mysql/user", "root"},
		{"mysql/password", "sd-9898w"},
		{"log/file", "log/server.log"},
		{"log/level", "all"},
		{"token/create/ttl", "3600"},
		{"token/create/num", "2"},
	}

	for _, pair := range pairs {
		testAddConfigEntry(t, serviceName, pair.key, pair.value)
	}

	defer func() {
		testDeletePrefix(t, serviceName)
	}()

	// Test
	prefix := fmt.Sprintf("%v/%v", testConfigPrefix, serviceName)
	loader := NewConfigCenterLoader(testConfigServerAddr, prefix)
	if err = loader.Init(); err != nil {
		t.Fatalf("failed to init config, error: %v\n", err)
	}

	type config struct {
		Name  string `yaml:"name"`
		Id    string `yaml:"id"`
		Mysql struct {
			Host     string `yaml:"host"`
			Port     int    `yaml:"port"`
			User     string `yaml:"user"`
			Password string `yaml:"password"`
		} `yaml:"mysql"`
		Log struct {
			File  string `yaml:"file"`
			Level string `yaml:"level"`
		} `yaml:"log"`
		Token struct {
			Create struct {
				Ttl int32 `yaml:"ttl"`
				Num int32 `yaml:"num"`
			} `yaml:"create"`
		} `yaml:"token"`
	}

	var cfg = &config{}
	var v *Value
	if v, err = loader.Get(ROOT); err != nil {
		t.Fatalf("failed to get ROOT, error: %v\n", err)
	}
	if err = v.Populate(cfg); err != nil {
		t.Fatalf("failed to populate, error: %v\n", err)
	}

	expectedCfg := &config{
		Name: "supergui",
		Id:   "1234",
		Mysql: struct {
			Host     string `yaml:"host"`
			Port     int    `yaml:"port"`
			User     string `yaml:"user"`
			Password string `yaml:"password"`
		}{
			Host:     "localhost",
			Port:     3306,
			User:     "root",
			Password: "sd-9898w",
		},
		Log: struct {
			File  string `yaml:"file"`
			Level string `yaml:"level"`
		}{
			File:  "log/server.log",
			Level: "all",
		},
		Token: struct {
			Create struct {
				Ttl int32 `yaml:"ttl"`
				Num int32 `yaml:"num"`
			} `yaml:"create"`
		}{
			Create: struct {
				Ttl int32 `yaml:"ttl"`
				Num int32 `yaml:"num"`
			}{
				Ttl: 3600,
				Num: 2,
			},
		},
	}

	if !reflect.DeepEqual(cfg, expectedCfg) {
		t.Errorf("loaded config data not correct, loaded: %v, expected: %v\n", cfg, expectedCfg)
	}
}

func TestConfigCenterLoaderWatch(t *testing.T) {
	var err error
	serviceName := fmt.Sprintf("test-service-%v", testRandString())

	pairs := []struct {
		key   string
		value string
	}{
		{"name", "supergui"},
		{"mysql/host", "localhost"},
		{"mysql/port", "3306"},
	}

	for _, pair := range pairs {
		testAddConfigEntry(t, serviceName, pair.key, pair.value)
	}

	defer func() {
		testDeletePrefix(t, serviceName)
	}()

	prefix := fmt.Sprintf("%v/%v", testConfigPrefix, serviceName)
	loader := NewConfigCenterLoader(testConfigServerAddr, prefix)
	if err = loader.Init(); err != nil {
		t.Fatalf("failed to init config, error: %v\n", err)
	}

	type config struct {
		Name  string `yaml:"name"`
		Mysql struct {
			Host string `yaml:"host"`
			Port int    `yaml:"port"`
		} `yaml:"mysql"`
	}

	cfg := &config{}
	expectedCfg := &config{
		Name: "supergui",
		Mysql: struct {
			Host string `yaml:"host"`
			Port int    `yaml:"port"`
		}{
			Host: "localhost",
			Port: 3306,
		},
	}

	var v *Value
	if v, err = loader.Get(ROOT); err != nil {
		t.Fatalf("failed to get ROOT, error: %v\n", err)
	}
	if err = v.Populate(cfg); err != nil {
		t.Fatalf("failed to populate, error: %v\n", err)
	}

	if !reflect.DeepEqual(cfg, expectedCfg) {
		t.Errorf("loaded config data not correct, loaded: %v, expected: %v\n", cfg, expectedCfg)
	}

	var ch chan bool
	ch, err = loader.Watch(ROOT)
	if err != nil {
		t.Fatal(err)
	}

	testAddConfigEntry(t, serviceName, "name", "supergui-modified")
	expectedCfg.Name = "supergui-modified"

	if v, err = testCheckLoaderUpdate(loader, ch); err != nil || v == nil {
		t.Fatalf("testCheckLoaderUpdate fail: %v", err)
	}
	updatedCfg := &config{}
	if err = v.Populate(updatedCfg); err != nil {
		t.Fatalf("failed to populate to updatedCfg, error: %v\n", err)
	}

	if !reflect.DeepEqual(updatedCfg, expectedCfg) {
		t.Errorf("updated config not correct, updated: %v, expected: %v\n", updatedCfg, expectedCfg)
	}
}

func TestConfigCenterLoaderSingleMatch(t *testing.T) {

	type testConfig struct {
		TestKey string `yaml:"testKey"`
	}

	for _, info := range []struct {
		dcName       string
		nodeName     string
		instanceName string
	}{
		{"", "", ""},
		{"", "", "instance1"},
		{"", "node1", ""},
		{"", "node1", "instance1"},
		{"dc1", "", ""},
		{"dc1", "", "instance1"},
		{"dc1", "node1", ""},
		{"dc1", "node1", "instance1"},
	} {
		var err error
		serviceName := fmt.Sprintf("test-service-%v", testRandString())

		var opts = []configCenterLoaderOptionFunc{}
		if info.dcName != "" {
			opts = append(opts, ConfigCenterLoaderWithProperty("dc", info.dcName))
		}
		if info.nodeName != "" {
			opts = append(opts, ConfigCenterLoaderWithProperty("node", info.nodeName))
		}
		if info.instanceName != "" {
			opts = append(opts, ConfigCenterLoaderWithProperty("instance", info.instanceName))
		}
		prefix := fmt.Sprintf("%v/%v", testConfigPrefix, serviceName)
		loader := NewConfigCenterLoader(testConfigServerAddr, prefix, opts...)
		if err = loader.Init(); err != nil {
			t.Errorf("failed to init config, error: %v\n", err)
			continue
		}
		var watchCh chan bool
		watchCh, err = loader.Watch(ROOT)
		if err != nil {
			t.Error(err)
			continue
		}

		// Delete service config prefix on exit
		defer func() {
			testDeletePrefix(t, serviceName)
		}()

		for i, c := range []struct {
			dcName       string
			nodeName     string
			instanceName string
			key          string
			value        string
			expectMatch  bool
		}{
			{TEST_INVALID_DC, info.nodeName, info.instanceName, "testKey", "testvalue-0", false},
			{TEST_INVALID_DC, info.nodeName, TEST_INVALID_INSTANCE, "testKey", "testvalue-1", false},
			{TEST_INVALID_DC, info.nodeName, "", "testKey", "testvalue-2", false},
			{TEST_INVALID_DC, TEST_INVALID_NODE, info.instanceName, "testKey", "testvalue-3", false},
			{TEST_INVALID_DC, TEST_INVALID_NODE, TEST_INVALID_INSTANCE, "testKey", "testvalue-4", false},
			{TEST_INVALID_DC, TEST_INVALID_NODE, "", "testKey", "testvalue-5", false},
			{TEST_INVALID_DC, "", info.instanceName, "testKey", "testvalue-6", false},
			{TEST_INVALID_DC, "", TEST_INVALID_INSTANCE, "testKey", "testvalue-7", false},
			{TEST_INVALID_DC, "", "", "testKey", "testvalue-8", false},
			{"", info.nodeName, info.instanceName, "testKey", "testvalue-9", true},
			{"", info.nodeName, TEST_INVALID_INSTANCE, "testKey", "testvalue-10", false},
			{"", info.nodeName, "", "testKey", "testvalue-11", true},
			{"", TEST_INVALID_NODE, info.instanceName, "testKey", "testvalue-12", false},
			{"", TEST_INVALID_NODE, TEST_INVALID_INSTANCE, "testKey", "testvalue-13", false},
			{"", TEST_INVALID_NODE, "", "testKey", "testvalue-14", false},
			{"", "", info.instanceName, "testKey", "testvalue-15", true},
			{"", "", TEST_INVALID_INSTANCE, "testKey", "testvalue-16", false},
			{"", "", "", "testKey", "testvalue-17", true},
			{info.dcName, "", "", "testKey", "testvalue-18", true},
			{info.dcName, info.nodeName, TEST_INVALID_INSTANCE, "testKey", "testvalue-19", false},
			{info.dcName, info.nodeName, "", "testKey", "testvalue-20", true},
			{info.dcName, info.nodeName, info.instanceName, "testKey", "testvalue-21", true},
			{info.dcName, TEST_INVALID_NODE, info.instanceName, "testKey", "testvalue-22", false},
			{info.dcName, TEST_INVALID_NODE, TEST_INVALID_INSTANCE, "testKey", "testvalue-23", false},
			{info.dcName, TEST_INVALID_NODE, "", "testKey", "testvalue-24", false},
			{info.dcName, "", info.instanceName, "testKey", "testvalue-25", true},
			{info.dcName, "", TEST_INVALID_INSTANCE, "testKey", "testvalue-26", false},
		} {

			cfg := &testConfig{"testDefaultValue"}

			var realKey string
			var filters = []string{}
			if c.dcName != "" {
				filters = append(filters, "dc="+c.dcName)
			}
			if c.nodeName != "" {
				filters = append(filters, "node="+c.nodeName)
			}
			if c.instanceName != "" {
				filters = append(filters, "instance="+c.instanceName)
			}
			if len(filters) > 0 {
				realKey = "["
				for ii, filter := range filters {
					realKey += filter
					if ii < len(filters)-1 {
						realKey += ","
					}
				}
				realKey += "]"
			}
			realKey += c.key
			testAddConfigEntry(t, serviceName, realKey, c.value)

			var v *Value
			if v, err = testCheckLoaderUpdate(loader, watchCh); err != nil {
				t.Errorf("[%v case] testCheckLoaderUpdate, error: %v\n", i, err)
			}
			if err = v.Populate(cfg); err != nil {
				t.Errorf("[%v case] failed to populate, error: %v\n", i, err)
			}

			if c.expectMatch == true {
				if cfg.TestKey != c.value {
					t.Errorf("[%v case] [info %v] expectMatch %v, '%v'(got)!='%v'(expect)\n", i, info, c.expectMatch, cfg.TestKey, c.value)
				}
			} else {
				if cfg.TestKey != "testDefaultValue" {
					t.Errorf("[%v case] [info %v] expectMatch %v, '%v'(got)=='%v'(expect)\n", i, info, c.expectMatch, cfg.TestKey, "testDefaultValue")
				}
			}

			testDelConfigEntry(t, serviceName, realKey)
			_, err = testCheckLoaderUpdate(loader, watchCh)
			if err != nil {
				t.Errorf("testCheckLoaderUpdate for entry delete error: %v\n", err)
			}

		} // endof for i, c

	} // endof for _, info
}

func TestConfigCenterLoaderMultipleMatch(t *testing.T) {

	type testConfig struct {
		TestKey string `yaml:"testKey"`
	}

	testKey := "testKey"
	dcName := "dc1"
	nodeName := "node1"
	instanceName := "instance1"

	for i, c := range []struct {
		entries []struct {
			dcName       string
			nodeName     string
			instanceName string
			key          string
			value        string
		}
		expectValue string
	}{
		{
			[]struct {
				dcName       string
				nodeName     string
				instanceName string
				key          string
				value        string
			}{
				{"", "", "", testKey, "testvalue-1"},
				{"", "", instanceName, testKey, "testvalue-2"},
				{"", nodeName, "", testKey, "testvalue-3"},
				{dcName, "", "", testKey, "testvalue-4"},
				{"", nodeName, instanceName, testKey, "testvalue-5"},
				{dcName, nodeName, "", testKey, "testvalue-6"},
				{dcName, "", instanceName, testKey, "testvalue-7"},
				{dcName, nodeName, instanceName, testKey, "testvalue-8"},
			},
			"testvalue-8",
		},
		{
			[]struct {
				dcName       string
				nodeName     string
				instanceName string
				key          string
				value        string
			}{
				{"", "", "", testKey, "testvalue-1"},
				{"", "", instanceName, testKey, "testvalue-2"},
				{"", nodeName, "", testKey, "testvalue-3"},
				{dcName, "", "", testKey, "testvalue-4"},
				{dcName, nodeName, "", testKey, "testvalue-6"},
			},
			"testvalue-6",
		},
		{
			[]struct {
				dcName       string
				nodeName     string
				instanceName string
				key          string
				value        string
			}{
				{"", "", "", testKey, "testvalue-1"},
				{"", "", instanceName, testKey, "testvalue-2"},
				{"", nodeName, "", testKey, "testvalue-3"},
				{dcName, "", "", testKey, "testvalue-4"},
				{"", nodeName, instanceName, testKey, "testvalue-5"},
				{dcName, "", instanceName, testKey, "testvalue-7"},
			},
			"testvalue-7",
		},
		{
			[]struct {
				dcName       string
				nodeName     string
				instanceName string
				key          string
				value        string
			}{
				{"", "", "", testKey, "testvalue-1"},
				{"", "", instanceName, testKey, "testvalue-2"},
				{"", nodeName, "", testKey, "testvalue-3"},
				{dcName, "", "", testKey, "testvalue-4"},
				{"", nodeName, instanceName, testKey, "testvalue-5"},
			},
			"testvalue-5",
		},
		{
			[]struct {
				dcName       string
				nodeName     string
				instanceName string
				key          string
				value        string
			}{
				{"", "", "", testKey, "testvalue-1"},
				{"", "", instanceName, testKey, "testvalue-2"},
			},
			"testvalue-2",
		},
		{
			[]struct {
				dcName       string
				nodeName     string
				instanceName string
				key          string
				value        string
			}{
				{"", "", "", testKey, "testvalue-1"},
				{"", nodeName, "", testKey, "testvalue-3"},
			},
			"testvalue-3",
		},
		{
			[]struct {
				dcName       string
				nodeName     string
				instanceName string
				key          string
				value        string
			}{
				{"", "", "", testKey, "testvalue-1"},
				{dcName, "", "", testKey, "testvalue-4"},
			},
			"testvalue-4",
		},
		{
			[]struct {
				dcName       string
				nodeName     string
				instanceName string
				key          string
				value        string
			}{
				{"", "", "", testKey, "testvalue-1"},
			},
			"testvalue-1",
		},
	} {
		var err error
		serviceName := fmt.Sprintf("test-service-%v", testRandString())

		defer func() {
			testDeletePrefix(t, serviceName)
		}()

		var opts = []configCenterLoaderOptionFunc{}
		opts = append(opts, ConfigCenterLoaderWithProperty("dc", dcName))
		opts = append(opts, ConfigCenterLoaderWithProperty("node", nodeName))
		opts = append(opts, ConfigCenterLoaderWithProperty("instance", instanceName))
		prefix := fmt.Sprintf("%v/%v", testConfigPrefix, serviceName)
		loader := NewConfigCenterLoader(testConfigServerAddr, prefix, opts...)
		if err = loader.Init(); err != nil {
			t.Errorf("failed to init config, error: %v\n", err)
			continue
		}
		var watchCh chan bool
		watchCh, err = loader.Watch(ROOT)
		if err != nil {
			t.Error(err)
			continue
		}

		// Add multiple entries
		for _, entry := range c.entries {
			var realKey string
			var filters = []string{}
			if entry.dcName != "" {
				filters = append(filters, "dc="+entry.dcName)
			}
			if entry.nodeName != "" {
				filters = append(filters, "node="+entry.nodeName)
			}
			if entry.instanceName != "" {
				filters = append(filters, "instance="+entry.instanceName)
			}
			if len(filters) > 0 {
				realKey = "["
				for ii, filter := range filters {
					realKey += filter
					if ii < len(filters)-1 {
						realKey += ","
					}
				}
				realKey += "]"
			}
			realKey += entry.key
			testAddConfigEntry(t, serviceName, realKey, entry.value)
		}

		// Wait time for notification
		time.Sleep(100 * time.Millisecond)

		// Check value
		cfg := &testConfig{}
		var v *Value
		if v, err = testCheckLoaderUpdate(loader, watchCh); err != nil {
			t.Errorf("testCheckLoaderUpdate, error: %v\n", err)
		}
		if err = v.Populate(cfg); err != nil {
			t.Errorf("failed to populate, error: %v\n", err)
		}

		if cfg.TestKey != c.expectValue {
			t.Errorf("[%v case] '%v'(got)!='%v'(expect)\n", i, cfg.TestKey, c.expectValue)
		}

	}
}
