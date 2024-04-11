package main

/*
工厂方法模式是简单工厂模式的进一步规范化，此模式定义了工厂类来产出产品。
-- 即不同的产品由不同的工厂来创建。
*/

func exampleUsage() {
	var creatorA ProductFactory = NewProductAFactory()
	productA := creatorA.Create()
	productA.GetPrice()
}

// 产品类

type Product interface {
	GetPrice() int
}

// 产品工厂接口

type ProductFactory interface {
	Create() Product
}

// ---------------------------------------
// 产品A的工厂类（实现产品工厂接口）

type ProductAFactory struct{}

func (*ProductAFactory) Create() Product {
	return &ProductA{100}
}

func NewProductAFactory() *ProductAFactory {
	return &ProductAFactory{}
}

// 产品A具体实现

type ProductA struct {
	price int
}

func (p *ProductA) GetPrice() int { return p.price }

//------------------------------------------
// 产品B的工厂类

/*
type ProductBFactory struct{}

func (*ProductBFactory) Create() Product {
	return &ProductB{100}
}

...
*/
