package main

import "fmt"

func exampleBridge() {
	println("exampleBridge")
	s := NewSMSBridge(NewSMSAPIImpl())

	_ = s.Send2Domestic("test", "13811111111")
	_ = s.Send2National("test", "85281111111")
}

// 定义短信抽象接口（抽象方）,直接被业务方调用

type SMSService interface {
	Send2Domestic(text, to string) error // 发送国内号码
	Send2National(text, to string) error // 发送国际号码
}

// 实现短信函数
// -- 函数负责实现功能即可, 抽象方不管实现方细节

type SMSAPIImpl struct{}

func (S SMSAPIImpl) Send2Domestic(text, to string) error {
	// 调用服务商接口发送短信。。
	return nil
}

func NewSMSAPIImpl() *SMSAPIImpl {
	return &SMSAPIImpl{}
}

// 桥接方的责任是将抽象接口与实现部分连接起来

var _ SMSService = (*SMSBridge)(nil)

type SMSBridge struct {
	s *SMSAPIImpl
}

func NewSMSBridge(s *SMSAPIImpl) *SMSBridge {
	return &SMSBridge{s: s}
}

func (s *SMSBridge) Send2Domestic(text, to string) error {
	return s.s.Send2Domestic(text, to)
}

func (s *SMSBridge) Send2National(text, to string) error {
	return fmt.Errorf("not support yet")
}
