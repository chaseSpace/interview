package main

import (
	"net/http"
)

func exampleChainOfResponsibility() {
	println("exampleChainOfResponsibility")

	framework := NewFramework(&HandlePanic{}, &HandleAuth{})
	framework.Run(&http.Request{})
}

// 定义处理者接口

type Handler interface {
	Handle(req *http.Request)
	SetNext(handler Handler)
}

var _ Handler = (*HandleAuth)(nil)

// 实现处理者接口——异常捕捉

type HandlePanic struct {
	next Handler
}

func (h *HandlePanic) Handle(req *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			// 异常处理逻辑
			//fmt.Printf("HandlePanic:%v\n", err)
		}
	}()
	if h.next != nil {
		h.next.Handle(req)
	}
}

func (h *HandlePanic) SetNext(next Handler) {
	h.next = next
}

// 实现处理者接口——鉴权

type HandleAuth struct {
	next Handler
}

func (h *HandleAuth) Handle(req *http.Request) {
	// 鉴权。。。
	// 鉴权失败不会调用 next
	if req.Host == "" {
		panic("host empty") // 被第一个handler拦截
	}
}

func (h *HandleAuth) SetNext(next Handler) {
	h.next = next
}

// 定义web框架作为客户端

type Framework struct {
	middlewareChain Handler
}

func NewFramework(handlers ...Handler) *Framework {
	var f = &Framework{}
	if len(handlers) > 0 {
		h1 := handlers[0]
		for i := 1; i < len(handlers); i++ {
			h1.SetNext(handlers[i])
		}
		f.middlewareChain = h1
	}
	return f
}

func (f *Framework) Run(req *http.Request) {
	if f.middlewareChain != nil {
		f.middlewareChain.Handle(req)
	}
}
