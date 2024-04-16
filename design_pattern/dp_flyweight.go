package main

import "fmt"

func exampleFlyweight() {
	println("exampleFlyweight")

	// With the context of TextEditor rendering
	f := NewFlyweightFactory()

	char := f.GetCharacter("A", "BOLD")
	char2 := f.GetCharacter("A", "BOLD")
	char3 := f.GetCharacter("V", "ITALIC")

	char.Display(1)
	char2.Display(2)
	char3.Display(3)

	// 由于享元工厂的作用，客户端不需要为每个字符创建一个新的对象，从而减少了内存的使用。
	if f.TotalCharactersCreated() != 2 {
		panic("TotalCharactersCreated() != 2")
	}
}

// 定义享元（对象的）接口

type CharacterAPI interface {
	// Display 位置是外部状态
	// -- 注意：享元对象是不可变的，所以不会有setter接口
	Display(position int)
}

var _ CharacterAPI = (*Character)(nil)

// 实现享元接口

type Character struct {
	Character string
	Style     string
}

func (c *Character) Display(position int) {
	// 渲染
	_ = fmt.Sprintf("Character: %s, Style: %s, Position: %d\n", c.Character, c.Style, position)
}

// 定义享元工厂，负责创建和管理字符对象

type CharacterFactory interface {
	// GetCharacter 字符和样式是内部状态，可以被多个字符对象共享
	GetCharacter(char, style string) *Character
	TotalCharactersCreated() int
}

var _ CharacterFactory = (*CharacterFactoryImpl)(nil)

// 实现享元工厂

type CharacterFactoryImpl struct {
	characters map[string]*Character
}

func (c *CharacterFactoryImpl) TotalCharactersCreated() int {
	return len(c.characters)
}

func (c *CharacterFactoryImpl) GetCharacter(char, style string) *Character {
	key := fmt.Sprintf("%s_%s", char, style)
	// 这里为了简单起见，忽略了并发问题
	if v := c.characters[key]; v == nil {
		v = &Character{
			Character: char,
			Style:     style,
		}
		c.characters[key] = v
		return v
	} else {
		return v
	}
}

func NewFlyweightFactory() CharacterFactory {
	return &CharacterFactoryImpl{
		characters: make(map[string]*Character),
	}
}
