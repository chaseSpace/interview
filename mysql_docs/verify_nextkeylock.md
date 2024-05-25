# 验证临键锁

## 理论

记录所 + 间隙锁的组合，锁定一个范围，并且锁定边界记录本身，即**左开右闭区间**，如`(1,5]`。是默认的粒度最大的行锁算法。

**使用的时机**

- 唯一索引+非等值查询+记录存在；
- 普通索引+任意查询+记录存在；

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
CREATE TABLE students_nk_lock
(
    id    INT PRIMARY KEY,
    name  VARCHAR(50),
    score INT,
    key   idx_score(score)
);

INSERT INTO students_nk_lock (id, name, score)
VALUES (1, 'Alice', 85),
       (4, 'Bob', 90),
       (7, 'Carol', 95),
       (10, 'Lucy', 100);
```

## 正例 1：唯一索引+范围查询+命中边界记录

非等值查询包括大小于和模糊查询。这里使用范围查询，且使用`>=`匹配记录+正无穷。

### 事务 A

```
BEGIN;
SELECT * FROM students_nk_lock WHERE id >= 4 FOR UPDATE;
```

预期锁住范围`[4, 7)`和`[7, +inf)`。

### 事务 B

```plain
BEGIN;

-- 不锁 (-inf, 1]
INSERT INTO students_nk_lock VALUES(0, 'Dave', 84); -- 非阻塞
UPDATE students_nk_lock SET id=1 WHERE id=1; -- 非阻塞

-- 不锁 (1, 4)
INSERT INTO students_nk_lock VALUES(2, 'Dave', 84); -- 非阻塞
INSERT INTO students_nk_lock VALUES(3, 'Dave', 84); -- 非阻塞

-- 锁住 [4, 7)
INSERT INTO students_nk_lock VALUES(5, 'Dave', 84); -- 阻塞
UPDATE students_nk_lock SET id=4 WHERE id=4; -- 阻塞
UPDATE students_nk_lock SET id=7 WHERE id=7; -- 阻塞

-- 锁住 [7, 10)
INSERT INTO students_nk_lock VALUES(8, 'Dave', 84); -- 阻塞
INSERT INTO students_nk_lock VALUES(9, 'Dave', 84); -- 阻塞

-- 锁住 [10, +inf)
INSERT INTO students_nk_lock VALUES(11, 'Dave', 84); -- 阻塞

ROLLBACK;
```

## 正例 2：唯一索引+范围查询2+命中边界记录

本例中的范围查询使用`BETWEEN`，使用两个存在记录作为命中记录。

### 事务 A

```plain
BEGIN;
SELECT * FROM students_nk_lock WHERE id BETWEEN 4 and 7 FOR UPDATE;
```

预期锁住范围`[4, 7]`、`[7, 10]`，包含了右侧间隙记录10。

### 事务 B

```
BEGIN;

-- 不锁 (-inf, 1)
INSERT INTO students_nk_lock VALUES(0, 'Dave', 84); -- 非阻塞

-- 不锁 [1, 4)
INSERT INTO students_nk_lock VALUES(2, 'Dave', 84); -- 非阻塞
UPDATE students_nk_lock SET id=1 where id = 1; -- 非阻塞

-- 锁住 [4, 7]
INSERT INTO students_nk_lock VALUES(5, 'Dave', 84); -- 阻塞
UPDATE students_nk_lock SET id=4 where id = 4; -- 阻塞
UPDATE students_nk_lock SET id=7 where id = 7; -- 阻塞

-- 锁住 (7, 10]
INSERT INTO students_nk_lock VALUES(8, 'Dave', 84); -- 阻塞
UPDATE students_nk_lock SET id=10 where id = 10; -- 阻塞

-- 不锁 (10, +inf)
INSERT INTO students_nk_lock VALUES(11, 'Dave', 84); -- 阻塞
```

## 正例 3：普通索引+范围查询+未命中边界记录

```
BEGIN;
SELECT * FROM students_nk_lock WHERE score > 90 FOR UPDATE;
```

预期锁住间隙 `(90, +inf)`。如果是`>=90`，将会较大的锁范围，看下个例子。

### 事务B

```
BEGIN;

-- 不锁 (-inf, 85)
INSERT INTO students_nk_lock VALUES(5, 'Dave', 84); -- 非阻塞

-- 不锁 [85, 90]
INSERT INTO students_nk_lock VALUES(6, 'Dave', 86); -- 非阻塞
UPDATE students_nk_lock SET score = 85 WHERE score=85; -- 非阻塞
UPDATE students_nk_lock SET score = 90 WHERE score=90; -- 非阻塞

-- 锁住 (90, +inf)
INSERT INTO students_nk_lock VALUES(7, 'Dave', 91); -- 阻塞
UPDATE students_nk_lock SET score = 95 WHERE score=95; -- 阻塞
UPDATE students_nk_lock SET score = 100 WHERE score=100; -- 阻塞
```

## 正例 4：普通索引+范围查询2+命中边界记录

### 事务A

```
BEGIN;
SELECT * FROM students_nk_lock WHERE score >=90 FOR UPDATE;
```

预期锁住90左右的全部范围，包含数据。

### 事务B

```
BEGIN;

-- 锁住 (-inf, 85)
INSERT INTO students_nk_lock VALUES(5, 'Dave', 84); -- 阻塞

-- 锁住 [85, 90)
INSERT INTO students_nk_lock VALUES(6, 'Dave', 86); -- 阻塞
UPDATE students_nk_lock SET score = 85 WHERE score=85; -- 阻塞

-- 锁住 [90, +inf)
INSERT INTO students_nk_lock VALUES(6, 'Dave', 91); -- 阻塞
UPDATE students_nk_lock SET score = 90 WHERE score=90; -- 阻塞
UPDATE students_nk_lock SET score = 95 WHERE score=95; -- 阻塞
INSERT INTO students_nk_lock VALUES(6, 'Dave', 96); -- 阻塞
INSERT INTO students_nk_lock VALUES(6, 'Dave', 101); -- 阻塞

ROLLBACK;
```

结果有点逆天。

## 正例 5：普通索引+范围查询3+命中边界记录

### 事务A

```
BEGIN;
SELECT * FROM students_nk_lock WHERE score BETWEEN 90 and 95 FOR UPDATE;
```

预期锁住90和95分别的左右侧间隙，包含边界数据。

### 事务B

```
BEGIN;

-- 不锁 (-inf, 85]
INSERT INTO students_nk_lock VALUES(5, 'Dave', 84); -- 非阻塞
UPDATE students_nk_lock SET score = 85 WHERE score=85; -- 非阻塞

-- 锁住 (85, 90]
INSERT INTO students_nk_lock VALUES(6, 'Dave', 86); -- 阻塞
UPDATE students_nk_lock SET score = 90 WHERE score=90; -- 阻塞

-- 锁住 (90, 95]
INSERT INTO students_nk_lock VALUES(6, 'Dave', 91); -- 阻塞
UPDATE students_nk_lock SET score = 95 WHERE score=95; -- 阻塞

-- 锁住 (95, 100]
INSERT INTO students_nk_lock VALUES(6, 'Dave', 96); -- 阻塞
UPDATE students_nk_lock SET score = 100 WHERE score=100; -- 阻塞

-- 不锁 (100, +inf)
INSERT INTO students_nk_lock VALUES(6, 'Dave', 101); -- 非阻塞
```

## 小结

- 唯一索引+正无穷查询+命中边界记录：锁住记录及正无穷范围；
- 唯一索引+有限范围查询+未命中边界记录：锁住记录及其右侧间隙，含边界数据；
- 普通索引+正无穷查询+未命中边界记录：锁住记录及正无穷范围；
- 普通索引+正无穷查询+命中边界记录：锁住整个索引（原因未知？）；
- 普通索引+有限范围查询+命中边界记录：锁住记录及其两侧间隙，含边界数据；

规律不明显，但普通索引的临键锁范围显然大于唯一索引。