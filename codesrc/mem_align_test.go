package codesrc

import (
	"testing"
	"unsafe"
)

/*
# 什么是内存对齐
1. 计算机系统中用于优化内存访问速度的一组规则
2. 解决部分处理器无法访问任意地址上的任意数据
*/

/*
# Go内存对齐规则
 1. 对齐系数
	- 字段对齐
		结构体中的每个字段都按其对齐系数对齐。第一个字段从偏移量0开始，后续字段的起始位置是前一个字段结束位置后按字段对齐系数对齐的位置。
		- 注意：空结构体的对齐系数是0，但当其作为最后一个成员时，需要当做非0大小字段进行对齐。
	- 结构体对齐
		结构体的对齐系数是所有字段对齐系数的最大值。
 2. 填充字节
	- 当字段无法自然对齐时，编译器会在上一个字段或当前字段后面填充字节，以对齐。
	- 当结构体无法自然对齐时，编译器会在结构体后面填充字节，以对齐。
总结，字段排列顺序会影响结构体的实际占用空间。
*/

/*
# Go特定内存对齐规则 https://go.dev/ref/spec#Size_and_alignment_guarantees
1. 对于任意类型的变量 x ，unsafe.Alignof(x) 至少为 1。
2. 对于 struct 结构体类型的变量 x，计算 x 每一个字段 f 的 unsafe.Alignof(x.f)，unsafe.Alignof(x) 等于其中的最大值。
3. 对于 array 数组类型的变量 x，unsafe.Alignof(x) 等于构成数组的元素类型的对齐倍数。
4. 空结构体和空数组占用0字节空间。
*/

/*
# 空 struct{} 的对齐
空 struct{} 大小为 0，作为其他 struct 的字段时，一般不需要内存对齐。但是有一种情况除外：即当 struct{} 作为结构体最后一个字段时，
需要内存对齐。因为如果有指针指向该字段, 返回的地址将在结构体之外，如果此指针一直存活不释放对应的内存，
就会有内存泄露的问题（该内存不因结构体释放而释放）。
*/

func TestSimpleMemAlign(t *testing.T) {
	// 示例1：自然对齐（无填充）
	type v1 struct {
		a struct{} // 对齐系数是0
		b int64    // 对齐系数是8（8和0取较大值）
	}
	//struct{}是个特殊存在，因此实际对齐边界是0（作为首字段时）
	t.Logf("Sizeof(v1{}): %d, Offsetof(v1{}.b): %d", unsafe.Sizeof(v1{}), unsafe.Offsetof(v1{}.b)) // 8, 0

	// 示例2：结构体无法对齐（有填充）
	type v2 struct {
		a int64    // 对齐系数是8，占用8B
		b struct{} // 对齐系数是8（8和0取较大值），占用0B。
		// 但是，当struct{}作为最后一个成员时，需要当做非0大小字段进行对齐，按照go的规则，struct{}默认对齐系数1，则v2总大小=8+1=9B
		// 然而9不是8的倍数，不满足结构体对齐要求，向b后面填充7B完成对齐
	}
	t.Logf("Sizeof(v2{}): %d, Offsetof(v2{}.a): %d", unsafe.Sizeof(v2{}), unsafe.Offsetof(v2{}.b)) // 16, 8

	// 示例3：字段无法对齐（有填充）
	type v3 struct {
		a int8  // 对齐系数是1, 占用1B
		b int16 // 对齐系数是2（2和1取较大值），占用2B（当前偏移量是1，不是2倍数，无法对齐，向a后面填充1B完成对齐）
	}
	t.Logf("Sizeof(v3{}): %d, Offsetof(v3{}.b): %d", unsafe.Sizeof(v3{}), unsafe.Offsetof(v3{}.b)) // 4, 2

	/*
		小结：
			- 当类型较大的字段排列在结构体尾部时，一般需要填充前面的字段完成字段对齐
			- 当类型较小的字段排列在结构体尾部时，一般需要填充该字段完成结构体对齐
	*/
}

func TestMemAlign(t *testing.T) {
	// 示例1：字段对齐
	type v1 struct {
		a int8  // 对齐系数是1, 占用1B
		b int16 // 对齐系数是2（2和1取较大值），占用2B。当前b偏移量是1，不是2倍数，无法对齐，向a后面填充1B完成对齐
		c int64 // 对齐系数是8（8和0取较大值），占用8B。当前c偏移量是4，不是8倍数，无法对齐，向b后面填充4B完成对齐
	}
	t.Logf("Sizeof(v1{}): %d, Sizeof(v1{}.a): %d", unsafe.Sizeof(v1{}), unsafe.Sizeof(v1{}.b)) // 16, 1
}
