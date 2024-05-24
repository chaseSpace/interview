# 验证间隙锁

## 理论

记录所 + 间隙锁的组合，锁定一个范围，并且锁定记录本身，即**左开右闭区间**，如`(1,5]`。是默认的粒度最大的行锁算法。

**使用的时机**

- 唯一索引+非等值查询时，使用临键锁；
- 普通索引。

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

## 正例 1：唯一索引+非等值查询

### 事务 A

```plain
BEGIN;
SELECT * FROM students_rec_lock WHERE id >= 3 FOR UPDATE;
```

预期锁住间隙`(3, +inf)`。

### 事务 B

```plain
BEGIN;

INSERT INTO students_rec_lock VALUES(4, 'Dave', 86); -- 阻塞
INSERT INTO students_rec_lock VALUES(5, 'Dave', 86); -- 阻塞
INSERT INTO students_rec_lock VALUES(1000, 'Dave', 86); -- 阻塞

UPDATE students_rec_lock SET score = 100 WHERE id = 3; -- 阻塞（锁边界）

-- 间隙内不存在的记录的查改不阻塞
SELECT * FROM students_rec_lock WHERE id = 4 FOR UPDATE; -- 非阻塞
UPDATE students_rec_lock SET score = 100 WHERE id = 4; -- 非阻塞

INSERT INTO students_rec_lock VALUES(0, 'Dave', 86); -- 成功，不影响(-inf, 1)

ROLLBACK;
```
