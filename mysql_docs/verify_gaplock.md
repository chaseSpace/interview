# 验证间隙锁

## 理论

锁定一个范围，但是不包含记录本身，即**开区间**，不包括双端端点，如`(1,5)`。冲突概率大于记录锁。只存在于（RR）隔离级别，
目的是为了解决 RR 隔离级别下幻读的现象。此外，间隙锁只应用于非唯一索引列。唯一索引列（含主键）适用于记录锁。

间隙锁也包含 S 锁和 X 锁，由于相互兼容，所以不需要特别区分。间隙锁和行级排他锁互斥，它的唯一目的是防止幻读，即避免其他事务在间隙中插入数据，
造成 RR 及以上级别中一个事务多次读取时结果不一致的情况。

**使用的时机**

- 唯一索引+等值查询+记录不存在
- 唯一索引+非等值查询（范围或模糊查询）+记录不存在
- 普通索引（任意查询）

注意大前提是，对于读都要显式加锁，对于插入更新是隐式加锁。

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
CREATE
DATABASE IF NOT EXISTS testdb;
use
testdb;
-- 注意：必须为条件列建立非唯一索引，否则锁全表，下文会验证
CREATE TABLE students_gap_lock
(
    id    INT PRIMARY KEY,
    name  VARCHAR(50),
    score INT,
    key   idx_score(score)
);

INSERT INTO students_gap_lock (id, name, score)
VALUES (1, 'Alice', 85),
       (2, 'Bob', 90),
       (3, 'Carol', 95);
```

表中`score`列的隐藏间隙：

- (-inf, 85]
- (85, 90]
- (90, 95]
- (95, +inf)

## 正例 1：唯一索引+等值查询+记录不存在

### 事务A

```plain
BEGIN;
SELECT * FROM students_rec_lock WHERE id = 4 FOR UPDATE;
```

预期锁住间隙`(4, +inf)`，之所以是正无穷是因为表中最大记录是3，否则取表中存在的比4大的记录作为右边界。

### 事务B

```plain
BEGIN;

INSERT INTO students_rec_lock VALUES(4, 'Dave', 86); -- 阻塞
INSERT INTO students_rec_lock VALUES(5, 'Dave', 86); -- 阻塞
INSERT INTO students_rec_lock VALUES(1000, 'Dave', 86); -- 阻塞

SELECT * FROM students_rec_lock WHERE id = 4 FOR UPDATE; -- 非阻塞（间隙锁不锁边界）
UPDATE students_rec_lock SET score = 100 WHERE id = 4; -- 非阻塞（间隙锁不锁边界）

INSERT INTO students_rec_lock VALUES(0, 'Dave', 86); -- 成功（左边界不受影响）

ROLLBACK;
```

注意，上面的事务最终都会回滚。

> [!IMPORTANT]
> 如上可知，唯一索引+等值查询+记录不存在时，存在间隙锁，但只阻塞插入操作，不阻塞查询和更新。

### 查看 Innodb 锁信息

在事务B执行被阻塞的SQL时，使用单独的会话查询锁信息，如下。

```sql
SELECT *
FROM INFORMATION_SCHEMA.INNODB_LOCKS;
+-------------+-------------+-----------+-----------+------------------------------+------------+------------+-----------+----------+------------------------+
| lock_id     | lock_trx_id | lock_mode | lock_type | lock_table                   | lock_index | lock_space | lock_page | lock_rec | lock_data              |
+-------------+-------------+-----------+-----------+------------------------------+------------+------------+-----------+----------+------------------------+
| 5894:33:3:1 | 5894        | X         | RECORD    | `testdb`.`students_rec_lock` | PRIMARY    |         33 |         3 |        1 | supremum pseudo-record |
| 5893:33:3:1 | 5893        | X         | RECORD    | `testdb`.`students_rec_lock` | PRIMARY    |         33 |         3 |        1 | supremum pseudo-record |
+-------------+-------------+-----------+-----------+------------------------------+------------+------------+-----------+----------+------------------------+
```

如上可知，间隙锁无法通过此命令查看。`SHOW ENGINE INNODB STATUS\G`输出中也显示为记录锁。部分输出为：

```
------- TRX HAS BEEN WAITING 49 SEC FOR THIS LOCK TO BE GRANTED:
RECORD LOCKS space id 33 page no 3 n bits 72 index PRIMARY of table `testdb`.`students_rec_lock` trx id 5894 lock_mode X insert intention waiting
```

## 正例 2：唯一索引+非等值查询

非等值查询包括大小于和`like`查询，下面以`>=`为例。

### 事务A

```plain
BEGIN;
SELECT * FROM students_rec_lock WHERE id >= 3 FOR UPDATE;
```

预期锁住间隙`[3, +inf)`，注意这里包含3是因为Where条件含3。

### 事务B

```plain
BEGIN;

INSERT INTO students_rec_lock VALUES(4, 'Dave', 86); -- 阻塞
INSERT INTO students_rec_lock VALUES(5, 'Dave', 86); -- 阻塞
INSERT INTO students_rec_lock VALUES(1000, 'Dave', 86); -- 阻塞

UPDATE students_rec_lock SET score = 100 WHERE id = 3; -- 阻塞，因为在间隙内 [3,+inf)

-- 间隙内不存在的记录的查改不阻塞
SELECT * FROM students_rec_lock WHERE id = 4 FOR UPDATE; -- 非阻塞
UPDATE students_rec_lock SET score = 100 WHERE id = 4; -- 非阻塞

INSERT INTO students_rec_lock VALUES(0, 'Dave', 86); -- 成功，不影响(-inf, 1)

ROLLBACK;
```

## 正例 3：普通索引

TODO
