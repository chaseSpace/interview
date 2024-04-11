package main

import "sync"

// ListAPI 列表接口，可以由数组或链表实现。
type ListAPI interface {
	Data() []interface{}
	Len() int
	Insert(v interface{}, idx ...int) // O(n)
	Remove(idx int)                   // O(n)
}

var _ ListAPI = (*LinkedList)(nil)

// LinkedList 这是一个双向链表的实现
type LinkedList struct {
	head, tail *ListNode
	len        int
	mutex      sync.RWMutex
}

type ListNode struct {
	val        interface{}
	prev, next *ListNode
}

func NewLinkedList() *LinkedList {
	return &LinkedList{}
}

func (l *LinkedList) Data() []interface{} {
	data := make([]interface{}, 0, l.len)
	v := l.head
	for v != nil {
		data = append(data, v.val)
		v = v.next
	}
	return data
}

func (l *LinkedList) Len() int {
	l.mutex.RLock()
	defer l.mutex.RUnlock()
	return l.len
}

func (l *LinkedList) Insert(v interface{}, idx ...int) {
	insert := &ListNode{
		val:  v,
		next: nil,
	}
	l.mutex.Lock()
	defer l.mutex.Unlock()
	if l.head == nil {
		l.head = insert
		l.tail = insert
	} else {
		_idx := 0
		if len(idx) > 0 {
			_idx = idx[0]
		}
		n := l.head
		for i := 0; i < _idx; i++ {
			if n.next == nil {
				break
			}
			n = n.next
		}
		if n.prev == nil { // insert head
			insert.next = l.head
			l.head.prev = insert
			l.head = insert
		} else if n.next == nil { // insert tail
			insert.prev = l.tail
			l.tail.next = insert
			l.tail = insert
		} else {
			insert.prev = n.prev
			insert.next = n
			n.prev.next = insert
			n.prev = insert
		}
	}

	l.len++
}

func (l *LinkedList) Remove(idx int) {
	if idx < 0 || idx >= l.len {
		panic("index out of range")
	}
	l.mutex.Lock()
	defer l.mutex.Unlock()

	var remove = l.head
	for i := 0; i < idx; i++ {
		remove = remove.next
	}
	prev := remove.prev
	if prev == nil { // remove head
		l.head = remove.next
		if l.head != nil { // more than one node
			l.head.prev = nil
		}
	} else {
		prev.next = remove.next
	}
	next := remove.next
	if next == nil { // remove tail
		l.tail = prev
	} else {
		next.prev = prev
	}

	l.len--
}
