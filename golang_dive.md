# Golang 深挖

## interface原理

## Map原理

### 实现

典型的哈希表 + 链地址法（但这里的链是 bucket 链）。

### hmap结构（简化版）

```
type hmap struct {
    count     int    // 当前元素数量
    flags     uint8
    B         uint8  // bucket 数量 = 2^B
    noverflow uint16 // overflow bucket 数量
    hash0     uint32 // 随机 hash 种子，作为计算 key 的哈希的第二参数

    buckets    unsafe.Pointer // 指向 buckets 数组，大小为 2^B，初始化nil
    // 等量扩容的时候，buckets 长度和 oldbuckets 相等
    // 双倍扩容的时候，buckets 长度会是 oldbuckets 的两倍
    oldbuckets unsafe.Pointer 
    nevacuate  uintptr        // 已迁移 bucket 数，指示扩容进度
}
```

关键点:

- B 代表桶数量的对数：桶数量为2的B次方
- hash0：防 hash 攻击（每个 map 不同）
- 扩容不是一次性，而是渐进式 rehash

**架构图**

<img src="./img/gomap.png" alt="" width="800" height="550">

### hmap内的buckets是切片结构吗

**不是**，hmap.buckets 不是 Go 的切片（`[]bucket`），而是一块连续分配的 bucket 数组的裸内存指针。
这是一个运行时私有、绕过 Go 类型系统的结构。

逻辑结构：

```
[bucket0][bucket1][bucket2]...[bucket(2^B - 1)]
```

它等价于一个长度为 2^B 的 bmap 数组。

**为何不使用切片？**

切片有长度和容量2个字段，占用16B，没必要。

### bucket 结构（重点）

```
type bmap struct {
    topbits  [8]uint8    // 存储每个槽位 key 的高8位哈希值，用于快速比较，避免直接比较完整的key
    keys     [8]keytype  // 存储8个key的数组，一个bucket最多存储8个键值对
    values   [8]valuetype // 存储8个value的数组，与keys按索引对应
    pad      uintptr     // 填充字段，用于内存对齐，确保结构体大小为2的幂次，优化内存访问
    overflow uintptr     // 指向溢出bucket的指针，当当前bucket存满时链接到下一个overflow bucket
}
```

- 一个 bucket = 8 个槽位
- 一个 bucket 最多存 8 个 kv
- 超过 8 个 → 挂 overflow bucket

这就是 Go 的“拉链法”。

**桶结构图**

<img src="./img/gomap_bmap.png" alt="" width="400" height="500">

### 初始化Map

底层调用的是 `makemap` 函数，主要做的工作就是初始化 hmap 结构体的各种字段，例如计算 B 的大小，设置哈希种子 hash0 等等。

### Write过程

1. **计算哈希值**: 首先计算 key 的哈希值 `hash = h(key, hash0)`，其中 hash0 是随机种子，用于防止哈希冲突攻击。

2. **定位 bucket**: 使用哈希值的低 B 位（即 `hash & (2^B - 1)`）确定目标 bucket 的索引位置。

3. **检查是否已存在**: 在目标 bucket 内顺序扫描 8 个槽位，比较 tophash（哈希值的高 8 位）和 key 本身，如果找到相同的
   key，则直接覆盖其对应的 value，操作结束。

4. **寻找空槽**:
    - 如果当前 bucket 有空位（槽位的 tophash 为 `empty` 或 ` evacuatedEmpty`），则将 key-value 存储到该空槽中：
        - 计算 key 的 tophash 并存储到 `topbits` 数组对应位置
        - 将 key 存储到 `keys` 数组对应位置
        - 将 value 存储到 `values` 数组对应位置
        - 更新 hmap 的 count 计数器

    - 如果当前 bucket 已满（8 个槽位都非空），则继续在 overflow bucket 链中寻找空槽：
        - 遍历 overflow bucket 链，寻找有空槽的 bucket
        - 找到后执行与当前 bucket 相同的存储操作

5. **创建 overflow bucket**: 如果当前 bucket 和其 overflow 链上都没有空槽，则创建一个新的 overflow bucket，并将新的
   key-value 存储到新 bucket 的第一个槽位。

6. **触发扩容检查**: 在完成插入操作后，会检查是否需要扩容：
    - 装载因子过高（元素数量 / bucket 数量 > 6.5）
    - 或者 overflow bucket 过多（即使元素数量不多，但分布极不均匀）

   如果满足扩容条件，则开始准备渐进式扩容过程（但不是立即完成，而是后续操作中逐步迁移元素）。

7. **更新 map 状态**: 更新 hmap 结构中的元素计数器 `count`。

整个 Write 过程会自动处理哈希冲突、bucket 溢出、以及适时的扩容操作，确保 map 能够高效运行。

### Read过程

1. **计算哈希值**: 首先计算 key 的哈希值 `hash = h(key, hash0)`，其中 hash0 是随机种子，用于防止哈希冲突攻击，确保每个 map
   实例的哈希值不同。

2. **定位 bucket**: 使用哈希值的低 B 位（即 `hash & (2^B - 1)`）确定目标 bucket 的索引位置，这确保了索引值在 [0, 2^B-1]
   范围内。

3. **检查扩容状态**: 如果 map 正在进行扩容（oldbuckets 非空），则需要先检查 key 是否可能存在于老的 bucket 中（因为扩容是渐进式的，
   部分数据可能还在老 bucket 中）。此时会同时在新旧 bucket 中查找。

4. **bucket内查找**: 在目标 bucket 中顺序扫描 8 个槽位：
    - 比较当前槽位的 tophash 与 key 哈希值的高 8 位是否匹配
    - 如果 tophash 匹配，则进一步比较 key 本身是否完全相同（防碰撞检查）
    - 如果 key 也匹配，则返回对应的 value，查找成功

5. **处理 tophash 不匹配**: 如果当前槽位的 tophash 不匹配，继续扫描下一个槽位，直到扫描完当前 bucket 的所有 8 个槽位。

6. **遍历 overflow bucket 链**: 如果当前 bucket 中没有找到匹配的 key，则沿着 overflow bucket 链继续查找：
    - 访问当前 bucket 的 overflow 指针指向的下一个 bucket
    - 重复步骤 4-5 的查找过程
    - 直到遍历完所有相关的 overflow bucket

7. **返回结果**: 如果在整个 bucket 链中都未找到匹配的 key，则返回零值，表示该 key 不存在于 map 中。

**优化机制**: tophash 机制（存储哈希值的高 8 位）起到了类似布隆过滤器的作用，先快速过滤掉不匹配的槽位，避免了直接比较完整的
key，提升了查找效率。

### Hash函数的选择

在程序启动时，会检测 cpu 是否支持 aes，如果支持，则使用 aes hash，否则使用 memhash。这是在函数 alginit() 中完成，位于路径：
`src/runtime/alg.go` 下。

> hash 函数，有加密型和非加密型。 加密型的一般用于加密数据、数字摘要等，典型代表就是 md5、sha1、sha256、aes256 这种；
> 非加密型的一般就是查找。在 map 的应用场景中，用的是查找。 选择 hash 函数主要考察的是两点：性能、碰撞概率。

### 扩容过程

#### 触发因素

1. **装载因子过高**: 当元素数量 count 与 bucket 数量 2^B 的比值超过 6.5 时触发扩容（count / (2^B) >
   6.5）。这是最常见的扩容条件，表示哈希表过于拥挤，查找效率下降。

2. **溢出桶过多**: 即使元素数量不多，但若存在大量 overflow bucket，也会触发扩容。这种情况表明哈希函数分布不均或存在大量哈希冲突，导致链式结构过长，影响性能。

#### 扩容方式

- **双倍扩容**: 正常情况下，bucket 数量从 2^B 增加到 2^(B+1)，即 buckets 指针指向的新数组大小是原数组的两倍，B 的值加 1。
- **等量扩容**: 在某些特殊情况下（如溢出桶过多但元素数量不多），会进行等量扩容，即 B 值不变，但重新分配 bucket，以减少溢出桶数量。
    - 这种情况的扩容，会自动将overflow桶链内的所有元素”压缩“到新的 bucket 中，从而释放那些桶内的Empty槽位空间。

#### 渐进式 rehash 过程

Go map 采用渐进式 rehash 机制，避免一次性迁移所有数据导致的性能问题：

1. **准备阶段**:
    - 将当前 buckets 指针赋值给 oldbuckets（即 `oldbuckets = buckets`，初始值nil）
    - 分配新的 buckets 数组，大小为原数组的两倍（双倍扩容）或与原数组相同（等量扩容）
    - 初始化迁移进度指示器 nevacuate = 0

2. **迁移阶段**:
    - 每次对 map 进行读写操作时，除了完成正常的读写逻辑，还会顺带迁移一部分 bucket
    - 选择需要迁移的 bucket（从索引 0 开始，直至 nevacuate 指示的位置）
    - 将 oldbucket 中的数据按新的哈希值重新分配到新 buckets 的相应位置
    - 更新 nevacuate 计数器，记录已迁移的 bucket 数量

3. **完成阶段**:
    - 当所有 oldbucket 中的数据都迁移完毕（nevacuate == 2^B），表示扩容完成
    - 清理 oldbuckets 指针，释放旧的 bucket 数组内存

#### 迁移细节

- 在双倍扩容时，一个 oldbucket 中的数据会被迁移到新的两个 bucket 中（因为新哈希值的低位会决定具体位置）
- 在迁移过程中，如果需要查找某个 key，会先在新 buckets 中查找，如果没找到再在对应的 oldbuckets 中查找
- 写操作时，如果目标 key 所在的 bucket 尚未迁移，则会先执行迁移操作再进行写入

这种渐进式 rehash 策略确保了 map 在扩容过程中的可用性，避免了因一次性 rehash 大量数据而导致的长时间阻塞。

### Map并发不安全的本质原因

Go map 并发不安全的根本原因在于其内部操作不是原子性的，主要体现在以下几个方面：

1. **扩容操作的非原子性**:
    - 当多个 goroutine 同时对 map 进行写操作时，可能同时触发扩容条件
    - 扩容过程涉及内存重新分配、数据迁移等复杂操作，这些操作不是原子的
    - 如果一个 goroutine 在扩容过程中，另一个 goroutine 仍在访问旧的 bucket 结构，可能导致访问已释放的内存或数据不一致

2. **bucket 链修改的非原子性**:
    - 插入新元素可能导致创建新的 overflow bucket 并修改链表结构
    - 删除元素可能修改 bucket 的 overflow 指针
    - 这些链表结构的修改不是原子操作，多个 goroutine 同时修改可能造成链表损坏

3. **渐进式 rehash 的并发问题**:
    - 在扩容过程中，map 的状态处于新旧数据结构并存的状态
    - 一个 goroutine 可能在新 buckets 中查找，而另一个 goroutine 仍在旧 buckets 中修改
    - 这种不一致的视图可能导致数据丢失或程序崩溃

4. **内存访问竞争**:
    - 多个 goroutine 同时访问和修改同一块内存区域（如 bucket 内的 key-value 对）
    - 可能导致数据损坏或程序 panic

Go runtime 为了性能考虑，没有在 map 内部实现并发安全机制，因此在并发环境下需要使用额外的同步机制（如 sync.Mutex 或
sync.RWMutex）或使用 sync.Map 等并发安全的数据结构。

### Map的key是否有序，为什么

Go map 中的 key **是无序的**，遍历时无法保证任何特定顺序。具体原因如下：

1. **哈希函数的无序性**:
    - map 使用哈希函数将 key 映射到特定的 bucket 中
    - 哈希函数的目的是均匀分布数据，而不是保持顺序
    - 相邻的 key（如 "a", "b", "c"）经过哈希后可能分散到不同的 bucket 中

2. **bucket 和 overflow 链结构**:
    - 数据按照哈希值分布在多个 bucket 中
    - 每个 bucket 内部存储 8 个 key-value 对
    - 当 bucket 满了时，数据会存储在 overflow bucket 中
    - 遍历时按 bucket 顺序访问，但同一个逻辑上的连续序列在物理上可能分布在不同的 bucket 和 overflow 链中

3. **Go runtime 的刻意随机化**:
    - Go 语言在设计上刻意随机化了 map 的遍历起始点
    - 每次遍历 map 时，Go runtime 会从一个随机的 bucket 开始遍历
    - 这是为了防止程序员错误地依赖 map 的遍历顺序，因为这种依赖可能导致程序在不同运行环境或 Go 版本下表现不一致

4. **扩容对顺序的影响**:
    - map 在扩容时会重新分配数据到新的 bucket 结构中
    - 即使在一次运行中，扩容操作也可能改变数据的物理存储位置，进一步影响遍历顺序

### 删除过程

Go map 的删除操作通过 `delete(map, key)` 函数实现，主要包括以下步骤：

1. **计算哈希值**：计算待删除 key 的哈希值，确定目标 bucket
2. **定位元素**：在目标 bucket 及其 overflow 链中查找目标 key
3. **执行删除**：找到后将对应槽位的 tophash 标记为 `empty`，清理 key 和 value，并递减元素计数器
4. **内存管理**：删除仅标记槽位为空，并不立即回收内存

删除操作的时间复杂度平均为 O(1)。

### 遍历过程

Go map 的遍历操作使用 `for k, v := range map` 语法实现：

1. **随机起始点**：为防止程序依赖遍历顺序，Go 会从随机的 bucket 开始遍历
2. **遍历策略**：按顺序遍历所有 buckets 及其 overflow 链，访问每个槽位的元素
3. **扩容处理**：如果 map 正在扩容，会同时考虑新旧 buckets 中的数据
4. **安全检查**：遍历过程中检测到并发写入会触发 panic
5. **性能特点**：时间复杂度为 O(n)

**重要提示**：Go map 的遍历顺序是不固定的，不应依赖遍历顺序实现业务逻辑。

### Map的Key有何类型要求

Go 语言中 map 的 key 需要满足特定的要求，主要是因为需要对 key 进行比较操作以实现查找、插入和删除功能：

#### 1. 可比较类型（Comparable Types）

map 的 key 必须是**可比较类型**，即可以使用 `==` 和 `!=` 操作符进行比较的类型。Go 语言中可比较的类型包括：

- **基本类型**：
    - 布尔类型（bool）
    - 数值类型（int, int8, int16, int32, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr）
    - 浮点类型（float32, float64）
    - 复数类型（complex64, complex128）
    - 字符串类型（string）

- **复合类型**：
    - 指针类型（*T）
    - 通道类型（chan T）
    - 接口类型（interface{}）
    - 数组类型（[N]T）- 当且仅当元素类型 T 是可比较的
    - 结构体（struct）- 当且仅当其所有字段都是可比较的

#### 2. 不可作为 key 的类型

以下类型不能作为 map 的 key：

- **切片（slice）**：切片类型不可比较
- **映射（map）**：map 类型不可比较
- **函数（function）**：函数类型不可比较
- **包含不可比较字段的结构体**：如果结构体中包含切片、map 或函数等不可比较的字段，则该结构体不可比较

#### 3. 比较规则

- **数值类型**：按数值大小比较
- **字符串**：按字典序比较
- **指针**：比较指针指向的地址
- **通道**：相同通道值相等，不同通道值不等
- **接口**：如果两个接口的动态类型相同且动态值相等，则接口值相等

## Channel 介绍

### 目标

Channel 设计出来的目标就是践行 “通过通信的方式共享内存” 的思想。这有一个前提，那就是编程语言实现了用户态可调度的协程，
因为这样才能避免线程切换的上下文开销。Go 语言的 GPM 调度机制实现了可调度协程，这样即使 Channel 内部有锁，也是一种低开销的任务切换模型。

Channel 内部虽然使用了一把互斥锁，但该锁是一把基于 CAS 和自旋算法的乐观锁，能最大程度降低 goroutine 的切换次数。

### 优势

这里的优势是相比传统锁机制而言，Channel 简化了并发编程，同时使用了乐观锁来提高性能。同时 Channel 提供了一个队列结构，支持先进先出。

### 基本特性

Channel 支持带缓冲和不带缓冲的读写模式。不管哪种模式，当消费者来不及消费数据时，生产者都无法继续发送数据（阻塞），
直到消费者消费完 Channel 内的数据或 Channel 内的数据量小于缓冲区大小。

### 实现原理

#### 数据结构

Channel 内部是一个**有锁循环队列**实现，`hchan`是 Channel 运行时使用的数据结构。

```go
package runtime

type hchan struct {
	qcount   uint           // 队列中元素个数
	dataqsiz uint           // 队列长度（固定）
	buf      unsafe.Pointer // 缓冲区数据指针
	elemsize uint16         // 收发的元素size
	closed   uint32         // chan是否关闭
	elemtype *_type         // 收发的元素类型
	sendx    uint           // 发送操作处理到的位置
	recvx    uint           // 接收操作处理到的位置
	recvq    waitq          // 存储了等待的消费者goroutine列表
	sendq    waitq          // 存储了等待的发送者goroutine列表

	lock mutex // 保护 hchan 中的所有字段以及在该channel上阻塞的 recvq/sendq 中的sudogs中的几个字段
}
```

其中的`recvq`和`sendq`都是`waitq`类型，该类型是一个**双向链表**，用来存储等待的协程。`waitq`结构如下：

```go
package runtime

// 链表中所有的元素都是 runtime.sudog 结构
type waitq struct {
	first *sudog
	last  *sudog
}
```

链表中所有的元素都是 [`runtime.sudog`][sudog] 结构，`runtime.sudog` 表示一个在等待列表中的 Goroutine。

[sudog]: https://github.com/golang/go/blob/41d8e61a6b9d8f9db912626eb2bbc535e929fefc/src/runtime/runtime2.go#L345

#### 创建管道

Channel 只能通过`make`函数创建，创建时需要指定类型，大小可选，默认为 0 表示无缓冲区。

编译阶段，会根据缓冲大小来决定使用 [`runtime.makechan`][runtime.makechan] 或者 [`runtime.makechan64`][runtime.makechan64]
函数来创建 Channel，当缓冲区大小大于 2 的 32 次方时使用后者，很少见，这里只需要关注前者即可。

[runtime.makechan]: https://github.com/golang/go/blob/41d8e61a6b9d8f9db912626eb2bbc535e929fefc/src/runtime/chan.go#L71

[runtime.makechan64]: https://github.com/golang/go/blob/41d8e61a6b9d8f9db912626eb2bbc535e929fefc/src/runtime/chan.go#L63

`makechan`函数逻辑如下：

```go
package runtime

import (
	"runtime/internal/math"
	"unsafe"
)

func makechan(t *chantype, size int) *hchan {
	elem := t.elem
	mem, _ := math.MulUintptr(elem.size, uintptr(size)) // 元素大小 * 容量

	var c *hchan
	switch {
	case mem == 0: // 容量为0，即无缓冲区时，则仅创建chan，不创建缓冲区buf
		c = (*hchan)(mallocgc(hchanSize, nil, true))
		c.buf = c.raceaddr()
	case elem.kind&kindNoPointers != 0: // 元素非指针类型，为 Channel 和底层的数组分配一块连续的内存空间
		c = (*hchan)(mallocgc(hchanSize+mem, nil, true))
		c.buf = add(unsafe.Pointer(c), hchanSize)
	default: // 默认情况，先创建chan，然后再单独分配buf
		c = new(hchan)
		c.buf = mallocgc(mem, elem, true)
	}
	c.elemsize = uint16(elem.size)
	c.elemtype = elem
	c.dataqsiz = uint(size)
	return c
}

```

#### 发送数据

使用 `ch <- i` 语句发送数据到 Channel，运行时的函数调用逻辑是`runtime.chansend1` -> `runtime.chansend`，后者包含了全部发送逻辑。

```go
package runtime

import "unsafe"

func chansend(c *hchan, ep unsafe.Pointer, block bool, callerpc uintptr) bool {
	// ...

	// 先为当前 Channel 加锁, 防止并发修改channel
	lock(&c.lock)

	if c.closed != 0 {
		unlock(&c.lock)
		panic(plainError("send on closed channel"))
	}
	// ... 后面还有很多逻辑
}

```

因为 `runtime.chansend` 函数的实现比较复杂，所以我们这里将该函数的执行过程分成以下的三个部分：

- 当存在等待的接收者时，通过 runtime.send 直接将数据发送给阻塞的接收者；
- 当缓冲区存在空余空间时，将发送的数据写入 Channel 的缓冲区；
- 当不存在缓冲区或者缓冲区已满时，等待其他 Goroutine 从 Channel 接收数据；

##### 1. 直接发送

如果目标 Channel 没有被关闭并且已经有处于读等待的 Goroutine，那么 `runtime.chansend` 会从接收队列 `recvq` 中取出最先陷入等待的
Goroutine 并直接向它发送数据：

```plain
if sg := c.recvq.dequeue(); sg != nil {
  send(c, sg, ep, func() { unlock(&c.lock) }, 3)
  return true
}
```

发送数据时调用的是 [`runtime.send`][runtime.send] 函数，该函数的执行可以分成两个部分：

- 调用 `runtime.sendDirect` 将发送的数据直接拷贝到 x = <-c 表达式中变量 x 所在的内存地址上；
- 调用 `runtime.goready` 将等待接收数据的 Goroutine 标记成可运行状态 `Grunnable` 并把该 Goroutine 放到发送方所在的处理器的
  `runnext` 上等待执行，该处理器在下一次调度时会立刻唤醒数据的接收方；

[runtime.send]: https://github.com/golang/go/blob/41d8e61a6b9d8f9db912626eb2bbc535e929fefc/src/runtime/chan.go#L292

##### 2. 缓冲区有空余

如果创建的 Channel 包含缓冲区并且 Channel 中的数据没有装满，会执行下面这段代码：

```plain
func chansend(c *hchan, ep unsafe.Pointer, block bool, callerpc uintptr) bool {
	...
	if c.qcount < c.dataqsiz {
		qp := chanbuf(c, c.sendx)
		typedmemmove(c.elemtype, qp, ep)
		c.sendx++
		if c.sendx == c.dataqsiz {
			c.sendx = 0
		}
		c.qcount++
		unlock(&c.lock)
		return true
	}
	...
}
```

在这里我们首先会使用 `runtime.chanbuf` 计算出下一个可以存储数据的位置，然后通过 `runtime.typedmemmove` 将发送的数据**拷贝到
**缓冲区中，然后增加
`sendx` 索引和 `qcount` 计数器。这里本质上是将发送的数据拷贝到缓冲区中，而不是直接发送到接收方。

> [!NOTE]
> 注意 channel 的 buf 是一个数组实现的循环队列，所以当 `sendx` 等于 `dataqsiz` 时会重新回到数组开始的位置。

##### 3. 阻塞发送

当 Channel 没有接收者能够处理数据时，向 Channel 发送数据会被下游阻塞。当然，我们可以使用 select 非阻塞地往 Channel 发送数据。
阻塞发送数据会执行下面的代码，简单梳理一下这段代码的逻辑：

```plain
func chansend(c *hchan, ep unsafe.Pointer, block bool, callerpc uintptr) bool {
	...
	if !block {
		unlock(&c.lock)
		return false
	}

	gp := getg()
	mysg := acquireSudog()
	mysg.elem = ep
	mysg.g = gp
	mysg.c = c
	gp.waiting = mysg
	c.sendq.enqueue(mysg)
	goparkunlock(&c.lock, waitReasonChanSend, traceEvGoBlockSend, 3)

	gp.waiting = nil
	gp.param = nil
	mysg.c = nil
	releaseSudog(mysg)
	return true
}
```

1. 调用 `runtime.getg` 获取发送数据使用的 Goroutine；
2. 执行 `runtime.acquireSudog` 获取 `runtime.sudog` 结构并设置这一次阻塞发送的相关信息，例如发送的 Channel、是否在 select
   中和待发送数据的内存地址等；
3. 将刚刚创建并初始化的 `runtime.sudog` 加入发送等待队列，并设置到当前 Goroutine 的 `waiting` 上，表示 Goroutine 正在等待该
   `sudog` 准备就绪；
4. 调用 `runtime.goparkunlock` 将当前的 Goroutine 挂起等待唤醒（让出 P 和 M）；
5. 被调度器唤醒后会执行一些收尾工作，将一些属性置零并且释放 `runtime.sudog` 结构体；

函数在最后会返回 true 表示这次我们已经成功向 Channel 发送了数据。

##### 4. 小结

简单梳理和总结一下使用 `ch <- i` 表达式向 Channel 发送数据时遇到的几种情况：

- 如果当前 Channel 的 `recvq` 上存在已经被阻塞的 Goroutine，那么会直接将数据发送给当前 Goroutine 并将其设置成下一个运行的
  Goroutine；
- 如果没有等待接收者，且 Channel 存在缓冲区并且其中还有空闲的容量，我们会直接将数据存储到缓冲区 sendx 所在的位置上；
- 如果不满足上面的两种情况，会创建一个 `runtime.sudog` 结构并将其加入 Channel 的 `sendq` 队列中，当前 Goroutine
  也会陷入阻塞等待其他的协程从 Channel 接收数据；

Goroutine 挂起时触发 Goroutine 调度，关键函数是`runtime.goparkunlock`。

#### 接收数据

我们可以通过 2 种方式接收数据：

```plain
i <- ch
i, ok <- ch
```

分别对应 `runtime.chanrecv1` 和 `runtime.chanrecv2`
两种不同函数的调用，但最终都会调用 [`runtime.chanrecv`][runtime.chanrecv] 函数。

[runtime.chanrecv]: https://github.com/golang/go/blob/41d8e61a6b9d8f9db912626eb2bbc535e929fefc/src/runtime/chan.go#L454

接收数据分为 2 种特殊情况和 3 种常规情况，2 种特殊情况是：

1. 从一个空 Channel 接收数据时会直接调用 `runtime.gopark` 让出处理器的使用权
2. 如果当前 Channel 已经被关闭并且缓冲区中不存在任何数据，那么会清除 ep 指针中的数据并立刻返回。

第一种情况对应的代码如下：

```plain
var cc chan int
<- cc
```

其余 3 种情况如下：

1. 当存在等待的发送者时，通过 `runtime.recv` 从阻塞的发送者或者缓冲区中获取数据；
2. 当缓冲区存在数据时，从 Channel 的缓冲区中接收数据；
3. 当缓冲区中不存在数据时，等待其他 Goroutine 向 Channel 发送数据；

##### 直接接收

当 Channel 的 `sendq` 队列中包含处于等待状态的 Goroutine 时，该函数会取出队列头等待的 Goroutine，处理的逻辑和发送时相差无几，
只是发送数据时调用的是 `runtime.send` 函数，而接收数据时使用 `runtime.recv`。

```plain
if sg := c.sendq.dequeue(); sg != nil {
		recv(c, sg, ep, func() { unlock(&c.lock) }, 3)
		return true, true
	}
```

`runtime.recv`实现如下：

```plain
func recv(c *hchan, sg *sudog, ep unsafe.Pointer, unlockf func(), skip int) {
    // 如果channel没有缓冲区，调用 runtime.recvDirect 将 Channel 发送队列中 Goroutine 存储的 elem 数据拷贝到目标内存地址中；
	if c.dataqsiz == 0 {
		if ep != nil {
			recvDirect(c.elemtype, sg, ep)
		}
	} else {
	    // 如果 Channel 存在缓冲区
	    //  1. 将队列中的数据拷贝到接收方的内存地址；
	    //  2. 将发送队列头的数据拷贝到缓冲区中，释放一个阻塞的发送方；
		qp := chanbuf(c, c.recvx)
		if ep != nil {
			typedmemmove(c.elemtype, ep, qp)
		}
		typedmemmove(c.elemtype, qp, sg.elem)
		c.recvx++
		c.sendx = c.recvx // c.sendx = (c.sendx+1) % c.dataqsiz
	}
	gp := sg.g
	gp.param = unsafe.Pointer(sg)
	goready(gp, skip+1)
}
```

无论发生哪种情况，运行时都会调用 `runtime.goready` 将当前处理器的 runnext 设置成发送数据的 Goroutine，在调度器下一次调度时将阻塞的发送方唤醒。


> [!NOTE]
> 你可能已经发现了，【直接接收】也包含了【缓冲区接收】的逻辑，但前提条件是存在阻塞的发送者。

##### 从缓冲区接收

当 Channel 的缓冲区中已经包含数据时，从 Channel 中接收数据会直接从缓冲区中 `recvx` 的索引位置中取出数据进行处理：

```plain
func chanrecv(c *hchan, ep unsafe.Pointer, block bool) (selected, received bool) {
	...
	if c.qcount > 0 {
	    // 将队列中的元素拷贝到接收变量中
		qp := chanbuf(c, c.recvx)
		if ep != nil {
			typedmemmove(c.elemtype, ep, qp)
		}
		// 重置队列元素
		typedmemclr(c.elemtype, qp)
		// 接收下标偏移
		c.recvx++
		// 偏移量达到channel容量时，重置下标为0（循环）
		if c.recvx == c.dataqsiz {
			c.recvx = 0
		}
		// 队列长度减1
		c.qcount--
		return true, true
	}
	...
}
```

##### 阻塞接收

当 Channel 的发送队列中不存在等待的 Goroutine 并且缓冲区中也不存在任何数据时，从管道中接收数据的操作会变成阻塞的。

```plain
func chanrecv(c *hchan, ep unsafe.Pointer, block bool) (selected, received bool) {
	...
	if !block {
		unlock(&c.lock)
		return false, false
	}

	gp := getg()
	mysg := acquireSudog()
	mysg.elem = ep
	gp.waiting = mysg
	mysg.g = gp
	mysg.c = c
	c.recvq.enqueue(mysg)
	goparkunlock(&c.lock, waitReasonChanReceive, traceEvGoBlockRecv, 3)

	gp.waiting = nil
	closed := gp.param == nil
	gp.param = nil
	releaseSudog(mysg)
	return true, !closed
}
```

在正常的接收场景中，我们会使用 runtime.sudog 将当前 Goroutine 包装成一个处于等待状态的 Goroutine 并将其加入到接收队列中。
入队后，会调用 `runtime.goparkunlock` 将当前 Goroutine 挂起，并等待其他 Goroutine 向 Channel 发送数据。

> [!NOTE]
> 不是所有的接收操作都是阻塞的，与 select 语句结合使用时就可能会使用到非阻塞的接收操作。

#### 关闭管道

对应调用的是 [`runtime.closechan`][runtime.closechan] 函数。当 Channel 是一个空指针或者已经被关闭时，运行时会直接崩溃并抛出异常：

```plain
func closechan(c *hchan) {
	if c == nil {
		panic(plainError("close of nil channel"))
	}

	lock(&c.lock)
	if c.closed != 0 {
		unlock(&c.lock)
		panic(plainError("close of closed channel"))
	}
	...
}
```

然后开始执行关闭 channel 的逻辑：

```plain
func closechan(c *hchan) {
    ...
	c.closed = 1

	var glist gList
	for {
		sg := c.recvq.dequeue()
		if sg == nil {
			break
		}
		if sg.elem != nil {
			typedmemclr(c.elemtype, sg.elem)
			sg.elem = nil
		}
		gp := sg.g
		gp.param = nil
		glist.push(gp)
	}

	for {
		sg := c.sendq.dequeue()
		...
	}
	for !glist.empty() {
		gp := glist.pop()
		gp.schedlink = 0
		goready(gp, 3)
	}
}
```

大致步骤：

- 将 `recvq` 和 `sendq` 两个队列中的 goroutine 加入 `gList` 临时链表
- 然后遍历链表，调用`goready`函数将 goroutine 唤醒。

> [!NOTE]
> `goready`函数实际只是触发了 goroutine 的调度，并没有立即执行该 goroutine。

[runtime.closechan]: https://github.com/golang/go/blob/41d8e61a6b9d8f9db912626eb2bbc535e929fefc/src/runtime/chan.go#L355

#### gopark 函数

前面我们提到的`gopark`函数，其作用是让出当前协程的控制权，并阻塞当前协程，等待其他协程唤醒。下面通过一个情景来说明。

```plain
goparkunlock(&c.lock, "chan send", traceEvGoBlockSend, 3)
```

发送数据到 channel 时，当队列已满时，会调用 `goparkunlock` 函数将当前协程挂起，并等待接收者读取数据。该函数的内部逻辑如下：

```plain
func goparkunlock(lock *mutex, reason string, traceEv byte, traceskip int) {
	gopark(parkunlock_c, unsafe.Pointer(lock), reason, traceEv, traceskip)
}
```

最终会调用`gopark`函数，其内部逻辑如下：

```plain
func gopark(unlockf func(*g, unsafe.Pointer) bool, lock unsafe.Pointer, reason string, traceEv byte, traceskip int) {
	mp := acquirem()
	gp := mp.curg
	status := readgstatus(gp)
	if status != _Grunning && status != _Gscanrunning {
		throw("gopark: bad g status")
	}
	mp.waitlock = lock
	mp.waitunlockf = *(*unsafe.Pointer)(unsafe.Pointer(&unlockf))
	gp.waitreason = reason
	mp.waittraceev = traceEv
	mp.waittraceskip = traceskip
	releasem(mp)
	// can't do anything that might move the G between Ms here.
	mcall(park_m)
}
```

`gopark`函数的大致逻辑是获取当前协程 G 和线程 M，在判断 G 的状态正常后，开始设置 M 的相关属性，然后调用`releasem(mp)`
使得当前 G 允许被强占。

然后调用`mcall(park_m)`，内部逻辑是将当前 G 挂起，并取消当前 G 和 M 的关联关系，去寻找并运行下一个可运行的 goroutine。

```plain
func park_m(gp *g) {
	_g_ := getg()

	if trace.enabled {
		traceGoPark(_g_.m.waittraceev, _g_.m.waittraceskip)
	}

	casgstatus(gp, _Grunning, _Gwaiting)
	dropg()

	if _g_.m.waitunlockf != nil {
		fn := *(*func(*g, unsafe.Pointer) bool)(unsafe.Pointer(&_g_.m.waitunlockf))
		ok := fn(gp, _g_.m.waitlock)
		_g_.m.waitunlockf = nil
		_g_.m.waitlock = nil
		if !ok {
			if trace.enabled {
				traceGoUnpark(gp, 2)
			}
			casgstatus(gp, _Gwaiting, _Grunnable)
			execute(gp, true) // Schedule it back, never returns.
		}
	}
	schedule()
}
```

注意其中有一个 `if _g_.m.waitunlockf != nil` 的分支，即要处理 M 上处理可能存在的解锁函数，如果当前 M
有一个等待解锁函数（`waitunlockf`），则调用该函数尝试解锁。若解锁失败，需要恢复当前 G 的状态，并继续执行当前 G。

### 总结

Channel 内部使用了以下技术点来完成：

- 缓冲区使用循环队列
- 发送&接收队列使用双向链表
- 需要阻塞时使用`gopark`函数挂起协程，调度其他可运行的协程
- 发送数据和接收数据都需要 Channel 级别的**互斥锁**保护
    - 锁是一个乐观锁实现，内部先使用 CAS 操作来尝试获取锁，如果获取失败则使用自旋等待
    - 若自旋几次仍然获取失败，则通过操作系统提供的异步事件通知系统调用来阻塞 G，等待其他协程（释放锁时）调用系统调用通知，避免空耗
      CPU
    - Go 将这些系统调用封装在了几个跨平台的函数中：semacreate/semasleep/semawakeup
    - 在 windows 上的异步事件通知系统调用接口是:
        - _CreateEventA: 创建事件，对应 semacreate
        - _WaitForSingleObject: 等待事件信号，对应 semasleep
        - _SetEvent：设置事件信号，对应 semawakeup

### 问题 1：chan 内的锁是如何实现的？

首先是**加锁**，chan 内不管读写都是通过`lock(&c.lock)`来完成添加和释放的，[代码链接][lock_sema]，内部实现如下：

```plain
func lock2(l *mutex) {
	gp := getg()
	if gp.m.locks < 0 {
		throw("runtime·lock: lock count")
	}
	gp.m.locks++

	// 投机性的抢一次（可能一次就抢到）
	if atomic.Casuintptr(&l.key, 0, locked) {
		return
	}
	// 先在M（OS线程）上创建2个信号量（waitsema, resumesema）
	semacreate(gp.m)
    
	// 下面是一段典型的自旋锁实现
	
	// 1. 单核处理器上不会自旋，直接进入等待
	// 2. 多核处理器上，自旋次数为 active_spin （4次）
	spin := 0
	if ncpu > 1 {
		spin = active_spin
	}
Loop:
	for i := 0; ; i++ {
		v := atomic.Loaduintptr(&l.key) // v是等待锁的M链表头指针
		if v&locked == 0 { // 没有任何M占用锁（反之表示至少有一个M占用锁）
			// 乐观性地获得锁
			if atomic.Casuintptr(&l.key, v, v|locked) { // 成功直接return
				return
			}
			i = 0 // 否则进入自旋
		}
		if i < spin {
			// 在一定次数内（spin=30次），通过忙等待尝试获取锁
			// 这是一个轻量级的忙等待函数，它不会导致线程进入睡眠状态，而是通过执行一些空操作（如 PAUSE 指令）或短时间的忙等待，让出 CPU 给其他线程。
			// 适用于锁竞争激烈但锁持有时间非常短的场景。
			procyield(active_spin_cnt) 
		} else if i < spin+passive_spin {
			// 在更多的次数内（spin + passive_spin 次）
			// 这是一种比procyield更重量级的让出 CPU 的方式，它会主动触发操作系统的线程调度（产生系统线程的上下文切换）。
			osyield()
		} else {
			// 经过上面多次自旋后，锁仍然被其他G占用，则当前M进入等待队列且睡眠等待唤醒
			for {
			    // 1.将处于锁等待队列的M头指针复制给nextwaitm（nextwaitm是一个队列链表指针，v去掉标志位后恰好是等待同一把锁的M链表头指针）
			    // （注意这里对locked的处理仅仅是去掉该标志位，为了拿到M的指针）
				gp.m.nextwaitm = muintptr(v &^ locked) 
				// 2.再将当前M指针加上标志位，并设置到l.key
				// （这两步操作共同实现了：将当前M加入等待队列）
				if atomic.Casuintptr(&l.key, v, uintptr(unsafe.Pointer(gp.m))|locked) {
					break // 入列成功则跳出循环，进入步骤4
				}
				// 3.锁状态发生变化（有M抢占或释放锁，就会更新l.key），则重新获取锁状态
				v = atomic.Loaduintptr(&l.key)
				if v&locked == 0 { // 又是一个新的chan操作，继续自旋
					continue Loop
				}
			}
			// 4. M入列后，再检查一次锁是否被占用，是则当前M进入睡眠状态，等待被唤醒（减少CPU消耗）
			// （在unlock2函数内有唤醒的逻辑）
			if v&locked != 0 {
				semasleep(-1)
				i = 0
			}
		}
	}
}
```

这段代码实现了一个混合自旋锁：

- 先通过自旋尝试获取锁，减少上下文切换开销。

- 如果自旋加锁失败，将线程加入等待队列并进入睡眠，避免浪费 CPU 资源。

- 通过原子操作（`Loaduintptr` 和 `Casuintptr`）确保锁状态的线程安全。

- 通过 `procyield` 和 `osyield` 实现不同强度的自旋等待策略。

然后是**释放锁**，[代码链接][lock_sema-unlock2]，内部实现如下：

```plain
// chan的每个操作都是先获得锁，然后再释放锁
func unlock2(l *mutex) {
	gp := getg()
	var mp *m
	
	// 这个循环的主要目的就是释放当前M占用的锁，如果M进入了队列，则需要从队列中弹出并唤醒；若没有进入队列，则直接释放锁
	for {
		v := atomic.Loaduintptr(&l.key) // 先获取锁状态
		if v == locked {  // 如果没有进入队列，锁恰好被当前M占用（根据lock2的逻辑可知，v==locked时表示仅有1个M占用锁），则直接尝试释放锁
			if atomic.Casuintptr(&l.key, locked, 0) {
				break // 释放成功，退出循环
			}
		} else {
			// 否则就是进入了队列，需要从队列中弹出并唤醒
			mp = muintptr(v &^ locked).ptr() // 先得到头M指针（l.key去掉lock标志位后就是M指针）
			if atomic.Casuintptr(&l.key, v, uintptr(mp.nextwaitm)) { // l.key切换到下一个等待锁的M指针（出列操作）
				semawakeup(mp) // 唤醒当前M，退出循环
				break
			}
		}
	}
	gp.m.locks--
	if gp.m.locks < 0 {
		throw("runtime·unlock: lock count")
	}
	if gp.m.locks == 0 && gp.preempt { // restore the preemption request in case we've cleared it in newstack
		gp.stackguard0 = stackPreempt
	}
}
```

### 参考

- [draveness-Channel 实现原理](https://draveness.me/golang/docs/part3-runtime/ch06-concurrency/golang-channel/)