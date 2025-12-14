package main

import (
	"strings"
	"sync"
	"testing"
)

type MyStruct struct {
	noCopy sync.Mutex
	n2     sync.Once
	n3     sync.WaitGroup
	n4     sync.Pool
	n5     sync.Cond
	// 其他字段...
}

// go vet 可以检测嵌入 noCopy 类型的变量是否被复制
func TestNoCopySyncLcker(t *testing.T) {
	v := MyStruct{}

	v2 := v // go vet提示：无法复制sync.Locker成员字段
	_ = v2
}

func TestNoCopyStringBuilder(t *testing.T) {
	v3 := strings.Builder{}
	v3.WriteString("x")
	v4 := v3
	v4.WriteString("2") // panic
}
