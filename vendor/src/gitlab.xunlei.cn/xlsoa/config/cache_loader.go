package config

import (
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"log"
	"os"
)

type cacheLoader struct {
	Loader

	// We create a cacher from local file, which concrete is YamlLoader, when the upper Loader is abnormal.
	// Once the upper Loader become normal, we will remove the cacher.
	cacher Loader

	dir        string
	file       string
	watcherChs []chan bool
}

func NewCacheLoader(dir string, file string, loader Loader) Loader {
	return &cacheLoader{
		Loader:     loader,
		dir:        dir,
		file:       file,
		watcherChs: make([]chan bool, 0),
	}
}

func (l *cacheLoader) String() string {
	return "CacheLoader"
}

func (l *cacheLoader) Init() error {

	var err error
	var v *Value

	// Make data dir
	err = l.mkdir()
	if err != nil {
		return err
	}

	// We MUST have the ch before Get, avoid missing update events between Get and Watch.
	ch, _ := l.Loader.Watch(ROOT)

	if v, err = l.Loader.Get(ROOT); err != nil || v == nil {
		//log.Println("Get from Loader fail, Load from local file now")
		err = l.initCacherFromFile()
		if err != nil {
			return errors.Wrap(err, "However fail with local file")
		}
	} else {
		l.refreshFile(v)
	}

	if ch != nil {
		go l.watchCh(ch)
	}

	return nil
}

func (l *cacheLoader) mkdir() error {
	var err error
	err = os.MkdirAll(l.dir, 0755)
	if err != nil {
		return errors.Wrap(err, "Make temperary dir fail")
	}
	return nil
}

func (l *cacheLoader) Get(key string) (*Value, error) {

	if l.cacher != nil {
		return l.cacher.Get(key)
	}

	return l.Loader.Get(key)
}

func (l *cacheLoader) Watch(key string) (chan bool, error) {

	if key != ROOT {
		panic("Currently only 'ROOT' key support watch.")
	}

	var ch = make(chan bool, 1)
	l.watcherChs = append(l.watcherChs, ch)

	return ch, nil
}

func (l *cacheLoader) refreshFile(v *Value) error {

	var err error
	var data []byte
	data, err = v.ToYamlByte()
	if err != nil {
		return errors.Wrap(err, "Value to yaml byte fail")
	}

	var tmpFile *os.File
	var tmpName string

	tmpFile, err = ioutil.TempFile(l.dir, "")
	if err != nil {
		log.Println(err)
		return err
	}
	tmpName = tmpFile.Name()

	_, err = tmpFile.Write([]byte(data))
	if err != nil {
		log.Println(err)
		return err
	}
	tmpFile.Close()

	targetName := fmt.Sprintf("%v/%v", l.dir, l.file)
	err = os.Rename(tmpName, targetName)
	if err != nil {
		log.Println(err)
		os.Remove(tmpName)
		return err
	}

	//log.Printf("Refresh cache file success: %v\n", targetName)
	return nil

}

func (l *cacheLoader) initCacherFromFile() error {

	path := fmt.Sprintf("%v/%v", l.dir, l.file)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("ReadFile %v error", l.file))
	}

	loader := NewYamlLoader(data)
	err = loader.Init()
	if err != nil {
		return errors.Wrap(err, "YamlLoader.Init from file data fail")
	}

	l.cacher = loader

	//log.Printf("init cache loader from file success %v/%v\n", l.dir, l.file)
	return nil
}

func (l *cacheLoader) watchCh(ch chan bool) error {

	for {
		select {
		case <-ch:
			v, err := l.Loader.Get(ROOT)
			if err != nil || v == nil {
				continue
			}

			// Remove cacher here.
			// CacheLoader don't want to keep a copy data, if the upper Loader is in work.
			l.cacher = nil

			l.broadcastUpdate()
			l.refreshFile(v)
		}
	}

	return nil
}

func (l *cacheLoader) broadcastUpdate() {

	for _, ch := range l.watcherChs {

		select {
		case ch <- true:
			break
		default:
			break
		}

	}
}
