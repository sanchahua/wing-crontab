package config

import (
	"gitlab.xunlei.cn/xlsoa/common/utility"
)

type pathTraverser struct {
	m map[string]interface{}
}

func newPathTraverser() *pathTraverser {
	return &pathTraverser{
		m: make(map[string]interface{}),
	}
}

func (t *pathTraverser) digest(path string, value interface{}) {

	sp := utility.NewPathPattern(path)

	cur := t.m
	for i := 0; i < sp.Depth(); i++ {

		s := sp.LevelName(i)

		_, ok := cur[s]
		if !ok {

			if i == sp.Depth()-1 {
				// Assign value at leaf
				cur[s] = value
			} else {
				// Create child
				cur[s] = make(map[string]interface{})
				cur = cur[s].(map[string]interface{})
			}

		} else {

			// Leaf
			if i == sp.Depth()-1 {

				// Already a path here, don't touch it
				_, ok := cur[s].(map[string]interface{})
				if ok {
					continue
				}

				cur[s] = value

			} else {

				// If not a path
				// Remove previous value, make a path instead
				_, ok := cur[s].(map[string]interface{})
				if !ok {
					cur[s] = make(map[string]interface{})
				}

				// Next
				cur = cur[s].(map[string]interface{})

			}

		} //endof if ok

	} // endof for
}

func (t *pathTraverser) get() map[string]interface{} {
	return t.m
}
