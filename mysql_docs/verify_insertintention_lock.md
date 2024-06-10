## 验证插入意向锁

插入意向锁 (Insert Intention Lock) 是一种特殊的 InnoDB 行锁算法，旨在避免插入操作之间发生冲突，同时提高并发性和性能。

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

## 案例一

本例验证主键（`id`）索引上的插入意向锁。由于是唯一索引，所以插入加的排他锁是间隙锁，不是临键锁，不含右边界。

### 事务 A

```
BEGIN;

INSERT INTO students_lock VALUES(6, 'Alice', 6);
```

预期不会锁住主键索引上的`(5,10)`间隙的插入，但会阻塞这个间隙内的获得排他锁的操作。

### 事务 B

```
BEGIN;

-- 不锁相同间隙内的插入
INSERT INTO students_lock VALUES(7, 'Alice', 7); -- 非阻塞
INSERT INTO students_lock VALUES(8, 'Lucy', 8); -- 非阻塞

-- 锁住相同间隙内的其他需要排他锁的操作
SELECT * FROM students_lock WHERE id > 5 FOR UPDATE; -- 阻塞 
DELETE FROM students_lock WHERE id > 5; -- 阻塞

ROLLBACK;
```

## 案例二

本例在阻塞时观察事务等待状态。

### 事务操作

事务 A：

```
BEGIN;

DELETE FROM students_lock WHERE id > 5; 
```

事务 B：

```
BEGIN;

INSERT INTO students_lock VALUES(6, 'Alice', 6); -- 阻塞
```

### 观察等待状态

阻塞时，观察事务 B 的等待状态：

```
> SHOW ENGINE INNODB STATUS;

...
---TRANSACTION 504086, ACTIVE 12 sec inserting
mysql tables in use 1, locked 1
LOCK WAIT 2 lock struct(s), heap size 1136, 1 row lock(s)
MySQL thread id 2, OS thread handle 277810910976, query id 53 localhost root update
INSERT INTO students_lock VALUES(6, 'Alice', 6)
------- TRX HAS BEEN WAITING 12 SEC FOR THIS LOCK TO BE GRANTED:
RECORD LOCKS space id 36 page no 3 n bits 80 index PRIMARY of table `testdb`.`students_lock` trx id 504086 
lock_mode X locks gap before rec insert intention waiting
...
```

识别关键字`... rec insert intention waiting`。

> [!NOTE]
> 注意本例和案例一的区别，本例是后执行插入操作，这样才是插入意向锁被阻塞，才能观察到对应的关键字。