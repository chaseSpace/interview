package main

func exampleAdapter() {
	println("exampleAdapter")

	unCompatibleAPI := NewApiImpl()

	adapter := NewAdapter(unCompatibleAPI)
	adapter.CompatibleMethod()
}

// 需要适配的目标接口

type UnCompatibleAPI interface {
	UnCompatibleMethod()
}

var _ UnCompatibleAPI = (*ApiImpl)(nil)

// 目标接口实现

type ApiImpl struct{}

func (a *ApiImpl) UnCompatibleMethod() {
	//println("UnCompatibleMethod...")
}

func NewApiImpl() *ApiImpl {
	return &ApiImpl{}
}

// 适配器

type Adapter struct {
	UnCompatibleAPI
}

func (a *Adapter) CompatibleMethod() {
	// 适配时通常需要做兼容性处理
	// ...

	// 然后调用目标接口的方法
	a.UnCompatibleMethod()
}

func NewAdapter(api UnCompatibleAPI) *Adapter {
	return &Adapter{api}
}
