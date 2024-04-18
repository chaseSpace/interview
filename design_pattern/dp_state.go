package main

import "fmt"

func exampleState() {
	orderContext := NewOrderContext()

	// 新订单->确认 √
	err := orderContext.ConfirmOrder()
	assertNil(err)

	// 确认->取消 √
	err = orderContext.CancelOrder()
	assertNil(err)

	// 取消后无法发货 X
	err = orderContext.ShipOrder()
	assertEqual(err, fmt.Errorf("已经取消的订单无法发货"))

	// 新订单2
	orderContext = NewOrderContext()

	// 新订单无法直接发货 X
	err = orderContext.ShipOrder()
	assertEqual(err, fmt.Errorf("订单无法发货，需要先确认订单"))

	// 新订单->确认 √
	err = orderContext.ConfirmOrder()
	assertNil(err)
	// 确认->发货 √
	err = orderContext.ShipOrder()
	assertNil(err)
	// 发货->完成 √
	err = orderContext.CompletedOrder()
	assertNil(err)
}

// OrderStatus 定义订单状态的接口（每个方法都是一种状态）
type OrderStatus interface {
	ConfirmOrder(ctx *OrderContext) error   // 已确认
	ShipOrder(ctx *OrderContext) error      // 已发货
	CompletedOrder(ctx *OrderContext) error // 已完成
	CancelOrder(ctx *OrderContext) error    // 已取消
}

// State 定义订单状态枚举
type State int8

const (
	NewState       State = iota
	ConfirmedState       = iota
	ShippedState         = iota
	CompletedState       = iota
	CancelledState       = iota
)

// OrderContext 订单上下文，包含当前订单状态
type OrderContext struct {
	orderStatus OrderStatus
	// 上下文中拥有订单状态字段，状态变更受限于当前状态
	state State
}

// NewOrderContext 创建一个新订单上下文
func NewOrderContext() *OrderContext {
	return &OrderContext{orderStatus: &NewOrder{}}
}

// SetOrderStatus 设置订单状态
func (oc *OrderContext) SetOrderStatus(orderStatus OrderStatus) {
	oc.orderStatus = orderStatus
}

// ConfirmOrder 将订单状态设置为已确认
func (oc *OrderContext) ConfirmOrder() error {
	return oc.orderStatus.ConfirmOrder(oc)
}

// ShipOrder 将订单状态设置为已发货
func (oc *OrderContext) ShipOrder() error {
	return oc.orderStatus.ShipOrder(oc)
}

// CompletedOrder 将订单状态设置为已完成
func (oc *OrderContext) CompletedOrder() error {
	return oc.orderStatus.CompletedOrder(oc)
}

// CancelOrder 将订单状态设置为已取消
func (oc *OrderContext) CancelOrder() error {
	return oc.orderStatus.CancelOrder(oc)
}

// NewOrder 新建订单状态
type NewOrder struct{}

func (no *NewOrder) ConfirmOrder(ctx *OrderContext) error {
	ctx.state = ConfirmedState
	ctx.orderStatus = &ConfirmedOrder{}
	return nil
}

func (no *NewOrder) ShipOrder(ctx *OrderContext) error {
	return fmt.Errorf("订单无法发货，需要先确认订单")
}

func (no *NewOrder) CompletedOrder(ctx *OrderContext) error {
	return fmt.Errorf("订单无法完成，需要先确认订单")
}

func (no *NewOrder) CancelOrder(ctx *OrderContext) error {
	ctx.state = CancelledState
	ctx.orderStatus = &CanceledOrder{}
	return nil
}

// ConfirmedOrder 已确认订单状态
type ConfirmedOrder struct{}

func (co *ConfirmedOrder) ConfirmOrder(ctx *OrderContext) error {
	return fmt.Errorf("订单无法确认，已经确认")
}

func (co *ConfirmedOrder) ShipOrder(ctx *OrderContext) error {
	ctx.state = ShippedState
	ctx.orderStatus = &ShippedOrder{}
	return nil
}

func (co *ConfirmedOrder) CompletedOrder(ctx *OrderContext) error {
	return fmt.Errorf("订单无法完成，需要先发货")
}

func (co *ConfirmedOrder) CancelOrder(ctx *OrderContext) error {
	ctx.state = CancelledState
	ctx.orderStatus = &CanceledOrder{}
	return nil
}

// ShippedOrder 已发货订单状态
type ShippedOrder struct{}

func (*ShippedOrder) ConfirmOrder(ctx *OrderContext) error {
	return fmt.Errorf("订单无法确认，已经发货")
}

func (*ShippedOrder) ShipOrder(ctx *OrderContext) error {
	return fmt.Errorf("已经发货")
}

func (*ShippedOrder) CompletedOrder(ctx *OrderContext) error {
	ctx.state = CompletedState
	ctx.orderStatus = &CompletedOrder{}
	return nil
}

func (*ShippedOrder) CancelOrder(ctx *OrderContext) error {
	return fmt.Errorf("订单无法取消，已经发货")
}

// CompletedOrder 已完成订单状态
type CompletedOrder struct{}

func (*CompletedOrder) ConfirmOrder(ctx *OrderContext) error {
	return fmt.Errorf("已经完成的订单无法确认")
}

func (*CompletedOrder) ShipOrder(ctx *OrderContext) error {
	return fmt.Errorf("已经完成的订单无法发货")
}

func (*CompletedOrder) CompletedOrder(ctx *OrderContext) error {
	return fmt.Errorf("已经完成的订单无法完成")
}

func (*CompletedOrder) CancelOrder(ctx *OrderContext) error {
	return fmt.Errorf("已经完成的订单无法取消")
}

// CanceledOrder 已完成订单状态
type CanceledOrder struct{}

func (*CanceledOrder) ConfirmOrder(ctx *OrderContext) error {
	return fmt.Errorf("已经取消的订单无法确认")
}

func (*CanceledOrder) ShipOrder(ctx *OrderContext) error {
	return fmt.Errorf("已经取消的订单无法发货")
}

func (*CanceledOrder) CompletedOrder(ctx *OrderContext) error {
	return fmt.Errorf("已经取消的订单无法完成")
}

func (*CanceledOrder) CancelOrder(ctx *OrderContext) error {
	return fmt.Errorf("已经取消的订单无法取消")
}
