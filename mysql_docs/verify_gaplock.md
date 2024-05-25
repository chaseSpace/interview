# 验证间隙锁

## 理论

锁定一个范围，但是不包含记录本身，即**开区间**，不包括双端端点，如`(1,5)`。冲突概率大于记录锁。只存在于（RR）隔离级别，
目的是为了解决 RR 隔离级别下幻读的现象。此外，间隙锁只应用于非唯一索引列。唯一索引列（含主键）适用于记录锁。

间隙锁也包含 S 锁和 X 锁，由于相互兼容，所以不需要特别区分。间隙锁和行级排他锁互斥，它的唯一目的是防止幻读，即避免其他事务在间隙中插入数据，
造成 RR 级别中一个事务多次读取时结果不一致的情况。

**使用的时机**

- 唯一索引+等值查询+记录不存在
- 唯一索引+非等值查询（范围或模糊查询）+记录不存在
- 普通索引+任意查询+记录不存在

> [!TIP]
> 间隙锁的记忆关键是：有索引+记录不存在。

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

-- drop table students_gap_lock;

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

> [!NOTE]
> 要求不能命中记录，否则使用记录锁。

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
+-------------+-------------+-----------+-----------+------------------------------+------------+------------+-----------+----------+------------------------+
| lock_id     | lock_trx_id | lock_mode | lock_type | lock_table                   | lock_index | lock_space | lock_page | lock_rec | lock_data              |
+-------------+-------------+-----------+-----------+------------------------------+------------+------------+-----------+----------+------------------------+
| 5894:33:3:1 | 5894        | X         | RECORD    | `testdb`.`students_gap_lock` | PRIMARY    |         33 |         3 |        1 | supremum pseudo-record |
| 5893:33:3:1 | 5893        | X         | RECORD    | `testdb`.`students_gap_lock` | PRIMARY    |         33 |         3 |        1 | supremum pseudo-record |
+-------------+-------------+-----------+-----------+------------------------------+------------+------------+-----------+----------+------------------------+
```

如上可知，间隙锁无法通过此命令查看。`SHOW ENGINE INNODB STATUS\G`输出中也显示为记录锁。部分输出为：

```plain
------- TRX HAS BEEN WAITING 49 SEC FOR THIS LOCK TO BE GRANTED:
RECORD LOCKS space id 33 page no 3 n bits 72 index PRIMARY of table `testdb`.`students_gap_lock` trx id 5894 lock_mode X insert intention waiting
```

## 正例 2：唯一索引+非等值查询+记录不存在

非等值查询包括大小于和`Like`查询，下面以`>`为例。

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

对于普通索引的等值查询，会锁住条件所在的间隙，不含边界记录。

> [!NOTE]
> 要求不能命中记录，否则使用临键锁+间隙锁，请查看[验证临键锁](verify_nextkeylock.md)。

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
