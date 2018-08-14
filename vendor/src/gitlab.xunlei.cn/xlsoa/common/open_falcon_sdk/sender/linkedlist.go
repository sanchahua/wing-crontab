package open_falcon_sender

import (
	"container/list"
	"sync"
)

type SafeLinkedList struct {
	sync.RWMutex
	L *list.List
	maxLen int
}

func NewSafeLinkedList(maxLen int) *SafeLinkedList {
	return &SafeLinkedList{L: list.New(), maxLen:maxLen}
}

func (this *SafeLinkedList) PopBack(limit int) []*JsonMetaData {
	this.RLock()
	defer this.RUnlock()
	sz := this.L.Len()
	if sz == 0 {
		return []*JsonMetaData{}
	}

	if sz < limit {
		limit = sz
	}

	ret := make([]*JsonMetaData, 0, limit)
	for i := 0; i < limit; i++ {
		e := this.L.Back()
		ret = append(ret, e.Value.(*JsonMetaData))
		this.L.Remove(e)
	}

	return ret
}

func (this *SafeLinkedList) PushFront(v interface{}) *list.Element {
	this.Lock()
	defer this.Unlock()
	if this.L.Len() > this.maxLen {
		return nil
	}
	return this.L.PushFront(v)
}

func (this *SafeLinkedList) Front() *list.Element {
	this.RLock()
	defer this.RUnlock()
	return this.L.Front()
}

func (this *SafeLinkedList) Len() int {
	this.RLock()
	defer this.RUnlock()
	return this.L.Len()
}


