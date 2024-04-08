package main

import (
	"testing"
)

// 创建切片的3种方式
func TestNewSlice(t *testing.T) {
	_ = []int{1, 2, 3} // 字面量
	_ = make([]int, 3) // make

	// 数组 或 切片范围指定
	var _arr = [3]int{1, 2, 3}
	_ = _arr[:]

	// 指定索引和元素
	var s = []int{0: 1, 2: 3}
	println(cap(s) == 3)
}

// 多个切片共享一个数组（扩容时不再共享）
func TestSliceShareArr(t *testing.T) {
	var s1 = []int{1, 2, 3}
	var s2 = s1[1:] // 2,3
	s1[1] = 22
	println(s2[0] == 22) // true
}

// 切片扩容会创建新的数组
func TestSliceGrowing(t *testing.T) {
	var s1 = []int{1, 2, 3}
	println(cap(s1) == 3)

	s2 := s1[1:]
	s1 = append(s1, 4)
	s2[0] = 22
	println(s1[1] == 22) // false
}

func TestIterateSlice(t *testing.T) {
	type User struct {
		Name string
	}

	var ss = []User{
		{"a"},
		{"b"},
		{"c"},
	}

	for _, v := range ss {
		v.Name = "new"
	}

	println(ss[0].Name == "new") // false
}

// 切片可以通过下标修改非指针结构体元素的成员字段
func TestUpdateElemByIndex(t *testing.T) {
	type User struct {
		Name string
	}

	var ss = []User{{"a"}}
	ss[0].Name = "new"

	println(ss[0].Name == "new") // true
}
