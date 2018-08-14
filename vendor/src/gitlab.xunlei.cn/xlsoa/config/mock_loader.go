package config

// MockLoaders are used for testing.

import (
	"github.com/pkg/errors"
)

type mockSetter interface {
	Set(key string, value string) error
}

type mockRecover interface {
	Recover()
}

// MockLoader
type mockLoader struct {
	value      *Value
	watcherChs []chan bool
}

func newMockLoader() Loader {
	return &mockLoader{
		watcherChs: make([]chan bool, 0),
	}
}
func (l *mockLoader) Name() string {
	return "MockLoader"
}

func (l *mockLoader) Init() error {

	var err error

	if err = l.Set(ROOT, "supergui"); err != nil {
		return nil
	}

	return nil
}

func (l *mockLoader) Get(key string) (*Value, error) {
	return l.value, nil
}

func (l *mockLoader) Watch(key string) (chan bool, error) {
	var ch = make(chan bool, 1)
	l.watcherChs = append(l.watcherChs, ch)

	return ch, nil
}

func (l *mockLoader) Close() {
}

func (l *mockLoader) Set(key string, value string) error {

	var err error

	tree := newYamlTreeFromByte([]byte(value))
	err = tree.init()
	if err != nil {
		return err
	}

	l.value = newValue(tree)

	// Broadcast
	for _, ch := range l.watcherChs {

		select {
		case ch <- true:
			break
		default:
			break
		}

	}

	return nil
}

// MockLoaderGetFail
type mockLoaderGetFail struct {
	Loader
}

func newMockLoaderGetFail() Loader {
	return &mockLoaderGetFail{
		Loader: newMockLoader(),
	}
}

func (l *mockLoaderGetFail) Get(key string) (*Value, error) {
	return nil, errors.New("Fake fail")
}

// MockLoaderGetFailAndRecover
type mockLoaderGetFailAndRecover struct {
	Loader
	normal bool
}

func newMockLoaderGetFailAndRecover() Loader {
	return &mockLoaderGetFailAndRecover{
		Loader: newMockLoader(),
		normal: false,
	}
}

func (l *mockLoaderGetFailAndRecover) Get(key string) (*Value, error) {
	if !l.normal {
		return nil, errors.New("Fake fail")
	}

	return l.Loader.Get(key)
}

func (l *mockLoaderGetFailAndRecover) Set(key string, value string) error {
	if setter, ok := l.Loader.(mockSetter); ok {
		return setter.Set(key, value)
	}

	return errors.New("Not a setter ")
}
func (l *mockLoaderGetFailAndRecover) Recover() {
	l.normal = true
}
