# 验证间隙锁

## 理论

锁定索引记录之间的间隙，或者第一个索引记录之前或最后一个索引记录之后的间隙。但是不包含记录本身。只存在于可重复读（RR）隔离级别，目的是为了解决
RR 隔离级别下幻读的现象。

间隙锁包含 S 锁和 X 锁，且相互兼容。间隙锁和行级排他锁互斥，它的唯一目的是防止幻读，即避免其他事务在间隙中插入数据。

**产生的情况**

- 对于唯一索引列：
  - 当锁住的记录不存在时，会产生记录锁和间隙锁。例如`WHERE id = 5 FOR UPDATE`。
  - 当锁住一个范围内的记录时，会产生间隙锁。例如`WHERE id BETWEEN 1 AND 5 FOR UPDATE`。
- 对于非唯一索引列：
  - 任何查询都会产生间隙锁，包括等值查询和范围查询。

**获取锁的方式**

大部分情况下使用大小于或`BETWEEN`这样的范围查询加上排他锁`FOR UPDATE`获取间隙锁。但因为有时候等值查询也会获取间隙锁，
所以很多时候是隐式获取。

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

```sql
-- 注意：必须为条件列建立非唯一索引，否则锁全表，下文会验证
CREATE TABLE students
(
    id    INT PRIMARY KEY,
    name  VARCHAR(50),
    score INT,
    key   idx_score(score)
);

INSERT INTO students (id, name, score)
VALUES (1, 'Alice', 85),
       (2, 'Bob', 90),
       (3, 'Carol', 95);
```

表中`score`列的隐藏间隙：

- (-inf, 85]
- (85, 90]
- (90, 95]
- (95, +inf]

不论是等值查询还是范围查询，只要命中间隙，就会获得对应间隙的锁。

## 验证 1：唯一索引中的间隙锁

TODO

## 验证 2：非唯一索引中的间隙锁

表中的`score`字段是非唯一索引。

使用`SELECT ... WHERE score BETWEEN 85 and 90`获取 (85, 90) 的间隙锁，其他事务无法插入这个范围的数据，直到获得锁的事务提交。

事务 A：

```mysql
mysql>
select connection_id() \G *************************** 1. row ***************************
connection_id(): 2


START TRANSACTION;

-- 获取间隙锁 （85, 95）
SELECT *
FROM students
WHERE score BETWEEN 85 and 95 FOR
UPDATE;
```

事务 B：

```mysql
mysql>
select connection_id() \G *************************** 1. row ***************************
connection_id(): 3

INSERT INTO students (id, name, score)
VALUES (4, 'Dave', 86);
-- 阻塞！默认50s

-- Crtl+C 手动退出

todo

```

