package main

// 数组实现
/*
-- 数组包含一组连续存储的相同类型的元素，支持 O(1)随机访问，O(n)查找、插入和删除操作。
-- 数组一般不支持扩容，删除指定元素也是O(n)，倒序排列也是O(n)。
*/

type Array struct {
	data []interface{} // 使用切片模拟内存中的连续地址
	len  int           // 只能自己维护数组元数据，不能调用切片的内置函数len()
}

var _ ListAPI = (*Array)(nil)

func NewArray(capacity int) Array {
	// 注意：这里必需初始化一个长度等于容量的切片，才能完全模拟内存中的连续地址
	return Array{data: make([]interface{}, capacity)}
}

func (a *Array) Data() (data []interface{}) {
	data = make([]interface{}, a.len)
	for i := 0; i < a.len; i++ {
		data[i] = a.data[i]
	}
	return
}

// Get O(1)
func (a *Array) Get(idx int) interface{} {
	if idx < 0 || idx >= a.len {
		panic("index out of range")
	}
	return a.data[idx]
}

// Insert O(n)
func (a *Array) Insert(elem interface{}, idx ...int) {
	_idx := 0
	if len(idx) == 0 {
		_idx = 0
	} else {
		_idx = idx[0]
	}
	if _idx < 0 || _idx > cap(a.data)-1 { // 允许在容量范围内插入任意位置
		panic("index out of range")
	}

	// 满了，则从插入位置依次后移
	if a.len == cap(a.data) {
		for i := a.len - 1; i > _idx; i-- {
			a.data[i] = a.data[i-1]
		}
		a.data[_idx] = elem
	} else {
		// 未满
		if _idx > a.len-1 { // 未满位置插入，直接追加
			a.data[a.len] = elem
		} else { // 插入已填空间，依次后移（从最后一个元素的后一个开始）
			for i := a.len; i > _idx; i-- {
				a.data[i] = a.data[i-1]
			}
			a.data[_idx] = elem
		}
		a.len++ // 长度加一
	}
}

// Remove O(n)
func (a *Array) Remove(idx int) {
	if idx < 0 || idx >= a.len {
		panic("index out of range")
	}

	// 删除位置后的所有元素前移一位
	for i := idx; i < a.len-1; i++ {
		a.data[i] = a.data[i+1]
	}

	// 长度减一
	a.len--
}

// Len 数组长度等于实际元素个数
func (a *Array) Len() int {
	return a.len
}

// Cap 固定容量
func (a *Array) Cap() int {
	return cap(a.data)
}
