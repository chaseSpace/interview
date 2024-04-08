package main

import (
	"fmt"
	"strings"
	"testing"
)

func TestNewLiteral(t *testing.T) {
	utf8_str := "放下"               // 字符串字面量
	_ = "\xe6\x94\xbe\xe4\xb8\x8b" // 字节字面量，\x后面是16进制的编码
	_ = "\u653E\u4E0B"             // unicode字面量

	// 它们是完全相等的。

	fmt.Printf("%x\n", utf8_str)    // e694bee4b88b
	fmt.Printf("%q\n", utf8_str)    // "\u653e\u4e0b"  输出ASCII编码的字符，双引号包含
	fmt.Printf("%U\n", utf8_str[0]) // U+00E6   打印单个字节的unicode编码
}

func TestRune(t *testing.T) {
	var runeLiteral rune = 'a'
	println(runeLiteral)            // 97
	fmt.Printf("%x\n", runeLiteral) // 61（16进制）
	fmt.Printf("%v\n", runeLiteral) // 值的默认格式（万能）
	fmt.Printf("%c\n", runeLiteral) // 将rune转换为Unicode码点所表示的字符

	println(len("a菜"), len([]rune("a菜"))) // 4  2
}

func TestTrim(t *testing.T) {
	var s = strings.Trim("-Hell-o-", "-")
	t.Log(s, len(s)) // Hell-o 6

	s = strings.Trim("-Love-o-", "-o")
	t.Log(s, len(s)) // Love 4
}

func TestTrimLeft(t *testing.T) {
	var s = strings.TrimLeft("-Hell-o-", "-o")
	t.Log(s, len(s)) // Hell-o- 7
}

func TestTrimRight(t *testing.T) {
	var s = strings.TrimRight("-Hell-o-", "-o")
	t.Log(s, len(s)) // -Hell 5
}

func TestTrimSpace(t *testing.T) {
	// 清除两侧的空白符
	var s = strings.TrimSpace(" - -\t\n\v\f\r ")
	t.Log(s, len(s)) // - -  3
}
