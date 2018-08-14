package config

type Loader interface {
	Name() string

	Init() error

	// Get 'ROOT' for global value.
	// Or hierarchicall key is supported, example, 'mysql.user'.
	Get(key string) (*Value, error)

	// 1. Currently only 'ROOT' key is supported, panic if not.
	//    Hierarchical key support (exmpale 'mysql.user'), is a future issue.
	//
	// 2. For implementation of a concrete type XXXLoader, mutiple watchers MUST be supported.
	//    Which means you MUST create a new channel, for everytime 'Watch' is called.
	//    When change happens, event MUST be broadcasted to all channels.
	Watch(key string) (chan bool, error)

	Close()
}
