package config

import (
	"testing"
	"time"
)

func testCacheLoaderAssertStringValue(t *testing.T, loader Loader, value string) {
	var err error
	var v *Value
	v, err = loader.Get(ROOT)
	if err != nil {
		t.Fatalf("Get ROOT error: %v\n", err)
	}
	if v == nil {
		t.Fatalf("Get ROOT not exists\n")
	}

	s, ok := v.TryAsString()
	if !ok {
		t.Fatalf("Value is not string: %v\n", v)
	} else if s != value {
		t.Fatalf("Value '%v'(got)!='%v'(expect)\n", s, value)
	}
}

func testCacheLoaderAssertUpdate(t *testing.T, ch chan bool) {

	select {
	case <-ch:
		return
	case <-time.After(1 * time.Second):
		t.Fatalf("Not update event after updated")
	}
}

func TestCacheLoaderGetAndUpdate(t *testing.T) {

	var err error
	var ch chan bool

	mLoader := newMockLoader()
	err = mLoader.Init()
	if err != nil {
		t.Fatalf("Init MockLoader error: %v\n", err)
	}

	name := testRandString() + ".yaml"
	loader := NewCacheLoader(TEST_CACHE_LOADER_DIR, name, mLoader)
	err = loader.Init()
	if err != nil {
		t.Fatalf("Init CacheLoader error: %v\n", err)
	}
	ch, err = loader.Watch(ROOT)
	if err != nil {
		t.Fatalf("Get watch error: %v\n", err)
	}

	testCacheLoaderAssertStringValue(t, loader, "supergui")

	setter, _ := mLoader.(mockSetter)
	setter.Set(ROOT, "supergui-modified")

	testCacheLoaderAssertUpdate(t, ch)
	testCacheLoaderAssertStringValue(t, loader, "supergui-modified")
}

func TestCacheLoaderGetFailWithoutPreCache(t *testing.T) {
	var err error

	mLoader := newMockLoaderGetFail()
	err = mLoader.Init()
	if err != nil {
		t.Fatalf("Init MockLoader error: %v\n", err)
	}

	name := testRandString() + ".yaml"
	loader := NewCacheLoader(TEST_CACHE_LOADER_DIR, name, mLoader)
	err = loader.Init()
	if err == nil {
		t.Fatalf("Init CacheLoader success, expect error because mock loader Get fail\n")
	}
}

func TestCacheLoaderGetFailAndLoadWithPreCache(t *testing.T) {
	var err error

	name := testRandString() + ".yaml"

	// Pre-cache
	{
		mLoader := newMockLoader()
		err = mLoader.Init()
		if err != nil {
			t.Fatalf("Init MockLoader error: %v\n", err)
		}

		loader := NewCacheLoader(TEST_CACHE_LOADER_DIR, name, mLoader)
		err = loader.Init()
		if err != nil {
			t.Fatalf("Init CacheLoader error: %v\n", err)
		}

		testCacheLoaderAssertStringValue(t, loader, "supergui")
	}

	// Init with GetFail
	// Should Get success
	{
		var ch chan bool

		mLoader := newMockLoaderGetFailAndRecover()
		err = mLoader.Init()
		if err != nil {
			t.Fatalf("Init MockLoader error: %v\n", err)
		}

		loader := NewCacheLoader(TEST_CACHE_LOADER_DIR, name, mLoader)
		err = loader.Init()
		if err != nil {
			t.Fatalf("Init CacheLoader error: %v\n", err)
		}
		ch, err = loader.Watch(ROOT)
		if err != nil {
			t.Fatalf("Get watch error: %v\n", err)
		}

		testCacheLoaderAssertStringValue(t, loader, "supergui")

		// Recover and Set
		recover, _ := mLoader.(mockRecover)
		recover.Recover()
		setter, _ := mLoader.(mockSetter)
		setter.Set(ROOT, "supergui-modified")

		testCacheLoaderAssertUpdate(t, ch)
		testCacheLoaderAssertStringValue(t, loader, "supergui-modified")

	}
}
