package config

import (
	consul_api "github.com/hashicorp/consul/api"
	"github.com/pkg/errors"
	"gitlab.xunlei.cn/xlsoa/common/utility"
	"gopkg.in/yaml.v2"
	"log"
	"time"
)

type configCenterLoaderOptionFunc func(p *configCenterLoader)

type configCenterLoader struct {
	Loader

	addr         string
	prefix       string
	consulClient *consul_api.Client
	lastIndex    uint64
	properties   map[string]string
	watcherChs   []chan bool
	watcherStart int32
}

func NewConfigCenterLoader(addr string, prefix string, opts ...configCenterLoaderOptionFunc) Loader {

	p := &configCenterLoader{
		//Loader:       NewYamlLoader(),
		addr:         addr,
		prefix:       prefix,
		properties:   make(map[string]string),
		watcherChs:   make([]chan bool, 0),
		watcherStart: 0,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

func (l *configCenterLoader) Name() string {
	return "ConfigCenterLoader"
}

func (l *configCenterLoader) Init() error {

	var err error

	// New consul client
	l.consulClient, err = consul_api.NewClient(
		&consul_api.Config{
			Address: l.addr,
		},
	)
	if err != nil {
		return errors.Wrap(err, "configCenterLoader.Init(): New consul client fail")
	}

	// Load on start
	err = l.loadAndUpdate()
	if err != nil {
		log.Printf("ConfigCenterLoader:Init loadAndUpdate fail, error : %v\n", err)
	}

	go l.watch()

	return nil
}

func (l *configCenterLoader) Get(key string) (*Value, error) {

	if l.Loader == nil {
		return nil, errors.New("Loader not ready")
	}
	return l.Loader.Get(key)
}

func (l *configCenterLoader) Watch(key string) (chan bool, error) {

	if key != ROOT {
		panic("Currently only 'ROOT' key support watch.")
	}

	var ch = make(chan bool, 1)
	l.watcherChs = append(l.watcherChs, ch)

	return ch, nil
}

// Load entries from consul.
//
// 1. List prefix.
// 2. Filter with filters.
func (l *configCenterLoader) load() (map[string]*entry, error) {

	var err error
	var kvs consul_api.KVPairs
	var meta *consul_api.QueryMeta

	// Load or watch
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		kv := l.consulClient.KV()
		kvs, meta, err = kv.List(
			l.prefix,
			&consul_api.QueryOptions{
				WaitIndex: l.lastIndex,
			},
		)
		if err != nil {
			log.Println(err)
			time.Sleep(1 * time.Second)
			continue
		}
		break
	}
	if err != nil {
		return nil, errors.Wrap(err, "load from consul fail")
	}

	// Not updated
	if l.lastIndex == meta.LastIndex {
		log.Println("Not updated")
		return nil, nil
	}
	l.lastIndex = meta.LastIndex

	// Serialize consul kv to entry
	prefixDepth := utility.NewPathPattern(l.prefix).Depth()
	entries := make(map[string]*entry)
	for _, kv := range kvs {
		//log.Printf("Raw kv %v, value %v\n", kv.Key, string(kv.Value))

		// Remove prefix
		pp := utility.NewPathPattern(kv.Key)
		if pp.Depth() <= prefixDepth {
			log.Printf("Invalid key %v\n", kv.Key)
			continue
		}
		kv.Key = pp.Format(prefixDepth) // Remove prefix with prefix.Depth()

		// Convert to entry
		// Filter entry
		e := newEntry()
		e.decode(kv)

		if !e.match(l.properties) {
			//log.Printf("filters out kv '%v', filters '%v'\n", kv.Key, l.properties)
			continue
		}

		// Keep longest-conditions-match entry
		old, ok := entries[e.key]
		if ok && old.filterMatchCount >= e.filterMatchCount {
			continue
		}

		entries[e.key] = e
	}

	return entries, nil
}

func (l *configCenterLoader) loadAndUpdate() error {
	// Load entries
	var entries map[string]*entry
	var err error
	entries, err = l.load()
	if err != nil {
		return errors.Wrap(err, "load error")
	}
	if entries == nil {
		return errors.Wrap(err, "load empty entry")
	}

	// Init yaml provider
	err = l.updateYamlLoader(entries)
	if err != nil {
		return errors.Wrap(err, "updateYamlLoader error")
	}
	return nil
}

func (l *configCenterLoader) watch() {

	for {

		err := l.loadAndUpdate()
		if err != nil {
			time.Sleep(3 * time.Second)
			continue
		}

		l.broadcastUpdate()

	}
}

func (l *configCenterLoader) broadcastUpdate() {

	for _, ch := range l.watcherChs {

		select {
		case ch <- true:
			break
		default:
			break
		}

	}
}

func (l *configCenterLoader) updateYamlLoader(entries map[string]*entry) error {

	// Convert entries to yaml text
	traverser := newPathTraverser()
	for _, entry := range entries {
		traverser.digest(entry.key, convertValue(string(entry.value)))
	}

	data, err := yaml.Marshal(traverser.get())
	if err != nil {
		return errors.Wrap(err, "configCenterLoader.updateYamlLoader():  yaml.Marshal(traverser.get()) fail")
	}

	newLoader := NewYamlLoader(data)
	err = newLoader.Init()
	if err != nil {
		return errors.Wrap(err, "configCenterLoader.updateYamlLoader(): yamlLoader.Init fail")
	}

	// TODO: race condition?
	l.Loader = newLoader

	return nil
}

func ConfigCenterLoaderWithProperty(name string, value string) configCenterLoaderOptionFunc {
	return func(l *configCenterLoader) {
		l.properties[name] = value
	}
}
