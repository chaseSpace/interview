package main

func exampleComposite() {
	println("exampleComposite")
	root := NewCompositeNode("root")
	root.Add(NewLeafNode("file1", true))

	dir := NewCompositeNode("dir1")
	dir.Add(NewLeafNode("file2 inside of dir1", true))
	root.Add(dir)

	// 统计所有文件夹和文件总数
	if root.Count() != 4 {
		panic("root.Count() != 4")
	}
}

// 组合模式中通常先定义一个 Component 接口作为节点对象
// -- 节点对象包含组合节点和叶子节点的公共方法
// -- 代表一个文件(夹)

type Component interface {
	Name() string
	Count() int
}

var _ Component = (*LeafNode)(nil)
var _ Component = (*CompositeNode)(nil)

// 实现叶子节点
// -- 代表一个空文件夹或一个文件

type LeafNode struct {
	name   string
	isfile bool
}

func (l *LeafNode) Count() int {
	return 1 // only self
}

func (l *LeafNode) Name() string {
	return l.name
}

func NewLeafNode(name string, isfile bool) Component {
	return &LeafNode{
		name:   name,
		isfile: isfile,
	}
}

// 实现组合节点，组合节点通常拥有两个方法
// - Add() 增加叶子节点
// - Remove() 删除叶子节点

// CompositeNode 在文件系统场景中，代表文件夹
type CompositeNode struct {
	name     string
	children []Component
}

func (c *CompositeNode) Count() int {
	var count int
	for _, cc := range c.children {
		if _, ok := cc.(*LeafNode); ok {
			count++
		} else {
			count += cc.Count() // 递归计算子目录文件(夹)数
		}
	}
	return count + 1 // 加上文件夹自己
}

func (c *CompositeNode) Name() string {
	return c.name
}

func (c *CompositeNode) Add(cc Component) {
	c.children = append(c.children, cc)
}

func (c *CompositeNode) Remove(rc Component) {
	for i, cc := range c.children {
		if cc.Name() == rc.Name() {
			c.children = append(c.children[:i], c.children[i+1:]...)
			return
		}
	}
}

func NewCompositeNode(name string) *CompositeNode {
	return &CompositeNode{
		name:     name,
		children: make([]Component, 0),
	}
}
