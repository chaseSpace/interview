package main

import (
	"time"
)

func exampleMemento() {
	println("exampleMemento")

	u := User{ms: make(map[string]Memento)}

	gs := &GameSystem{
		position: 1,
		life:     3,
	}
	u.AddMemento("snapshot1", gs.Save())

	// reset
	gs.position = 0
	gs.life = 0

	gs.Restore(u.GetMemento("snapshot1"))
	//fmt.Printf("%+v\n", gs)
}

// 定义备忘录接口（可选）
// -- 备忘录仅提供受限的信息

type Memento interface {
	GetCreateDate() string // 创建时间
}

var _ Memento = (*MementoImpl)(nil)

// 实现备忘录接口（包含需要备份的状态信息）

type MementoImpl struct {
	createDate string

	position int
	life     int
}

func (m *MementoImpl) GetCreateDate() string {
	return m.createDate
}

// -- 通过私有方法对发起人提供所备份的状态信息，外部用户无法调用
func (m *MementoImpl) getPosition() int {
	return m.position
}

// -- 备忘录提供构造函数给发起人调用，以保存状态信息

func NewMemento(pos, life int) Memento {
	return &MementoImpl{
		createDate: time.Now().String(),
		position:   pos,
		life:       life,
	}
}

// 定义发起人：游戏系统
// -- 发起人包含各种需要保存的状态信息，如玩家的位置、生命值等
// -- 发起人负责提供创建备忘录（即快照）的方法

type GameSystem struct {
	position int
	life     int
}

func (g *GameSystem) Save() Memento {
	return NewMemento(g.position, g.life)
}

func (g *GameSystem) Restore(m Memento) {
	// 强制转为 MementoImpl 类型，才能获取所备份的状态信息
	memento, ok := m.(*MementoImpl)
	if !ok {
		panic("invalid memento")
	}
	g.position = memento.position
	g.life = memento.life
}

// 定义负责人：用户
// -- 用户负责管理备忘录
// -- 用户与备忘录接口交互，无法获取所备份的状态信息
// -- 用户与发起人和备忘录属于不同的pkg，所以无法调用备忘录的私有方法

type User struct {
	ms map[string]Memento
}

func (c *User) AddMemento(name string, m Memento) {
	c.ms[name] = m
}

func (c *User) GetMemento(name string) Memento {
	return c.ms[name]
}
