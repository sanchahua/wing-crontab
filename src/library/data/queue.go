package data

type Queue struct {
	len int64
	cache []interface{}
}

func NewDataQueue(len int64) *Queue {
	c := new(Queue)
	c.len = len
	c.cache = make([]interface{}, len)
	return c
}
