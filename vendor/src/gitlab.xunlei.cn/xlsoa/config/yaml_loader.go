package config

import (
	"github.com/pkg/errors"
)

type yamlLoader struct {
	tree *yamlTree
}

func NewYamlLoader(data []byte) Loader {

	return &yamlLoader{
		tree: newYamlTreeFromByte(data),
	}

}

func (l *yamlLoader) Name() string {
	return "YamlLoader"
}

func (l *yamlLoader) Init() error {
	var err error

	err = l.tree.init()
	if err != nil {
		return errors.Wrap(err, "yamlLoader.Init(): Init tree fail")
	}

	return nil
}

func (l *yamlLoader) Get(key string) (*Value, error) {
	return l.tree.get(key)
}

// Yaml loader unsupport Watch
func (l *yamlLoader) Watch(key string) (chan bool, error) {
	return nil, errors.New("Watch not supported")
}

func (l *yamlLoader) Close() {
	// TODO: handle close
}
