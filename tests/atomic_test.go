package main

import (
	"sync/atomic"
	"testing"
)

// 在测试时开启竞态检测

// go test -race -run=TestIntegerUsage
func TestIntegerUsage(t *testing.T) {
	// atomic.AddInt* 系列函数允许多个goroutine安全的同时操作同一个变量
	var i int32
	go func() { atomic.AddInt32(&i, 1) }()
	go func() { atomic.AddInt32(&i, -1) }()
}

// go test -race -run=TestAtomicNotSafe
func TestAtomicNotSafe(t *testing.T) {
	var v string

	go func() {
		v = "Hi Siri!"
	}()

	go func() {
		_ = v
	}()
}

// go test -race -run=TestAtomicSafe
func TestAtomicSafe(t *testing.T) {
	var v atomic.Value

	go func() {
		v.Store("Hi Siri!\n")
	}()

	go func() {
		hi := v.Load()
		print(hi.(string))
	}()
}
