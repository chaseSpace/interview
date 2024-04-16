package main

// Go没有Java等面向对象语言中的继承机制，但可以通过接口和组合来实现类似的效果（实现方式与Java有所不同）

func exampleTemplateMethod() {
	println("exampleTemplateMethod")

	// 初始化咖啡机
	cm := NewCoffeeMachine()

	// 替换插件（指定咖啡类型，如浓缩）
	cm.UpdatePlugin(NewEspressoMachine())

	// 制作
	cm.makeCoffee()
}

type CoffeeMachine struct {
	plugin CoffeeMakingStepPlugin
}

// 制作步骤不可变，仅可改变部分步骤的实现
func (c *CoffeeMachine) makeCoffee() {
	c.plugin.grindBeans()
	c.boilWater()
	c.plugin.brew()
	c.pourInCup()
	c.plugin.addMilk()
}

func (c *CoffeeMachine) UpdatePlugin(plugin CoffeeMakingStepPlugin) {
	c.plugin = plugin
}

func (c *CoffeeMachine) boilWater() {

}
func (c *CoffeeMachine) pourInCup() {

}

func NewCoffeeMachine() CoffeeMachine {
	return CoffeeMachine{}
}

//  定义可替换的咖啡制作接口

type CoffeeMakingStepPlugin interface {
	grindBeans() // 磨豆
	brew()       // 冲泡
	addMilk()    // 加奶
}

var _ CoffeeMakingStepPlugin = (*EspressoMachine)(nil)

// 实现咖啡制作接口——浓缩咖啡

type EspressoMachine struct {
}

func (e *EspressoMachine) grindBeans() {
	// 磨浓缩咖啡豆子
}

func (e *EspressoMachine) brew() {
	// 冲泡浓缩咖啡
}

func (e *EspressoMachine) addMilk() {
	// 浓缩咖啡不加奶
}

func NewEspressoMachine() CoffeeMakingStepPlugin {
	return &EspressoMachine{}
}

// 实现咖啡制作接口——拿铁咖啡。。
