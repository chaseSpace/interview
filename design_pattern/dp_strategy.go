package main

func exampleStrategy() {
	println("exampleStrategy")

	// 在导航系统的上下文中，根据用户选择使用不同导航策略
	strategy := ChooseNavigationStrategy("walk")
	strategy.BuildRoute("A", "B")

	strategy = ChooseNavigationStrategy("bicycle")
	strategy.BuildRoute("A", "B")
}

// 定义导航策略接口

type NavigationStrategy interface {
	BuildRoute(startPos, endPos string)
}

var _ NavigationStrategy = (*WalkingNavigation)(nil)
var _ NavigationStrategy = (*BicycleNavigation)(nil)

// 实现导航策略——步行

type WalkingNavigation struct{}

func (w *WalkingNavigation) BuildRoute(startPos, endPos string) {
	// ...
}

// 实现导航策略——自行车

type BicycleNavigation struct{}

func (b *BicycleNavigation) BuildRoute(startPos, endPos string) {
	// ...
}

func ChooseNavigationStrategy(userChoice string) NavigationStrategy {
	switch userChoice {
	case "walk":
		return &WalkingNavigation{}
	case "bicycle":
		return &BicycleNavigation{}
	default:
		panic("not supported navigation strategy")
	}
}
