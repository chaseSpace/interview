//go:generate go test  ds_array_test.go ds_array.go -v
package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestArray_Insert(t *testing.T) {
	n := 2
	a := NewArray(n)

	assert.Equal(t, 0, a.Len())
	assert.Equal(t, n, a.Cap())
	assert.Equal(t, a.Data(), []interface{}{})

	a.Insert(1, 0) // [1]
	assert.Equal(t, 1, a.Len())
	assert.Equal(t, a.Data(), []interface{}{1})

	// 从未填空间插入，则前面元素不变
	a.Insert(2, 1) // [1 2]

	assert.Equal(t, a.Data(), []interface{}{1, 2})
	assert.Equal(t, 2, a.Len())

	// 满了，再插入最后一个位置，长度不变，最后一个元素被覆盖
	a.Insert(3, 1)
	assert.Equal(t, a.Data(), []interface{}{1, 3})
	assert.Equal(t, 2, a.Len())

	assert.Panics(t, func() {
		a.Insert(1, -1)
	}, "index out of range")

	assert.Panics(t, func() {
		a.Insert(1, n)
	}, "index out of range")

	// 满了，从非最后位置插入，则前面元素后移
	a.Insert(2, 0)
	assert.Equal(t, a.Data(), []interface{}{2, 1})
}

func TestArray_Insert2(t *testing.T) {
	n := 3
	a := NewArray(n)
	assert.Equal(t, 0, a.Len())

	a.Insert(3)
	a.Insert(2)
	a.Insert(1) // 1,2,3

	assert.Equal(t, 3, a.Len())
	assert.Equal(t, a.Data(), []interface{}{1, 2, 3})

	// 已满，从最后插入，则前面元素不变，直接替换最后一个元素
	a.Insert(4, 2)
	assert.Equal(t, a.Data(), []interface{}{1, 2, 4})

	// 已满，从中间插入，则从插入位置元素后移
	a.Insert(3, 1)
	assert.Equal(t, a.Data(), []interface{}{1, 3, 2})

	// 已满，从开头插入，则从插入位置元素后移
	a.Insert(0, 0)
	assert.Equal(t, a.Data(), []interface{}{0, 1, 3})
}

func TestArray_FindByIndex(t *testing.T) {
	a := NewArray(3)
	assert.Panics(t, func() {
		a.FindByIndex(0)
	}, "index out of range")

	a.Insert(3)
	assert.Equal(t, a.FindByIndex(0), 3)
	a.Insert(2) // 2,3
	assert.Equal(t, a.FindByIndex(0), 2)
	assert.Equal(t, a.FindByIndex(1), 3)
}

func TestArray_DeleteByIndex(t *testing.T) {
	a := NewArray(3)
	assert.Panics(t, func() {
		a.DeleteByIndex(0)
	}, "index out of range")

	a.Insert(3)
	a.DeleteByIndex(0)
	assert.Equal(t, 0, a.Len())

	// 删除最后一个元素
	a.Insert(3)
	a.Insert(2)
	a.DeleteByIndex(1)
	assert.Equal(t, a.Data(), []interface{}{2})
	assert.Equal(t, 1, a.Len())

	// 删除中间元素
	a.Insert(1)
	a.Insert(0) // 0,1,2
	assert.Equal(t, a.Data(), []interface{}{0, 1, 2})
	a.DeleteByIndex(1)
	assert.Equal(t, a.Data(), []interface{}{0, 2})
	assert.Equal(t, 2, a.Len())

	// 删除第一个元素
	// 0,2
	a.DeleteByIndex(0)
	assert.Equal(t, a.Data(), []interface{}{2})
	assert.Equal(t, 1, a.Len())
}
