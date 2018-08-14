package config

type filter struct {
	key   string
	value string
}

func newFilter(key string, value string) *filter {
	return &filter{
		key:   key,
		value: value,
	}
}

func (f *filter) match(properties map[string]string) bool {

	if v, ok := properties[f.key]; ok && v == f.value {
		return true
	}
	return false

}
