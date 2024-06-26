package main

func main() {
	runDesignPatternExample()
}

func runDesignPatternExample() {
	// 6种创建型模式
	exampleSimpleFactory()
	exampleFactoryMethod()
	exampleAbsFactory()
	exampleBuilder()
	examplePrototype()
	exampleSingleton()

	// 7种结构型模式
	exampleAdapter()
	exampleDecorator()
	exampleProxy()
	exampleFacade()
	exampleBridge()
	exampleComposite()
	exampleFlyweight()

	// 11种行为型模式
	exampleStrategy()
	exampleTemplateMethod()
	exampleObserver()
	exampleIterator()
	exampleChainOfResponsibility()
	exampleCommand()
	exampleMemento()
	exampleState()
	exampleVisitor()
	exampleMediator()
	exampleInterpreter()

	// 其他模式
	exampleNullObject()
	exampleInterceptor()
}
