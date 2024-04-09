# Redis 相关

本文内容摘自网络。

## 为什么 Redis 使用单线程

- 简单高效：单线程模型使 Redis 的开发和维护更加简单，不需要处理多线程带来的加锁、线程同步等复杂情况。
- 瓶颈不在 CPU：Redis 作为内存数据库，性能瓶颈主要在内存和网络带宽而非 CPU。
- 数据结构简单：Redis 的数据结构被专门设计得很简单高效，绝大部分操作的时间复杂度都是 O(1)，因此单线程已经足以应对大部分读写场景。
- I/O 多路复用：利用了操作系统提供的多路 I/O 复用 epoll 模型，可以高效地监听和处理多个客户端连接。

## 单线程的瓶颈

- 只能用一个 cpu 核(忽略后台线程)
- 如果 value 比较大，redis 的 QPS 会下降得很厉害，有时一个大 key 就可以拖垮
- QPS 难以更上一层楼

## 为什么 Redis 后来引入多线程

### Redis 4.x 的多线程

Redis 在 4.x 版本引入了多线程，用来**异步**执行`UNLINK`、`FLUSHALL ASYNC`、`FLUSHDB ASYNC`命令。
比如对于键的删除，我们一般不需要同步等待完成，而且删除大键是一个耗时操作。所以引入多线程是方便执行那些不需要同步返回的命令。

> [!NOTE]
> `UNLINK`是 Redis 4.0 新增的命令，用于异步删除一个（较大的）键，返回值是实际解除链接的键的数量。
> `DEL`命令仍然是同步删除一个键。

### Redis 6.x 的多线程 I/O

Redis 官方在 2020 年 5 月正式推出 6.0 版本，此版本正式引入了多线程 I/O。

首先要解释 **Redis 的单线程**：Redis 在处理客户端的请求时，包括获取 (socket 读)、解析、执行、内容返回 (socket 写)
等都由一个顺序串行的主线程处理。

随着硬件性能提升，Redis 的单线程性能瓶颈可能出现在网络 IO 的读写，也就是：单个线程处理网络读写的速度跟不上底层网络硬件的速度。
读写网络的 read/write 系统调用占用了 Redis 执行期间大部分 CPU 时间，瓶颈主要在于网络的 IO 消耗。
此时的优化方向：

- 提高网络 IO 性能，典型的实现比如使用 DPDK 来替代内核网络栈的方式。
- 使用多线程充分利用多核，提高网络请求读写的并行度，典型的实现比如 Memcached。

Redis 采用了第二种方式，即 Redis 采用多个 IO 线程来处理网络请求，提高网络请求处理的并行度。
**需要注意的是**，Redis 多 IO 线程模型只用来处理网络读写请求，对于 Redis 的读写命令，依然是单线程处理。

**开启多线程**

Redis 6.0 的多线程默认是禁用的，只使用主线程。如需开启需要修改 redis.conf 配置文件：

```shell
io-threads-do-reads yes
io-threads 4 # 建议为CPU核数-1
```

## Redis 的多路复用与 HTTP/2 有何不同

**应用层别不同**

HTTP/2 的多路复用发生在应用层，即在一个 TCP 连接上复用多条流。而 Redis 的多路复用发生在更底层的网络 IO 层，即在一个线程中同时处理多个客户端
socket 连接的 IO 操作。

**目的不同**

HTTP/2 多路复用的主要目的是减少 TCP 连接数，提高带宽利用率。Redis 的多路复用主要目的是保持单线程以及不必要的上下文切换开销。
