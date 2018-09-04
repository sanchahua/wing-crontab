package php

// return arr1 - arr2
func ArrayDiff(arr1 []int64, arr2 []int64) []int64 {
	var diff = make([]int64, 0)
	for _, v := range arr1 {
		found := false
		for _, v2 := range arr2 {
			if v2 == v {
				found = true
				break
			}
		}
		if !found {
			diff = append(diff, v)
		}
	}
	return diff
}

func InArray(i int64, arr []int64) bool {
	for _, v := range arr {
		if v == i {
			return true
		}
	}
	return false
}

// 移除指定的index
func Unset(arr []int64, index int) {

}

func ArraySearch(value int64, arr []int64) int {
	return 0
}
