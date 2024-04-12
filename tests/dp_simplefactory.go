package main

// 简单工厂模式中通过一个工厂类根据不同参数直接产出一个产品实例。
// -- Go 中，NewProduct() 函数就是一个工厂类，好比 Java 中的构造函数。
// -- 此模式是最简单的工厂模式。

func exampleSimpleFactory() {
	println("exampleSimpleFactory")
	api := NewProduct(1)
	api.SomeMethod()

	api2 := NewProduct(2)
	api2.SomeMethod()
}

type ProductX interface {
	SomeMethod()
}

// NewProduct 类似 Java 中的构造函数（工厂类）
func NewProduct(t int) ProductX {
	switch t {
	case 1:
		return &ProductImpl1{}
	case 2:
		return &ProductImpl2{}
	}
	panic("invalid arg t")
}

type ProductImpl1 struct{}

func (a *ProductImpl1) SomeMethod() {}

type ProductImpl2 struct{}

func (a *ProductImpl2) SomeMethod() {}
