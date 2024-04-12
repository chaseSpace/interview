package main

import "sync"

func exampleProxy() {
	println("exampleProxy")

	// 此代理为原对象提供延迟初始化功能
	proxy := NewProxy()
	proxy.DoSomething()
}

type ProductDAPI interface {
	DoSomething()
}

var _ ProductDAPI = (*ConcreteProductD)(nil)

type ConcreteProductD struct{}

func (p *ConcreteProductD) DoSomething() {
	// ...
}

func NewConcreteProductD() *ConcreteProductD {
	return &ConcreteProductD{}
}

type Proxy struct {
	wrapped ProductDAPI
	o       sync.Once
}

func (d *Proxy) DelayInit() {
	d.o.Do(func() {
		d.wrapped = NewConcreteProductD()
	})
}

// 代理需要实现目标对象接口

func (d *Proxy) DoSomething() {
	d.DelayInit()
	d.wrapped.DoSomething()
}

// 装饰器的工厂函数要返回相同的产品类（接口）

func NewProxy() ProductDAPI {
	return &Proxy{}
}
