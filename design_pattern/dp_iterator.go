package main

func exampleIterator() {
	println("exampleIterator")

	collection := &BookCollection{
		books: []*Book{
			{"Book 1", "Author 1"},
			{"Book 2", "Author 2"},
			{"Book 3", "Author 3"},
		},
	}

	iterator := collection.CreateIterator()

	for iterator.Next() {
		//book := iterator.Current()
		//fmt.Printf("Book: %s, Author: %s\n", book.Title, book.Author)
	}
}

// 定义集合对象

type Book struct {
	Title  string
	Author string
}

type BookCollection struct {
	books []*Book
}

func (bc *BookCollection) CreateIterator() Iterator {
	return &BookIterator{collection: bc, index: 0}
}

// 定义迭代器接口

type Iterator interface {
	// Next 方法用于移动迭代器到下一个元素，并返回是否成功移动
	Next() bool
	// Current 方法用于返回当前元素
	Current() *Book
}

// 实现迭代器接口

type BookIterator struct {
	collection *BookCollection
	index      int
}

func (bi *BookIterator) Next() bool {
	if bi.index < len(bi.collection.books) {
		bi.index++
		return true
	}
	return false
}

func (bi *BookIterator) Current() *Book {
	if bi.index <= 0 || bi.index > len(bi.collection.books) {
		return nil
	}
	return bi.collection.books[bi.index-1]
}
