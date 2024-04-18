package main

import (
	"time"
)

func exampleVisitor() {
	println("exampleVisitor")
	order := &Order{
		Items: []OrderItem{
			&Electronic{Name: "Laptop", Price: 1200, Count: 1},
			&Clothing{Name: "Shirt", Price: 50, Count: 2},
		},
	}

	// 折扣访问者为每个订单项赠送折扣
	discountCalc := &DiscountCalculator{}
	order.Execute(discountCalc)

	// 价格访问者计算所有订单项总价
	calc := &PriceCalculator{}
	order.Execute(calc)
	//fmt.Printf("Total Price:%.2f\n", calc.TotalPrice)

	//通过使用访问者模式，我们可以在不修改订单项(的方法)的情况下，灵活地添加新的操作，比如计算折扣、计算总价等。
	//这提高了系统的可扩展性和可维护性。
}

// OrderItem 定义了订单项接口
type OrderItem interface {
	Accept(v Visitor)
}

// Electronic 作为具体订单项
type Electronic struct {
	Name         string
	Price        float64
	Discount     float64 // 用户折扣券
	GiftDiscount float64 // 赠送折扣，初始为0
	Count        int

	state string
}

// GetState 获取订单项状态（作为被访问的元素，通常还有拥有其他方法）
func (e *Electronic) GetState() string {
	return e.state
}

// Accept 接受访问者（若不希望被访问者修改内部状态，可传值给访问者）
func (e *Electronic) Accept(v Visitor) {
	// 调用对应元素的访问方法
	v.VisitElectronic(e)
}

// Clothing 作为具体订单项
type Clothing struct {
	Name         string
	Price        float64
	Count        int
	Discount     float64 // 用户折扣券
	GiftDiscount float64 // 赠送折扣，初始为0
	TotalPrice   float64
}

// Accept 接受访问者
func (c *Clothing) Accept(v Visitor) {
	// 调用对应元素的访问方法
	v.VisitClothing(c)
}

// Visitor 定义了访问者接口，包含了对各种元素类型的访问方法
// -- 可能有一个或多个访问者实现，分别拥有不同的功能，访问者的逻辑变化与元素方法是解耦的，除非元素内部字段删减
type Visitor interface {
	VisitElectronic(o *Electronic)
	VisitClothing(o *Clothing)
}

// PriceCalculator 作为具体访问者，计算总价
type PriceCalculator struct {
	TotalPrice float64
}

func (pc *PriceCalculator) VisitElectronic(o *Electronic) {
	pc.TotalPrice += o.Price * float64(o.Count) * (1 - o.Discount - o.GiftDiscount)
}

func (pc *PriceCalculator) VisitClothing(o *Clothing) {
	pc.TotalPrice += o.Price * float64(o.Count) * (1 - o.Discount - o.GiftDiscount)
}

// DiscountCalculator 作为具体访问者，赠送折扣
type DiscountCalculator struct{}

func (dc *DiscountCalculator) VisitElectronic(e *Electronic) {
	if time.Now().Format("01-02") == "11-11" {
		e.GiftDiscount = 0.1 // 假设双十一当天赠送10%的折扣
		return
	}
	e.GiftDiscount = 0.05 // 日常电子产品赠送0.05%的折扣
}

func (dc *DiscountCalculator) VisitClothing(c *Clothing) {
	c.GiftDiscount = 0.2 // 服装固定赠送20%的折扣
}

// Order 作为对象结构
type Order struct {
	Items []OrderItem
}

// Execute 遍历访问对象结构
func (o *Order) Execute(v Visitor) {
	for _, item := range o.Items {
		item.Accept(v)
	}
}
