package main

func exampleCommand() {
	println("exampleCommand")
	// 注意command传递顺序：client -> invoker -> receiver
	// -- 但变量声明顺序相反

	tv := &ReceiverTV{}
	turnOn := &TurnOnCommand{tv}
	turnOff := &TurnOffCommand{tv}

	// client
	tvRemote := &TVRemote{}
	tvRemote.SetCommand(turnOn)
	tvRemote.ExecuteCommand()

	tvRemote.SetCommand(turnOff)
	tvRemote.ExecuteCommand()

	// 说明：
	// 1. 在实践中，可以将SetCommand、ExecuteCommand合并
	// 2. Command对象的Execute方法实现可以是一种远程调用。
	// 3. Command可以使用构造函数，并传入参数
}

// 定义接收者：TV
// -- 内部包含具体执行各项命令的方法

type ReceiverTV struct {
	isRunning bool
}

func (r *ReceiverTV) TurnOn() {
	if !r.isRunning {
		r.isRunning = true
	}
}

func (r *ReceiverTV) TurnOff() {
	if r.isRunning {
		r.isRunning = false
	}
}

// 定义TV支持的Command接口

type TvCommand interface {
	Execute()
}

var _ TvCommand = (*TurnOnCommand)(nil)
var _ TvCommand = (*TurnOffCommand)(nil)

// 实现具体的Command: TurnOnCommand

type TurnOnCommand struct {
	receiver *ReceiverTV
}

func (t *TurnOnCommand) Execute() {
	// Command对象内部调用Receiver来实际执行命令
	t.receiver.TurnOn()
}

// 实现具体的Command: TurnOffCommand

type TurnOffCommand struct {
	receiver *ReceiverTV
}

func (t *TurnOffCommand) Execute() {
	t.receiver.TurnOff()
}

// 定义Invoker：TVRemote
// -- 负责传递各种命令，一般包含方法：

type TVRemote struct {
	command TvCommand
}

func (i *TVRemote) SetCommand(command TvCommand) {
	i.command = command
}

func (i *TVRemote) ExecuteCommand() {
	i.command.Execute()
}
