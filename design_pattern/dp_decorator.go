package main

// 在Java和Python中可以通过为函数添加注解的方式来直观的实现装饰器模式。
// -- 但是Golang中没有注解的概念，只能通过新增一个装饰器对象来封装原始对象来实现装饰器模式。

func exampleDecorator() {
	println("exampleDecorator")
	obj := NewConcreteProduct()
	//obj.DoSomething()

	logDec := NewDecorator(obj)
	logDec.DoSomething()
}

type ProductAPI interface {
	DoSomething()
}

var _ ProductAPI = (*ConcreteProduct)(nil)

type ConcreteProduct struct{}

func (p *ConcreteProduct) DoSomething() {
	// ...
}

func NewConcreteProduct() *ConcreteProduct {
	return &ConcreteProduct{}
}

type ProductLoggingDecorator struct {
	wrapped ProductAPI
}

// 装饰器需要实现目标对象接口

func (d *ProductLoggingDecorator) DoSomething() {
	// 在执行前做一些操作，例如logging
	// println("logging %s before DoSomething()", time.Now())

	d.wrapped.DoSomething()
	// 在执行后做一些操作
}

// 装饰器的工厂函数要返回相同的产品类（接口）

func NewDecorator(wrapped ProductAPI) ProductAPI {
	return &ProductLoggingDecorator{wrapped}
}
