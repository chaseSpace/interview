package main

// 抽象工厂类 --产出--> （多个）工厂类 --产出--> （多个）实例

// 说明: 此模式的抽象程度较高（较难理解），业务中极少使用，一般在第三方工具库（如数据库）中使用。

func exampleAbsFactory() {
	println("exampleAbsFactory")

	samsungAbs := NewSamsungPhoneAbstractFactory() // 创建 samsung 的抽象工厂实例

	screen := samsungAbs.CreateScreen().Create()   // 创建 samsung 的屏幕实例
	battery := samsungAbs.CreateBattery().Create() // 创建 samsung 的电池实例

	screen.ScreenMethod()   // 调用 samsung 的屏幕方法
	battery.BatteryMethod() // 调用 samsung 的电池方法
}

// ------------------- 定义抽象工厂类（产出具体工厂类）

type PhoneAbstractFactory interface {
	CreateBattery() BatteryFactory
	CreateScreen() ScreenFactory
}

//  BatteryFactory 和 ScreenFactory 是两种产品的工厂类（产出具体产品）

type BatteryFactory interface {
	Create() BatteryAPI
}

type ScreenFactory interface {
	Create() ScreenAPI
}

// BatteryAPI 和 ScreenAPI 是两个产品接口

type BatteryAPI interface {
	BatteryMethod()
}
type ScreenAPI interface {
	ScreenMethod()
}

// ------------------- 实现 Samsung 的抽象工厂类

type SamsungPhoneAbstractFactory struct{}

func (*SamsungPhoneAbstractFactory) CreateBattery() BatteryFactory { return &SamsungBatteryFactory{} }
func (*SamsungPhoneAbstractFactory) CreateScreen() ScreenFactory   { return &SamsungScreenFactory{} }

func NewSamsungPhoneAbstractFactory() PhoneAbstractFactory {
	return &SamsungPhoneAbstractFactory{}
}

// ------------------- 实现 Samsung 的 Battery、Screen 的工厂类（产出实例）

type SamsungBatteryFactory struct{}

func (*SamsungBatteryFactory) Create() BatteryAPI { return &SamsungBattery{} }

type SamsungScreenFactory struct{}

func (*SamsungScreenFactory) Create() ScreenAPI { return &SamsungScreen{} }

// ------------------- 实现 Samsung 的 Battery、Screen

type SamsungBattery struct{}

func (*SamsungBattery) BatteryMethod() {}

type SamsungScreen struct{}

func (*SamsungScreen) ScreenMethod() {}

/*
Xiaomi的一系列实现

type XiaomiPhoneAbstractFactory struct{}

func (*XiaomiPhoneAbstractFactory) CreateBattery() BatteryFactory {...}
func (*XiaomiPhoneAbstractFactory) CreateScreen() ScreenFactory {...}

...
*/
