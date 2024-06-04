# 验证行锁的三种算法

行锁的三种算法：记录锁、间隙锁和临键锁。

- 记录锁：只锁住索引记录。
- 间隙锁：锁住索引记录之间的间隙，是左开右开区间，如`(1, 3)`。
- 临键锁：锁住索引记录和索引记录之间的间隙，是左开右闭区间，如`(1, 3]`。

> [!NOTE]
> 这里提到的索引记录，要么是普通索引，要么是主键索引，根据索引使用情况来辨别。

## 加锁规则

Innodb默认的加锁算法是Next-Key Lock，即临键锁。但在一些场景下会退化为粒度更小的间隙锁或者记录锁。总的来说，遵循以下规则：

- 原则1：加锁的默认算法是Next-Key Lock；
- 原则2：查找过程中访问到的对象会加锁；
- 优化1：唯一索引上的等值查询，给命中的记录加锁时，退化为记录锁；
- 优化2：索引上的等值查询，未命中记录时，退化为间隙锁；
- 一个bug：唯一索引上的范围查询会访问到不满足条件的第一个值为止，即会多锁定一个间隙；

为了确认以上规则，下文会尽可能列出多的加锁案例，并运用以上规则进行解释，若读者发现不合理的地方，请留言。

> [!NOTE]
> 以上规则并不来自MySQL官方，而是来自极客时间[MySQL 实战 45 讲-第21章](https://time.geekbang.org/column/intro/100020801)。
> 笔者结合了网络上的其他资料，只有这个链接的作者才对以上规则有详细的解释，并结合了大量案例进行说明。
> 同时，本文也包含了笔者的更多实验结果。

下面，请同笔者一起进入Innodb加锁的世界！

## 关于加锁

行锁的加锁方式有两种：一种是SELECT加`FOR UPDATE` 或 `LOCK IN SHARE MODE`，不加修饰则是快照读，无任何锁；
另一种是插入、更新、删除等操作自动隐式加锁。

## 准备环境

参照[教程][0]使用 Docker 启动 MySQL 实例。

[0]: https://github.com/chaseSpace/go-common-pkg-exmaples/blob/master/_dockerfile/mysql/light.md


进入 mysql shell，查看版本和隔离级别（默认RR）：

```plain
mysql> select version();
+-----------+
| version() |
+-----------+
| 5.7.44    |
+-----------+
mysql> SELECT @@transaction_isolation;
+-------------------------+
| @@transaction_isolation |
+-------------------------+
| REPEATABLE-READ         |
+-------------------------+
```

### 创建测试表

执行下面的 sql 创建测试表：

```
CREATE DATABASE IF NOT EXISTS testdb;
USE testdb;

drop table students_lock;
-- truncate table students_lock;

CREATE TABLE students_lock
(
    id    INT PRIMARY KEY,
    name  VARCHAR(50),
    score INT,
    key   idx_score(score)
);

INSERT INTO students_lock (id, name, score)
VALUES (5, 'Alice', 5),
       (10, 'Bob', 10),
       (15, 'Carol', 15),
       (20, 'Ciry', 20);
```

## 实战：记录锁

记录锁是索引记录上的锁，也是粒度最小最容易理解的锁，冲突概率最低，并发度最高。即使定义的表没有索引，InnoDB
创建一个隐藏的聚簇索引，并使用该索引进行记录锁定全部记录（相当于表级锁）。

要想SQL语句只锁住索引记录，要求SQL语句必须是唯一索引+等值查询（含`=`, `IN`和`IS NULL`）+完全命中记录。

> [!NOTE]
> 这里讨论的唯一索引包含了主键索引，它们的锁行为一致。

### 案例1：唯一索引+等值查询+命中记录

下面开始测试记录锁。

#### 事务 A

```plain
BEGIN;
SELECT * FROM students_lock WHERE id = 10 FOR UPDATE;
```

预期锁住id为10的单条记录，此处应用[加锁规则](#加锁规则)中的**优化1**。

#### 事务 B

```plain
BEGIN;

-- 锁住 id=10 的记录
SELECT * FROM students_lock WHERE id = 10 FOR UPDATE; -- 阻塞
UPDATE students_lock SET id = 10 WHERE id = 10; -- 阻塞

-- 不锁(-inf, 10) 和 (10, +inf)
INSERT INTO students_lock VALUES(9, 'Dave', 9); -- 非阻塞
INSERT INTO students_lock VALUES(11, 'Dave', 11); -- 非阻塞

ROLLBACK;
```

#### 查看锁等待信息

```
mysql> SHOW ENGINE INNODB STATUS\G
...
------- TRX HAS BEEN WAITING 25 SEC FOR THIS LOCK TO BE GRANTED:
RECORD LOCKS space id 33 page no 3 n bits 80 index PRIMARY of table `testdb`.`students_lock` trx id 6069 lock_mode X locks rec but not gap waiting
Record lock, heap no 5 PHYSICAL RECORD: n_fields 5; compact format; info bits 0
...
```

摘取关键字`locks rec but not gap`，表示使用记录锁算法。

> [!NOTE]
> `lock_mode X waiting`表示临键锁，`locks gap before rec`表示间隙锁。

### 案例2：唯一索引+IN查询+全部命中

IN查询中的全部索引值必须全部命中，否则会退化为间隙锁或临键锁，后续章节会举例说明。

#### 事务 A

```plain
BEGIN;
SELECT * FROM students_lock WHERE id IN (10,15) FOR UPDATE;
```

#### 事务 B

```
BEGIN;

-- 不锁(-inf, 10)、(15, +inf)
SELECT * FROM students_lock WHERE id = 5 FOR UPDATE; -- 非阻塞
SELECT * FROM students_lock WHERE id = 20 FOR UPDATE; -- 非阻塞
INSERT INTO students_lock VALUES(9, 'Dave', 9); -- 非阻塞
INSERT INTO students_lock VALUES(11, 'Dave', 11); -- 非阻塞
INSERT INTO students_lock VALUES(16, 'Dave', 16); -- 非阻塞

-- 锁住 id=10,15 的记录
SELECT * FROM students_lock WHERE id = 10 FOR UPDATE; -- 阻塞
SELECT * FROM students_lock WHERE id = 15 FOR UPDATE; -- 阻塞

ROLLBACK;
```

## 实战：间隙锁

间隙锁只锁住索引记录之间的空隙，不含记录。当然，根据查询条件不同，这个空隙也可能包含正无穷和负无穷。
间隙锁的锁范围大于记录锁，一般在记录不存在时应用。查询条件包含等值查询和范围查询。

### 案例1：唯一索引+等值查询+未命中记录

#### 事务 A

```plain
BEGIN;
SELECT * FROM students_lock WHERE id = 7 FOR UPDATE;
```

#### 事务 B

```
BEGIN;

-- 不锁(-inf, 5] 和 [10, +inf)
SELECT * FROM students_lock WHERE id = 5 FOR UPDATE; -- 非阻塞
SELECT * FROM students_lock WHERE id = 10 FOR UPDATE; -- 非阻塞
INSERT INTO students_lock VALUES(4, 'Dave', 4); -- 非阻塞
INSERT INTO students_lock VALUES(11, 'Dave', 11); -- 非阻塞

-- 锁住(5, 10)中的间隙
INSERT INTO students_lock VALUES(6, 'Dave', 6); -- 阻塞
INSERT INTO students_lock VALUES(8, 'Dave', 8); -- 阻塞
```

首先根据原则1使用临键锁锁住`(5, 10]`，同时又根据**优化2**退化为间隙锁。

### 案例2：普通索引+等值查询+未命中记录

此案例和案例1区别不大，但注意锁应用的索引对象不同。

#### 事务 A

```plain
BEGIN;
SELECT * FROM students_lock WHERE score = 7 FOR UPDATE;
```

#### 事务 B

```
BEGIN;

-- 不锁score索引上的(-inf, 5] 和 [10, +inf)
SELECT * FROM students_lock WHERE score = 5 FOR UPDATE;
SELECT * FROM students_lock WHERE score = 10 FOR UPDATE;
INSERT INTO students_lock VALUES(4, 'Dave', 4); -- 非阻塞
INSERT INTO students_lock VALUES(11, 'Dave', 11); -- 非阻塞

-- 锁住score索引上的(5, 10)中的间隙
INSERT INTO students_lock VALUES(6, 'Dave', 6); -- 阻塞
INSERT INTO students_lock VALUES(8, 'Dave', 8); -- 阻塞
```

### 案例3：普通索引+范围查询+未命中记录

#### 事务 A

```plain
BEGIN;
SELECT * FROM students_lock WHERE score > 20 FOR UPDATE;
```

#### 事务 B

```
BEGIN;

-- 不锁(-inf, 20]
SELECT * FROM students_lock WHERE score = 5 FOR UPDATE; -- 非阻塞
SELECT * FROM students_lock WHERE score = 20 FOR UPDATE; -- 非阻塞
INSERT INTO students_lock VALUES(18, 'Dave', 18);  -- 非阻塞
INSERT INTO students_lock VALUES(19, 'Dave', 20);  -- 非阻塞*

-- 锁住(20, +inf)
INSERT INTO students_lock VALUES(21, 'Dave', 20); -- 阻塞*
INSERT INTO students_lock VALUES(21, 'Dave', 21); -- 阻塞
```

#### 反直觉的情况

**为什么锁住{21, 20}，却没有锁住{19, 20}？**

上面打了`*`的语句是需要注意的项。为什么score列都是20，但不同的id值却有不同的锁定情况。这不是随机的，是我们之前没有正确理解
**行锁只针对索引记录**这句话。

不论是唯一索引还是普通索引，行级锁都是锁的**索引记录**，而不仅仅是索引列本身。
我们应当知道，普通索引也是一颗B+树，所有索引记录都存在叶子节点中，并且通过指针有序串联起来，
**顺序上先按照索引列排序，相同值再按照主键id排序**。其中每个二级索引记录包含的是主键id和索引值，不含其它字段。

在事务B开始前，表中的`idx_score`二级索引记录排列顺序如下：

```
-- idx_score索引不含name列，这里忽略。
...
{10, 10}
{15, 15}
{20, 20}
```

当我们说锁住`(20, +inf)`时，实际上说的是锁定`({id=21, score=21}, +inf)`。也就是说，这里必须将索引中的全部字段拿出来讨论。那么根据索引记录排序规则，
`{19, 20}`不会进入前述间隙，而`{21, 20}`一定会进入前述间隙。

## 实战：临键锁

临键锁由记录锁和间隙锁组成，锁定的范围是一个左开右闭区间，例如`(5, 10]`。它是三种算法中粒度最大，也是分析起来最复杂（但也还好）的算法。

### 案例1：唯一索引+范围查询+命中记录

#### 事务 A

todo

## 参考

- [极客时间：MySQL 实战 45 讲](https://lianglianglee.com/极客时间/MySQL实战45讲.md)
- [MySQL 手册：Innodb Locking](https://dev.mysql.com/doc/refman/5.7/en/innodb-locking.html)