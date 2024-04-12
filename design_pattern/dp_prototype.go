package main

func examplePrototype() {
	println("examplePrototype")
	obj := NewProductB()
	obj.SomeMethod()

	// 不需要原对象的构造方法，也可以创建出该对象实例
	obj2 := obj.Clone()
	obj2.(ProductBAPI).SomeMethod()
}

type Cloneable interface {
	Clone() Cloneable
}

type ProductBAPI interface {
	SomeMethod()
}

// 支持原型模式的产品必须实现 clone 接口
var _ Cloneable = (*ProductB)(nil)

var _ ProductBAPI = (*ProductB)(nil)

type ProductB struct {
	private int
	varA    int
	varB    string
}

func (p *ProductB) SomeMethod() {
	// ...
}

func NewProductB() *ProductB {
	return &ProductB{}
}

// 具体的复制行为由自己决定

func (p *ProductB) Clone() Cloneable {
	return &ProductB{
		//private: p.private  一般不复制私有变量
		varA: p.varA,
		varB: p.varB,
	}
}
