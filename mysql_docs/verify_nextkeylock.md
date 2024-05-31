# 验证临键锁

内容完成度99%。

> [!NOTE]
> 本篇文档由笔者结合众多资料亲自验证后编写，但其中仍可能存在错误，请保持怀疑态度阅读。

## 理论

记录所 + 间隙锁的组合，锁定一个范围，并且锁定**右边界**记录本身，即**左开右闭区间**，如`(1,5]`。是默认的粒度最大的行锁算法。

**使用的时机**

- 唯一索引+非等值查询+记录存在；
    - 非等值查询包含范围查询和模糊查询，示例以范围查询为主。
- 普通索引，包含以下情况：
    - 普通索引 + 等值查询 + 记录存在；
    - 普通索引 + 非等值查询 + 包含记录；
    - 普通索引 + 非等值查询 + 不含记录；

**锁范围扩散算法**

当范围查询条件匹配的边界记录存在时，直接遵循左开右闭区间规则；当范围匹配的边界记录不存在时，将范围进行左或右扩散，
直到找到一条存在的记录作为锁范围边界，然后遵循左开右闭区间规则。间隙锁也适用于这个算法。

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

### 事务 A

```
BEGIN;
SELECT * FROM students_nk_lock WHERE id >= 4 FOR UPDATE;
```

预期锁住的索引记录范围`({4, 90} -> +inf)`，其中4是id，90是score，含间隙中的记录。

**命中边界记录**

范围的左边界记录刚好存在，对于唯一索引，不会进行左扩散。

### 事务 B

```plain
BEGIN;

-- 不锁 (-inf -> {4, 90})
INSERT INTO students_nk_lock VALUES(2, 'Dave', 85); -- 非阻塞
INSERT INTO students_nk_lock VALUES(3, 'Dave', 90); -- 非阻塞
UPDATE students_nk_lock SET id=1 WHERE id=1; -- 非阻塞

-- 锁住 [{4,90} -> +inf)
INSERT INTO students_nk_lock VALUES(5, 'Dave', 90); -- 阻塞
INSERT INTO students_nk_lock VALUES(8, 'Dave', 90); -- 阻塞
UPDATE students_nk_lock SET id=4 WHERE id=4; -- 阻塞
UPDATE students_nk_lock SET id=7 WHERE id=7; -- 阻塞
UPDATE students_nk_lock SET id=10 WHERE id=10; -- 阻塞

ROLLBACK;
```

## 正例 2：唯一索引+范围查询：BETWEEN+命中边界记录

本例中的范围查询使用`BETWEEN`。

### 事务 A

```plain
BEGIN;
SELECT * FROM students_nk_lock WHERE id BETWEEN 4 and 7 FOR UPDATE;
```

预期锁住范围`[{4, 90} -> {10, 100}]`。

> [!IMPORTANT]
> 这里锁范围包含左边界`{4, 90}`是因为条件包含，包含右边界`{10, 100}`是因为临键锁的左开右闭规则。唯一索引和普通索引都会进行右扩散。

### 事务 B

```
BEGIN;

-- 不锁 (-inf -> {4,90})
INSERT INTO students_nk_lock VALUES(2, 'Dave', 85); -- 非阻塞
INSERT INTO students_nk_lock VALUES(3, 'Dave', 90); -- 非阻塞
UPDATE students_nk_lock SET id=1 WHERE id=1; -- 非阻塞

-- 锁住 [{4,90} -> {10,100}]
INSERT INTO students_nk_lock VALUES(4, 'Dave', 89); -- 阻塞
INSERT INTO students_nk_lock VALUES(5, 'Dave', 90); -- 阻塞
INSERT INTO students_nk_lock VALUES(7, 'Dave', 96); -- 阻塞
INSERT INTO students_nk_lock VALUES(8, 'Dave', 96); -- 阻塞
UPDATE students_nk_lock SET id=4 WHERE id=4; -- 阻塞
UPDATE students_nk_lock SET id=7 WHERE id=7; -- 阻塞
UPDATE students_nk_lock SET id=10 WHERE id=10; -- 阻塞

-- 不锁 ({10,100} -> +inf)
INSERT INTO students_nk_lock VALUES(11, 'Dave', 100); -- 非阻塞

ROLLBACK;
```

## 正例 3：普通索引+等值查询+记录存在

普通索引的等值查询，范围边界记录存在，使用左侧临键锁+右侧间隙锁。

> [!NOTE]
> 若是唯一索引+记录存在，则不会扩散，由默认临键锁进化为记录锁。也就是说，只要记录不存在或者是普通索引，都会导致锁范围扩散，
> 即最终使用间隙锁或临键锁或二者都有。

### 事务A

```
BEGIN;
SELECT * FROM students_nk_lock WHERE score = 90 FOR UPDATE;
```

预期锁住范围`({1, 85} -> {7, 95})`。

### 事务B

```
BEGIN;

-- 不锁 (-inf -> {1, 85}]
INSERT INTO students_nk_lock VALUES(-1, 'Dave', 84); -- 非阻塞
INSERT INTO students_nk_lock VALUES(0, 'Dave', 85); -- 非阻塞
UPDATE students_nk_lock SET score=85 WHERE score=85; -- 阻塞

-- 锁住 ({1, 85} -> {7, 95})
INSERT INTO students_nk_lock VALUES(2, 'Dave', 85); -- 阻塞，{2,85}会进入间隙
INSERT INTO students_nk_lock VALUES(2, 'Dave', 91); -- 阻塞
UPDATE students_nk_lock SET score=90 WHERE score=90; -- 阻塞

-- 不锁 [{7, 95} -> +inf)
INSERT INTO students_nk_lock VALUES(8, 'Dave', 95); -- 非阻塞
UPDATE students_nk_lock SET score=95 WHERE score=95; -- 阻塞
```

## 正例 4：普通索引+范围查询+命中边界记录

### 事务A

```
BEGIN;
SELECT * FROM students_nk_lock WHERE score >= 90 FOR UPDATE;
```

范围边界记录不存在，按照右扩散规则，预期锁住范围`({1, 85} -> +inf)`。

**命中边界记录**

范围的左边界记录刚好存在，对于普通索引，会进行左扩散，即从`{4, 90}`扩散到`{1, 85}`。

### 事务B

```
BEGIN;

-- 锁住 (-inf -> {1, 85}]
INSERT INTO students_nk_lock VALUES(0, 'Dave', 83); -- 阻塞
INSERT INTO students_nk_lock VALUES(2, 'Dave', 84); -- 阻塞
UPDATE students_nk_lock SET score=85 WHERE score=85; -- 阻塞

-- 锁住 [{1, 85} -> +inf)
INSERT INTO students_nk_lock VALUES(5, 'Dave', 85); -- 阻塞
INSERT INTO students_nk_lock VALUES(5, 'Dave', 91); -- 阻塞
INSERT INTO students_nk_lock VALUES(11, 'Dave', 101); -- 阻塞
```

为什么锁了全表？

## 正例 5：普通索引+范围查询：BETWEEN+命中边界记录

### 事务A

```
BEGIN;
SELECT * FROM students_nk_lock WHERE score between 85 and 90 FOR UPDATE;
```

### 事务B

```
BEGIN;

-- 不锁 (-inf -> {1, 85}]
INSERT INTO students_nk_lock VALUES(-1, 'Dave', 84); -- 非阻塞
INSERT INTO students_nk_lock VALUES(0, 'Dave', 85); -- 非阻塞
UPDATE students_nk_lock SET score=85 WHERE score=85; -- 阻塞

-- 锁住 ({1, 85} -> {7, 95})
INSERT INTO students_nk_lock VALUES(2, 'Dave', 85); -- 阻塞，{2,85}会进入间隙
INSERT INTO students_nk_lock VALUES(2, 'Dave', 91); -- 阻塞
UPDATE students_nk_lock SET score=90 WHERE score=90; -- 阻塞

-- 不锁 [{7, 95} -> +inf)
INSERT INTO students_nk_lock VALUES(8, 'Dave', 95); -- 非阻塞
UPDATE students_nk_lock SET score=95 WHERE score=95; -- 阻塞
```