package main

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLinkedList(t *testing.T) {
	a := NewLinkedList()
	assert.Equal(t, []interface{}{}, a.Data())
	assert.Equal(t, 0, a.Len())

	a.Insert(3)
	assert.Equal(t, []interface{}{3}, a.Data())
	assert.Equal(t, 1, a.Len())

	a.Insert(2)
	assert.Equal(t, []interface{}{2, 3}, a.Data())
	assert.Equal(t, 2, a.Len())

	a.Insert(1) // 1,2,3
	assert.Equal(t, []interface{}{1, 2, 3}, a.Data())

	// Remove
	a.Remove(2) // 1,2
	assert.Equal(t, []interface{}{1, 2}, a.Data())
	assert.Equal(t, 2, a.Len())

	assert.Panics(t, func() {
		a.Remove(2)
	}, "index out of range")

	a.Remove(1) // 1
	assert.Equal(t, []interface{}{1}, a.Data())
	assert.Equal(t, 1, a.Len())

	a.Remove(0) // []
	assert.Equal(t, []interface{}{}, a.Data())
	assert.Equal(t, 0, a.Len())
}
