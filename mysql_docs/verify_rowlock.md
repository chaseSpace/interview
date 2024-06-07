# 验证行锁的三种算法

行锁的三种算法：记录锁、间隙锁和临键锁。

- 记录锁：只锁住索引记录。
- 间隙锁：锁住索引记录之间的间隙，是左开右开区间，如`(1, 3)`。
- 临键锁：锁住索引记录和索引记录之间的间隙，是左开右闭区间，如`(1, 3]`。

> [!NOTE]
> 这里提到的索引记录，要么是普通索引，要么是主键索引，根据索引使用情况来辨别。

## 加锁规则

Innodb 默认的加锁算法是 Next-Key Lock，即临键锁。但在一些场景下会退化为粒度更小的间隙锁或者记录锁。总的来说，遵循以下规则：

- 原则 1：加锁的默认算法是 Next-Key Lock；
- 原则 2：查找过程中访问到的对象会加锁；
- 优化 1：唯一索引上的等值查询，给命中的记录加锁时，退化为记录锁；
- 优化 2：索引上的等值查询，未命中记录时，退化为间隙锁；
- bug1：唯一索引上的范围查询会访问到不满足条件的第一个值为止，即会多锁定一个间隙（由于原则 1，所以含右边界）；

为了确认以上规则，下文会尽可能列出多的加锁案例，并运用以上规则进行解释，若读者发现不合理的地方，请留言。

> [!NOTE]
以上规则并不来自 MySQL 官方，而是来自极客时间[MySQL 实战 45 讲-第 21 章](https://time.geekbang.org/column/intro/100020801)。
> 笔者结合了网络上的其他资料，只有这个链接的作者才对以上规则有详细的解释，并结合了大量案例进行说明。
> 同时，本文也包含了笔者的更多实验结果。

下面，请同笔者一起进入 Innodb 加锁的世界！

## 关于加锁

行锁的加锁方式有两种：一种是 SELECT 加`FOR UPDATE` 或 `LOCK IN SHARE MODE`，不加修饰则是快照读，无任何锁；
另一种是插入、更新、删除等操作自动隐式加锁。

## 准备环境

参照[教程][0]使用 Docker 启动 MySQL 实例。

[0]: https://github.com/chaseSpace/go-common-pkg-exmaples/blob/master/_dockerfile/mysql/light.md


进入 mysql shell，查看版本和隔离级别（默认 RR）：

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

```plain
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

要想 SQL 语句只锁住索引记录，要求 SQL 语句必须是唯一索引+等值查询（含`=`, `IN`和`IS NULL`）+完全命中记录。

> [!NOTE]
> 这里讨论的唯一索引包含了主键索引，它们的锁行为一致。

### 案例 1：唯一索引+等值查询+命中记录

下面开始测试记录锁。

#### 事务 A

```plain
BEGIN;
SELECT * FROM students_lock WHERE id = 10 FOR UPDATE;
```

预期锁住 id 为 10 的单条记录，此处应用[加锁规则](#加锁规则)中的**优化 1**。

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

```plain
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

### 案例 2：唯一索引+IN 查询+全部命中

IN 查询中的全部索引值必须全部命中，否则会退化为间隙锁或临键锁，后续章节会举例说明。

#### 事务 A

```plain
BEGIN;
SELECT * FROM students_lock WHERE id IN (10,15) FOR UPDATE;
```

#### 事务 B

```plain
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

### 案例 1：唯一索引+等值查询+未命中记录

#### 事务 A

```plain
BEGIN;
SELECT * FROM students_lock WHERE id = 7 FOR UPDATE;
```

#### 事务 B

```plain
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

首先根据原则 1 使用临键锁锁住`(5, 10]`，同时又根据**优化 2**退化为间隙锁。

### 案例 2：普通索引+等值查询+未命中记录

此案例和案例 1 区别不大，但注意锁应用的索引对象不同。

#### 事务 A

```plain
BEGIN;
SELECT * FROM students_lock WHERE score = 7 FOR UPDATE;
```

#### 事务 B

```plain
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

### 案例 3：普通索引+范围查询+未命中记录

#### 事务 A

```plain
BEGIN;
SELECT * FROM students_lock WHERE score > 20 FOR UPDATE;
```

#### 事务 B

```plain
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

#### 重点：锁住索引记录

**为什么锁住{21, 20}，却没有锁住{19, 20}？**

上面打了`*`的语句是需要注意的项。为什么 score 列都是 20，但不同的 id 值却有不同的锁定情况。这不是随机的，是我们之前没有正确理解
**行锁只针对索引记录**这句话。

不论是唯一索引还是普通索引，行级锁都是锁的**索引记录**，而不仅仅是索引列本身。
我们应当知道，普通索引也是一颗 B+树，所有索引记录都存在叶子节点中，并且通过指针有序串联起来，
**顺序上先按照索引列排序，相同值再按照主键 id 排序**。其中每个二级索引记录包含的是主键 id 和索引值，不含其它字段。

在事务 B 开始前，表中的`idx_score`二级索引记录排列顺序如下：

```plain
-- idx_score索引不含name列，这里忽略。
...
{10, 10}
{15, 15}
{20, 20}
```

当我们说锁住`(20, +inf)`时，实际上说的是锁定`({id=21, score=21}, +inf)`。也就是说，这里必须将索引中的全部字段拿出来讨论。那么根据索引记录排序规则，
`{19, 20}`不会进入前述间隙，而`{21, 20}`一定会进入前述间隙。

### 注意

并不是**未命中记录**就一定是使用间隙锁，比如 `WHERE score > 15 and score < 20`
没有命中记录，但由于默认加临键锁，所以锁住`(15, 20]`。

## 实战：临键锁

临键锁由记录锁和间隙锁组成，锁定的范围是一个左开右闭区间，例如`(5, 10]`。它是三种算法中粒度最大，也是分析起来最复杂（但也还好）的算法。

### 案例 1：唯一索引+范围查询+命中记录

#### 事务 A

```plain
BEGIN;
SELECT * FROM students_lock WHERE id between 10 and 15 FOR UPDATE;
```

根据[刚刚的分析](#反直觉的情况)，事务 A 应该锁住`[{10,10} -> {15,15}]`，即索引记录之间的间隙，以及左右边界记录。
但因为[加锁规则中的 bug1](#加锁规则)
，MySQL 还会继续向右访问到一个不满足条件的索引记录 id=20，事务 A 还会锁住`({15,15} -> {20,20}]`。

#### 事务 B

```plain
BEGIN;

-- 不锁(-inf, {10,10})、({20,20}, +inf)
SELECT * FROM students_lock WHERE id = 5 FOR UPDATE;  -- 非阻塞
INSERT INTO students_lock VALUES(9, 'Dave', 9);  -- 非阻塞
INSERT INTO students_lock VALUES(21, 'Dave', 21);  -- 非阻塞

-- 锁住[{10,10} -> {20,20}]
SELECT * FROM students_lock WHERE id = 10 FOR UPDATE; -- 阻塞
SELECT * FROM students_lock WHERE id = 15 FOR UPDATE; -- 阻塞
SELECT * FROM students_lock WHERE id = 20 FOR UPDATE; -- 阻塞
INSERT INTO students_lock VALUES(11, 'Dave', 11);  -- 阻塞
INSERT INTO students_lock VALUES(16, 'Dave', 16);  -- 阻塞
```

### 案例 2：普通索引+等值查询+命中记录

#### 事务 A

```plain
BEGIN;

SELECT * FROM students_lock WHERE score = 10 FOR UPDATE;
```

分析：根据原则 1，默认加临键锁，锁住`score`上的`(5, 10]`。然而实际并不是这样，请看下面实测。

#### 事务 B

```plain
BEGIN;

-- 不锁(-inf, 5] 和 [15, +inf)
SELECT * FROM students_lock WHERE score = 5 FOR UPDATE; -- 非阻塞
SELECT * FROM students_lock WHERE score = 15 FOR UPDATE; -- 非阻塞
SELECT * FROM students_lock WHERE score = 20 FOR UPDATE; -- 非阻塞
INSERT INTO students_lock VALUES(4, 'Dave', 4); -- 非阻塞
INSERT INTO students_lock VALUES(16, 'Dave', 16); -- 非阻塞

-- 锁住(5, 15)
SELECT * FROM students_lock WHERE score = 10 FOR UPDATE; -- 阻塞
INSERT INTO students_lock VALUES(6, 'Dave', 6); -- 阻塞
INSERT INTO students_lock VALUES(11, 'Dave', 11); -- 阻塞
```

可见，事务 A 不仅锁住了`score`列上的`(5, 10]`，还锁住了`(10, 15)`，这是为什么？请看下一小节。

#### 重点：区别看待普通索引

我们已经发现了事务 A 比预期多锁住了一个间隙`(10, 15)`，其实这并不是 bug，而是应用了[加锁规则中的优化 2](#加锁规则)。
这里我们来详细分析普通索引上的数据查找过程：

MySQL 默认按照升序从左到右遍历普通索引`score`列上的索引记录，使用**等值查询**的方式发现了`{10, 10}`
的索引记录，根据原则 1，锁住`(5,10]`；
此时扫描并未结束，由于普通索引允许存在重复列，所以 MySQL 还要继续向右遍历，直到发现了第一条不等于 10 的索引记录`{15, 15}`，
根据原则 1，锁住`(10,15]`，但根据优化 2，退化为间隙锁，即多锁一个间隙`(10, 15)`。

> [!IMPORTANT]
上述过程中，MySQL 继续遍历索引记录时，仍然使用的是**等值查询**的方式来匹配索引记录，所以能够应用加锁规则中的**优化 2**。

### 案例 3：普通索引+范围查询+命中记录

此案例可以按照案例 2 中的规律来分析，略。

### 案例 4：普通索引+范围查询+命中记录+降序排列

降序排列会影响索引记录的查找顺序。索引记录是以升序排列的，如果 SQL 语句是以降序排列，那会导致索引记录的查找顺序也是逆序的。

#### 事务 A

```plain
BEGIN;

SELECT * FROM students_lock WHERE score between 15 and 20 ORDER BY score DESC FOR UPDATE;
```

分析：MySQL 在索引树上逆序遍历索引记录，先通过等值查询找到`score`为 20 的记录。根据原则 1 和优化 2，
加上间隙锁`(20, +inf)`和临键锁`(10,15]`。
此时扫描并未结束，由于普通索引允许存在重复列，所以 MySQL 还要继续向左遍历，直到发现了第一个`score`
列不等于 15 的索引记录`{10, 10}`，扫描结束。根据原则 1，锁住`(5,10]`。所以最终锁定的是`(5, +inf)`，其中包含四个间隙，三条记录。

#### 事务 B

```plain
BEGIN;

-- 不锁(-inf, 5]
SELECT * FROM students_lock WHERE score = 5 FOR UPDATE; -- 非阻塞
INSERT INTO students_lock VALUES(4, 'Dave', 4); -- 非阻塞

-- 锁住(5, +inf)
SELECT * FROM students_lock WHERE score = 10 FOR UPDATE; -- 阻塞
SELECT * FROM students_lock WHERE score = 15 FOR UPDATE; -- 阻塞
SELECT * FROM students_lock WHERE score = 20 FOR UPDATE; -- 阻塞
INSERT INTO students_lock VALUES(6, 'Dave', 6); -- 阻塞
INSERT INTO students_lock VALUES(11, 'Dave', 11); -- 阻塞
INSERT INTO students_lock VALUES(16, 'Dave', 16); -- 阻塞
INSERT INTO students_lock VALUES(21, 'Dave', 21); -- 阻塞
```

## 实战：适用原则 2 的案例

前面讨论的案例都没有提及[加锁规则中的原则 2](#加锁规则)，不是因为它们不适用，而是因为它们无法明显体现**原则 2**。

**原则 2**讨论的是查找索引的过程中，如果只是用到了普通索引，则不会影响**普通索引上不存在的列**的`UPDATE`。
如果还用到了主键索引，则不能做对应数据行的任何增删改。那如何判断是锁住普通索引还是主键索引呢？

- 如果是 X 锁，则无论查找的索引是否主键索引，都一定会锁住主键索引；
- 如果是 S 锁，则观察 SQL 是否使用覆盖索引，若未使用覆盖索引则由于回表也会锁住主键索引（可通过`EXPLAIN`查看）；
  若使用覆盖索引，则不影响非覆盖索引上的列的更新。

前面案例中加锁的 SQL 语句都是`FOR UPDATE`加 X 锁，所以都是锁主键索引。 本节通过在【普通索引上加 S 锁+使用覆盖索引】来演示这种情况。

### 事务 A

```plain
BEGIN;

SELECT id FROM students_lock WHERE score = 15 LOCK IN SHARE MODE;
```

这个 SQL 在普通索引`idx_score`上加 S 锁，并且只锁住`score`列上的`(10, 20)`。并且此 SQL 使用覆盖索引，所以不会影响`name`列的更新。

### 事务 B

```plain
BEGIN;

-- 不锁主键索引，可以修改相同score值所属数据行中的name
UPDATE students_lock SET name = 'Lucy' WHERE id = 15; -- 非阻塞

-- 锁住 idx_score 索引上，score列上的(10, 20)
SELECT id FROM students_lock WHERE score = 15 FOR UPDATE; -- 阻塞
UPDATE students_lock SET score = 15 WHERE id = 15; -- 阻塞
```

## 实战：LIMIT 语句加锁

如果 limit 数量能够控制在较小的数值，那就可以减小锁的范围。

### 事务 A

```plain
BEGIN;

SELECT id FROM students_lock WHERE score >= 5 LIMIT 1 FOR UPDATE;
```

上述 SQL 中因为有了`LIMIT 1`，所以只会锁住`{5,5}`这一条记录。

### 事务 B

```plain
BEGIN;

-- 不锁 (-inf, 5), (5, +inf)
INSERT INTO students_lock VALUES(4, 'Dave', 4); -- 非阻塞
INSERT INTO students_lock VALUES(6, 'Dave', 6); -- 非阻塞

-- 锁住 [5,5]
SELECT id FROM students_lock WHERE score = 5 FOR UPDATE;
```

所以在实战中，如果要查询或删除数据时，可以考虑使用 limit 语句来减小锁的范围。

## 要点

首先牢记[加锁规则](#加锁规则)，其次要在脑海中推演 MySQL 在索引上的遍历过程，SQL 中的逆序排列会影响索引遍历顺序。
间隙锁很好理解，就是未命中任何记录。

## 参考

- [极客时间：MySQL 实战 45 讲](https://lianglianglee.com/极客时间/MySQL实战45讲.md)
- [MySQL 手册：Innodb Locking](https://dev.mysql.com/doc/refman/5.7/en/innodb-locking.html)