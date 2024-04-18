package main

import "fmt"

func exampleMediator() {
	println("exampleMediator")
	user1 := &ConcreteUser{name: "User1"}
	user2 := &ConcreteUser{name: "User2"}

	chatRoom := NewChatRoom()
	chatRoom.users[user1.name] = user1
	chatRoom.users[user2.name] = user2

	_ = user1.Send("User1", "Hello, User2!", chatRoom)
	_ = user2.Send("User2", "Hi, User1!", chatRoom)
}

// Mediator 定义了中介者接口
type Mediator interface {
	Send(name, message string) error
}

// UserObj 定义了用户接口
type UserObj interface {
	Receive(message string)
	Send(name, message string, mediator Mediator) error
}

// ChatRoom 作为具体中介者
type ChatRoom struct {
	users map[string]UserObj
}

func (c *ChatRoom) Send(name, message string) error {
	if user, ok := c.users[name]; ok {
		user.Receive(message)
		return nil
	}
	return fmt.Errorf("name not found")
}

func NewChatRoom() *ChatRoom {
	return &ChatRoom{users: make(map[string]UserObj)}
}

// ConcreteUser 作为具体同事类
type ConcreteUser struct {
	name string
}

func (u *ConcreteUser) Receive(message string) {
	//fmt.Printf("%s: %s\n", u.name, message)
}

func (u *ConcreteUser) Send(name, message string, mediator Mediator) error {
	return mediator.Send(name, message)
}
