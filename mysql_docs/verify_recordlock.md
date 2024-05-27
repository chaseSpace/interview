# 验证记录锁

## 理论

对于行级锁（包含 S 和 X 锁），有三种应用算法，记录锁是其中一种。

记录锁是索引记录上的锁，也是粒度最小最容易理解的锁，冲突概率最低，并发度最高。即使定义的表没有索引，InnoDB
创建一个隐藏的聚簇索引，并使用该索引进行记录锁定全部记录（相当于表级锁了）。

**使用的时机**

- 主键或唯一索引+等值查询+命中记录。

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

-- 注意：必须为条件列建立非唯一索引，否则锁全表，下文会验证
CREATE TABLE students_rec_lock
(
    id    INT PRIMARY KEY,
    name  VARCHAR(50),
    score INT,
    key   idx_score(score)
);

INSERT INTO students_rec_lock (id, name, score)
VALUES (1, 'Alice', 85),
       (2, 'Bob', 90),
       (3, 'Carol', 95);
```

## 正例 1：行级 X 锁

唯一索引+等值匹配+记录存在。

### 执行事务 A

```plain
BEGIN;
SELECT * FROM students_rec_lock WHERE id = 1 FOR UPDATE;
```

### 查看 Innodb 的事务信息

启动一个新的会话查看 INNODB 的事务信息，事务信息经常用来定位死锁问题。

```plain
> SHOW ENGINE INNODB STATUS\G
...
------------
TRANSACTIONS
------------
Trx id counter 5389
Purge
done for trx's n:o < 5388 undo n:o < 0 state: running but idle
History list length 0
LIST OF TRANSACTIONS FOR EACH SESSION:
---TRANSACTION 421625954609888, not started
0 lock struct(s), heap size 1136, 0 row lock(s)
---TRANSACTION 5388, ACTIVE 1374 sec
2 lock struct(s), heap size 1136, 1 row lock(s)
MySQL thread id 2, OS thread handle 140150691600128, query id 50 localhost root
Trx read view will not see trx with id >= 5388, sees < 5388
...
```

其中`LIST OF TRANSACTIONS FOR EACH SESSION`
文字下方列出的是所有活跃的事务信息，从上到下按事务序号大到小（新到旧）排列。其中第一个事务`421625954609888`
，可看出序号非常大，`not started`表示是一个未开启事务的会话的事务序号（就是当前会话）。

事务 A 对应其中第二个事务：

```plain
---TRANSACTION 5388, ACTIVE 1374 sec
2 lock struct(s), heap size 1136, 1 row lock(s)
MySQL thread id 2, OS thread handle 140150691600128, query id 50 localhost root
Trx read view will not see trx with id >= 5388, sees < 5388
```

即事务 A 的 TX-id 为 5388，已开启 1374 秒（实际开启时间，笔者转了个身），持有 1 个行锁，会话 id 是 2。

> [!NOTE]
> 会话 id 的查看 SQL 是`SELECT CONNECT_ID()`。


另一种查看活跃事务的方式：

```plain
> SELECT * FROM information_schema.INNODB_TRX\G
*************************** 1. row ***************************
                    trx_id: 5388
                 trx_state: RUNNING
               trx_started: 2024-05-21 03:00:42
     trx_requested_lock_id: NULL
          trx_wait_started: NULL
                trx_weight: 2
       trx_mysql_thread_id: 2
                 trx_query: NULL
trx_operation_state: NULL
trx_tables_in_use: 0
trx_tables_locked: 1
trx_lock_structs: 2
trx_lock_memory_bytes: 1136
trx_rows_locked: 1
...
1 row in set (0.00 sec)
```

### 查看 Innodb 锁信息

```plain
mysql> SELECT * FROM INFORMATION_SCHEMA.INNODB_LOCKS;
Empty set, 1 warning (0.00 sec)
```

在没有发生锁等待（竞争）时，INNODB_LOCKS 表为空。

### 执行事务 B

启动一个新的会话执行事务 B。

```plain
BEGIN;
SELECT * FROM students_rec_lock WHERE id = 1 FOR UPDATE; -- 阻塞
UPDATE students_rec_lock SET score = 100 WHERE id = 1; -- 阻塞
```

**注意**：阻塞时，立即进入上一个查看锁的会话中去再次查看事务 A 的锁信息（锁等待超时默认 50s）。

### 再次查看 Innodb 的事务信息

```plain
> SHOW ENGINE INNODB STATUS\G
...
------------
TRANSACTIONS
------------
Trx id counter 5390
Purge done for trx's n:o < 5388 undo n:o < 0 state: running but idle
History list length 0
LIST OF TRANSACTIONS FOR EACH SESSION:
---TRANSACTION 421625954610800, not started
0 lock struct(s), heap size 1136, 0 row lock(s)
---TRANSACTION 5389, ACTIVE 534 sec starting index read
mysql tables in use 1, locked 1
LOCK WAIT 2 lock struct(s), heap size 1136, 4 row lock(s)
MySQL thread id 3, OS thread handle 140150691329792, query id 76 localhost root statistics
SELECT * FROM students_rec_lock WHERE id = 1 FOR UPDATE
------- TRX HAS BEEN WAITING 1 SEC FOR THIS LOCK TO BE GRANTED:
RECORD LOCKS space id 33 page no 3 n bits 72 index PRIMARY of table `testdb`.`students_rec_lock` trx id 5389 lock_mode X locks rec but not gap waiting
Record lock, heap no 2 PHYSICAL RECORD: n_fields 5; compact format; info bits 0
 0: len 4; hex 80000001; asc     ;;
 1: len 6; hex 000000001507; asc       ;;
 2: len 7; hex a70000011b0110; asc        ;;
 3: len 5; hex 416c696365; asc Alice;;
 4: len 4; hex 80000055; asc    U;;

------------------
---TRANSACTION 5388, ACTIVE 2484 sec
2 lock struct(s), heap size 1136, 1 row lock(s)
MySQL thread id 2, OS thread handle 140150691600128, query id 74 localhost root
Trx read view will not see trx with id >= 5388, sees < 5388
```

其中事务 5389 是事务 B，事务 5388 是事务 A。可见事务 B 正在等待（等了 1s，`WAITING 1 SEC`）一个**记录锁**
，具体是`students_rec_lock`表的主键上的记录锁，`lock_mode X locks rec but not gap`表示锁定记录本身但不锁定间隙（Gap），
跟着是锁定位置的物理信息描述（表空间 ID、页号和堆编号）。当事务 B 的锁等待超时后（50s），这个输出中将不会再显示事务 B 的锁信息，表示事务
B 没有在等待锁。

> [!TIP]
> 可见，只有在发生**锁等待**的时候才能观察到事务中已持有的行锁所应用的具体算法（记录锁/间隙锁/临键锁之一）。

### 再次查看 Innodb 锁信息

```sql
-- 查看发生竞争的锁列表明细
SELECT *
FROM INFORMATION_SCHEMA.INNODB_LOCKS;
+-------------+-------------+-----------+-----------+------------------------------+------------+------------+-----------+----------+-----------+
| lock_id     | lock_trx_id | lock_mode | lock_type | lock_table                   | lock_index | lock_space | lock_page | lock_rec | lock_data |
+-------------+-------------+-----------+-----------+------------------------------+------------+------------+-----------+----------+-----------+
| 5389:33:3:2 | 5389        | X         | RECORD    | `testdb`.`students_rec_lock` | PRIMARY    |         33 |         3 |        2 | 1         |
| 5388:33:3:2 | 5388        | X         | RECORD    | `testdb`.`students_rec_lock` | PRIMARY    |         33 |         3 |        2 | 1         |
+-------------+-------------+-----------+-----------+------------------------------+------------+------------+-----------+----------+-----------+

-- 在锁竞争条目较多时，可以通过下面语句来将竞争相同记录的锁相邻显示（较少使用）
SELECT DISTINCT t1.*
FROM INFORMATION_SCHEMA.INNODB_LOCKS t1
         left join INFORMATION_SCHEMA.INNODB_LOCKS t2
                   on t1.lock_space = t2.lock_space and t1.lock_page = t2.lock_page and t1.lock_rec = t2.lock_rec;
```

如表所示，发生锁等待（竞争）时，INNODB_LOCKS 表会同时列出参于锁竞争的所有事务。如果列表数目较多，
可以通过`Where lock_id like '%:33:3:2'`来过滤，当然你要先通过前面介绍的语句来获取发生死锁事务对应锁住的空间、页号和堆编号。

## 正例 2：行级 S 锁

TODO