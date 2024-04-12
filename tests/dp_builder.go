package main

import "fmt"

func exampleBuilder() {
	println("exampleBuilder")

	builder := NewMacBookBuilder()
	builder.BuildRam("8Gi")
	builder.BuildCpu("12Core")

	mac := builder.GetComputer()
	mac.PrintConfig()
}

// ----- 定义产品接口

type Computer interface {
	PrintConfig()
}

// ----- 定义 *Builder 接口（必须包含build方法+Get产品的方法）

type ComputerBuilder interface {
	BuildCpu(string)
	BuildRam(string)
	GetComputer() Computer
}

// ----- 具体产品实现（和 *Builder 类配套）
// -- 没有自己的New方法，由 *Builder 类创建

type MacBookComputer struct {
	cpu, ram string
}

func (m *MacBookComputer) PrintConfig() {
	fmt.Printf("cpu:%s ram:%s\n", m.cpu, m.ram)
}

// ----- 具体产品的 *Builder 实现

type MacBookBuilder struct {
	MacBookComputer
}

func NewMacBookBuilder() ComputerBuilder {
	return &MacBookBuilder{}
}

func (h *MacBookBuilder) BuildCpu(cpu string) {
	h.cpu = cpu
}

func (h *MacBookBuilder) BuildRam(ram string) {
	h.ram = ram
}

func (h *MacBookBuilder) GetComputer() Computer {
	return &h.MacBookComputer
}
