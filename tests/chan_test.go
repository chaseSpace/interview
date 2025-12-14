package main

import "testing"

func TestReadClosedChan(t *testing.T) {
	var c = make(chan int, 1)
	close(c)

	<-c // 正常结束
}
