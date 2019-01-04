package php

import (
	"testing"
	"fmt"
)

func TestArrayDiff(t *testing.T) {
	var arr1 = []int64{1,2,3}
	var arr2 = []int64{2,5,6,7,8}

	diff := ArrayDiff(arr1, arr2)
	fmt.Println(diff)
	if len(diff) != 2 {
		t.Errorf("ArrayDiff error")
	}
	if diff[0] != 1 || diff[1] != 3 {
		t.Errorf("ArrayDiff error")
	}

	diff = ArrayDiff(arr1, nil)
	fmt.Println(diff)
	if len(diff) != 3 {
		t.Errorf("ArrayDiff error")
	}
	if diff[0] != 1 || diff[1] != 2 {
		t.Errorf("ArrayDiff error")
	}

	diff = ArrayDiff(nil, arr2)
	fmt.Println(diff)
}

func TestInArray(t *testing.T) {
	var arr = []int64{1,2,3}
	if !InArray(1, arr) {
		t.Errorf("InArray error")
	}
	if InArray(0, arr) {
		t.Errorf("InArray error")
	}
	if InArray(-1, arr) {
		t.Errorf("InArray error")
	}
}
