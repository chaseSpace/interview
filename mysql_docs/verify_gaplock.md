# 验证间隙锁

内容完成度99%。

> [!NOTE]
> 本篇文档由笔者结合众多资料亲自验证后编写，但其中仍可能存在错误，请保持怀疑态度阅读。

## 理论

间隙锁的锁定粒度大于记录锁，但小于临键锁。

间隙锁锁定一个范围，但是不包含记录本身，即**开区间**，不包括双端端点，如`(1,5)`。只存在于（RR）隔离级别，
目的是为了解决 RR 隔离级别下幻读的现象。此外，间隙锁只应用于非唯一索引列。唯一索引列（含主键）适用于记录锁。

间隙锁也包含 S 锁和 X 锁，由于相互兼容，所以不需要区分。但间隙锁和其他任何排他锁互斥。

**使用的时机**

- 唯一索引+等值查询+记录不存在
- 唯一索引+非等值查询（范围或模糊查询）+记录不存在
- 普通索引+等值查询+记录不存在
    - 若记录存在，就使用临键锁了。

## 准备环境

参照[教程][0]使用 Docker 启动 MySQL 实例。

[0]: https://github.com/chaseSpace/go-common-pkg-exmaples/blob/master/_dockerfile/mysql/light.md


进入 mysql shell，查看版本和隔离级别（默认也是 RR）：

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

执行下面的 sql 创建测试表：

```
CREATE DATABASE IF NOT EXISTS testdb;
USE testdb;

drop table students_gap_lock;

-- 包括一个主键索引 和 一个普通索引
CREATE TABLE students_gap_lock
(
    id    INT PRIMARY KEY,
    name  VARCHAR(50),
    score INT,
    key   idx_score(score)
);

INSERT INTO students_gap_lock (id, name, score)
VALUES (1, 'Alice', 85),
       (4, 'Bob', 90),
       (7, 'Carol', 95);
```

## 正例 1：唯一索引+等值查询+记录不存在

### 事务 A

```plain
BEGIN;
SELECT * FROM students_gap_lock WHERE id = 5 FOR UPDATE;
```

预期锁住间隙`(4, 7)`。

### 事务 B

```plain
BEGIN;

-- 不锁 (-inf, 1]
INSERT INTO students_gap_lock VALUES(0, 'Dave', 84); -- 非阻塞
UPDATE students_gap_lock SET id=1 WHERE id=1; -- 非阻塞

-- 不锁 (1, 4]
INSERT INTO students_gap_lock VALUES(2, 'Dave', 84); -- 非阻塞
UPDATE students_gap_lock SET id=4 WHERE id=4; -- 非阻塞

-- 锁住 (4, 7)
INSERT INTO students_gap_lock VALUES(5, 'Dave', 84); -- 阻塞
INSERT INTO students_gap_lock VALUES(6, 'Dave', 84); -- 阻塞

-- 不锁 [7, +inf)
UPDATE students_gap_lock SET id=7 WHERE id=7; -- 非阻塞
INSERT INTO students_gap_lock VALUES(8, 'Dave', 84); -- 非阻塞

ROLLBACK;
```

### 查看 Innodb 锁信息

在事务 B 执行被阻塞的 SQL 时，使用单独的会话查询锁信息，如下。

```sql
SELECT *
FROM INFORMATION_SCHEMA.INNODB_LOCKS;
+---------------+-------------+-----------+-----------+------------------------------+------------+------------+-----------+----------+-----------+
| lock_id       | lock_trx_id | lock_mode | lock_type | lock_table                   | lock_index | lock_space | lock_page | lock_rec | lock_data |
+---------------+-------------+-----------+-----------+------------------------------+------------+------------+-----------+----------+-----------+
| 503561:31:3:4 | 503561      | X,GAP     | RECORD    | `testdb`.`students_gap_lock` | PRIMARY    |         31 |         3 |        4 | 7         |
| 503560:31:3:4 | 503560      | X,GAP     | RECORD    | `testdb`.`students_gap_lock` | PRIMARY    |         31 |         3 |        4 | 7         |
+---------------+-------------+-----------+-----------+------------------------------+------------+------------+-----------+----------+-----------+
```

## 正例 2：唯一索引+非等值查询+记录不存在

非等值查询包括范围和`LIKE`查询，下面以`>`为例。

> [!NOTE]
> 要求不能命中记录，否则使用临键锁。

### 事务 A

```plain
BEGIN;
SELECT * FROM students_gap_lock WHERE id > 7 FOR UPDATE;
```

预期锁住主键间隙`(7, +inf)`。

### 事务 B

```plain
BEGIN;

-- 不锁 (-inf, 1]
INSERT INTO students_gap_lock VALUES(0, 'Dave', 84); -- 非阻塞
UPDATE students_gap_lock SET id=1 WHERE id=1; -- 非阻塞

-- 不锁 (1, 4]
INSERT INTO students_gap_lock VALUES(2, 'Dave', 84); -- 非阻塞
UPDATE students_gap_lock SET id=4 WHERE id=4; -- 非阻塞

-- 不锁 (4, 7]
INSERT INTO students_gap_lock VALUES(5, 'Dave', 84); -- 非阻塞
UPDATE students_gap_lock SET id=7 WHERE id=7; -- 非阻塞

-- 锁住 (7, +inf)
INSERT INTO students_gap_lock VALUES(8, 'Dave', 84); -- 阻塞
INSERT INTO students_gap_lock VALUES(9, 'Dave', 84); -- 阻塞

ROLLBACK;
```

## 正例 3：普通索引+等值查询+记录不存在

对于普通索引的等值查询，仅当记录不存在时，会锁住记录所在的间隙。

> [!NOTE]
> 普通索引的等值查询，当记录存在时，会以临键锁锁住记录两侧间隙，含右边界。

### 事务A

```
BEGIN;
SELECT * FROM students_gap_lock WHERE score = 91 FOR UPDATE;
```

预期锁住间隙 `(90, 95)`。

### 事务B

```
BEGIN;

-- 不锁 (-inf, 85]
INSERT INTO students_gap_lock VALUES(8, 'Dave', 84); -- 非阻塞
UPDATE students_gap_lock SET score = 1 WHERE score=85; -- 非阻塞

-- 不锁 (85, 90]
INSERT INTO students_gap_lock VALUES(9, 'Dave', 86); -- 非阻塞
UPDATE students_gap_lock SET score = 1 WHERE score=90; -- 非阻塞

-- 锁住 (90, 95)
INSERT INTO students_gap_lock VALUES(10, 'Dave', 91); -- 阻塞
INSERT INTO students_gap_lock VALUES(10, 'Dave', 94); -- 阻塞

-- 不锁 [95, +inf)
UPDATE students_gap_lock SET score = 1 WHERE score=95; -- 非阻塞
INSERT INTO students_gap_lock VALUES(10, 'Dave', 96); -- 非阻塞

ROLLBACK;
```

## 重要：正确理解间隙锁

在上个章节中，讨论的情况是：普通索引+等值查询+记录不存在。但该节并未列出所有可能的情况，特别是一些反直觉的锁情况。
为了方便理解，在本节中单独进行讨论。

### 事务A

与上节一致。

```
BEGIN;
SELECT * FROM students_gap_lock WHERE score = 91 FOR UPDATE;
```

预期锁住间隙 `(90, 95)`。

### 事务B

这里列出了一些不同的锁触发情况。

```
BEGIN;

-- 不锁 90
INSERT INTO students_gap_lock VALUES(2, 'Dave', 90); -- 非阻塞
INSERT INTO students_gap_lock VALUES(3, 'Dave', 90); -- 非阻塞

-- 锁住 90
INSERT INTO students_gap_lock VALUES(5, 'Dave', 90); -- 阻塞
INSERT INTO students_gap_lock VALUES(6, 'Dave', 90); -- 阻塞

-- 锁住 95
INSERT INTO students_gap_lock VALUES(5, 'Dave', 95); -- 阻塞
INSERT INTO students_gap_lock VALUES(6, 'Dave', 95); -- 阻塞

-- 不锁 95
INSERT INTO students_gap_lock VALUES(8, 'Dave', 95); -- 非阻塞
INSERT INTO students_gap_lock VALUES(9, 'Dave', 95); -- 非阻塞

ROLLBACK;
```

这里你会发现，对于边界值锁定，发生了奇怪的现象。对于相同的索引值，为什么有的会阻塞，有的不会阻塞？

这是因为，不论是唯一索引还是普通索引，行级锁都是锁的**索引记录**，而不仅仅是索引列本身。
我们应当知道，普通索引也是一颗B+树，所有索引记录都存在叶子节点中，并且通过指针有序串联起来，
**顺序上先按照索引列排序，相同值再按照主键id排序**。其中每个二级索引记录包含的是主键id和索引值，不含其它字段。

在事务B开始前，表中的二级索引记录排列顺序如下：

```
(1, 85)

(4, 90)

(7, 95)
```

当我们说锁住`(90, 95)`时，实际上指的是锁住 `(4, 90)` 和 `(7, 95)` 这两条索引记录之间的空隙。也就是说，凡是能插入这个空隙的索引记录都会被锁住。
哪些记录会进入这个空隙呢？参考下面的几条记录：

```
(3, 90) -- 不会进入空隙
(5, 90) -- 会进入空隙
(4, 91) -- 会进入空隙

(6, 95) -- 会进入空隙
(8, 95) -- 不会进入空隙
```

当索引记录中的索引列等值时，再按照主键id进行排序，明白了这一点，就能够豁然贯通了。

> [!NOTE]
> 唯一索引的索引列具有唯一性，不会出现重复的索引记录，所以唯一索引的索引记录是**仅**按照索引列排序的，不会出现上面的情况。
> 一旦唯一索引列值位于锁范围内，不论其他字段值是多少，都会被锁住。