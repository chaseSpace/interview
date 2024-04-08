package main

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

/*
  sync.Cond 是一种高级并发模式，相比于chan的简单的一个G通知另一个G的并发模型，
sync.Cond可以根据情况决定一个G通知一个或多个G。其中Broadcast函数会通知所有等待的G，而Signal函数只会通知一个等待的G。
根据实际情况决定调用Broadcast或Signal。

- sync.Cond 实例必须关联一个Locker，并在改变条件状态时或调用Wait时持有锁，确保对条件变量的操作上互斥的。
- 在标准库的h2实现中，sync.Cond 用于在需要关闭连接时通知所有打开的G
*/

func TestSyncCond(t *testing.T) {

	// 多个消费者启动后将会进入阻塞，等待上游通知
	for i := 0; i < consumers; i++ {
		go consumer(fmt.Sprintf("r%d", i))
	}

	producer()
	time.Sleep(time.Second)

}

// Locker可以是sync.Mutex或sync.RWMutex
var c = sync.NewCond(&sync.Mutex{})

// 标识上游生产是否完成，修改时必需要加锁
var produceDone bool
var consumers = 3

func producer() {
	t.Log("生产中。。。")
	time.Sleep(time.Second)

	// 生产完成，通知消费者
	c.L.Lock()
	produceDone = true
	c.L.Unlock()
	c.Broadcast()
}

func consumer(reader string) {
	c.L.Lock()

	// 重点解释
	// - 这里的阻塞点不在于无限循环，而在于c.Wait()
	// - 因为常规状态下for循环条件为真，所以会直接进入Wait，其内部实现了一个高效的阻塞机制
	for !produceDone { // 上游空闲时进入Wait阻塞
		c.Wait() // Wait结束时，还会判断一次条件，在条件为假时向下执行
	}

	fmt.Printf("%s消费中。。。\n", reader)
	c.L.Unlock()
}
