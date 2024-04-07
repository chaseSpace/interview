package tests

import (
	"testing"
)

type noCopy struct{}

func (*noCopy) Lock()   {}
func (*noCopy) Unlock() {}

type MyStruct struct {
	noCopy noCopy
	//sync.Locker
	// 其他字段...
}

// go vet 可以检测嵌入 noCopy 类型的变量是否被复制
func TestNoCopyStringBuilder(t *testing.T) {
	v := MyStruct{}

	// copy it
	v2 := v
	_ = v2
}
