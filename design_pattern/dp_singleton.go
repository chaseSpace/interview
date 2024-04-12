package main

import "sync"

func exampleSingleton() {
	println("exampleSingleton")

	obj := SingletonMgrInstance.NewProduct()
	obj.DoSomething()
}

type UniqueProduct struct {
}

type SingletonMgr struct {
	p *UniqueProduct
	o sync.Once
}

var SingletonMgrInstance = new(SingletonMgr)

func (s *SingletonMgr) NewProduct() *UniqueProduct {
	// 通过某种机制保证实例的唯一性
	s.o.Do(func() {
		s.p = &UniqueProduct{}
	})
	return s.p
}

func (*UniqueProduct) DoSomething() {}
