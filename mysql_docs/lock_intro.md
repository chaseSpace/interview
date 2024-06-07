# MySQL 锁机制详解

按锁类型分为共享锁、排它锁、意向锁。按锁粒度分为表级锁、行级锁。

## 基础

锁是 MySQL 用于实现事务的一个重要功能，用于并发事务中的同步。锁只能在事务中使用，事务结束后自动释放。

下文中的演示 SQL 省略了开启事务的语句。

## 全部锁

### 共享锁

共享锁允许事务并发读取相同行或表，但不允许并发读写，包括获取锁的事务也不能写数据。

#### 行级共享锁

用于 SELECT 语句，在尾部加上`LOCK IN SHARE MODE`即可。

```sql
-- 开启事务。。
SELECT *
FROM accounts
WHERE id = 1 LOCK IN SHARE MODE;

-- 结束事务，自动释放锁。
```

#### 表级共享锁

表级锁不同于行级锁，不属于事务管理，所以不是在事务结束后释放，而是与当前连接会话绑定，需要手动释放。

**使用场景**： 备份/恢复、批量更新：

```sql
-- LOCAL是MyISAM的特性，Innodb中不支持（仅语法支持）
-- 添加LOCAL修饰符表示锁仅影响当前会话，其他会话可以写入该表。
LOCK
TABLES accounts READ [LOCAL];
-- LOCK TABLES t1 READ, t2 READ; 可以锁定多张表

select *
from accounts;

-- 锁定后不能读取其他表
select *
from other;
-- ERROR 1100 (HY000) Table 'other' was not locked with LOCK TABLES

-- 释放
UNLOCK
tables;
```

> [!IMPORTANT]
> 执行`LOCK TABLES ...`时，会隐式地执行 COMMIT 语句，因此会结束任何正在进行的事务。
> 此外，还会释放当前会话持有的任何锁。会话断开也是一样的效果。

### 排它锁

独占锁，完全占有数据行，其他事务同一时刻不能读写该行。粒度较大，一般用于需要串行执行的事务场景，相当于分布式独占锁，同时是一种悲观锁。

#### 行级排它锁

用于 SELECT 语句，在尾部加上`FOR UPDATE`即可。

```sql
SELECT *
FROM accounts
WHERE id = 1 FOR UPDATE;
```

#### 表级排它锁

**使用场景**： 恢复、批量更新：

```sql
-- LOCAL是MyISAM的特性，Innodb中不支持（仅语法支持）
-- 添加LOCAL修饰符表示锁仅影响当前会话，其他会话可以写入该表。
LOCK
TABLES accounts WRITE [LOCAL];
-- LOCK TABLES t1 WRITE, t2 WRITE; 可以锁定多张表

-- 然后可以读写锁定的表

-- 锁定后不能读写*其他表
select *
from other;
-- ERROR 1100 (HY000) Table 'other' was not locked with LOCK TABLES
update f=2
from other
where f=1;
-- ERROR 1100 (HY000) Table 'other' was not locked with LOCK TABLES

-- 释放
UNLOCK
tables;
```

#### AUTO-INC 锁

在对包含自增列的表进行插入操作时，MySQL 会为该表加一个 AUTO-INC 锁，所以它是一种特殊的**表级锁**。
AUTO-INC 锁在插入语句完成（不是事务结束）后自动释放。在批量插入时存在性能问题。

从 5.1.22 版本开始，InnoDB 存储引擎提供了一种轻量级的锁来保护自增列的插入。轻量锁在为自增列分配值以后就释放了，不需要等到插入完成。

InnoDB 存储引擎提供了个 `innodb_autoinc_lock_mode` 的系统变量，是用来控制自增锁的模式。

- 当 `innodb_autoinc_lock_mode` = 0，就采用 AUTO-INC 锁，语句执行结束后才释放锁；
- 当 `innodb_autoinc_lock_mode` = 2，就采用轻量级锁，申请自增值后就释放锁；
- 当 `innodb_autoinc_lock_mode` = 1，默认锁模式。智能选择轻量级锁或 AUTO-INC 锁。
    - 简单插入使用轻量级锁。
    - 批量插入使用 AUTO-INC 锁。

关于简单插入和批量插入的定义：

- 简单插入：待插入行数已知的 SQL。包括没有嵌套子查询的单行和多行 INSERT 和 REPLACE
  语句，没有 `INSERT ... ON DUPLICATE KEY UPDATE`。
- 批量插入：待插入行数未知的 SQL。包括 `INSERT ... SELECT` 、 `REPLACE ... SELECT` 和 `LOAD DATA` 语句。

**轻量锁的问题**

它是性能最高的自增锁，但是当搭配 binlog 的日志格式是 `statement` 一起使用的时候，在「主从复制的场景」中会发生数据不一致的问题。
要解决这问题，binlog 日志格式要设置为 `row`。

### 意向锁

意向锁的意义的是为了解决行级锁的粒度太细，导致其他事务获取表级锁时检测行级锁效率太低的问题。
意向锁分为两种：意向共享锁、意向排它锁。分别对应了行级共享锁和行级排它锁。

**背景**

没有意向锁的时候。事务 A 对某一行加排他锁（X 锁）后，事务 B 需要获取此表的表级锁（共享或排他），
此时引擎需要逐行扫描才能检测到某些行的排他锁（然后判断是否冲突），效率太低。有了意向锁以后，事务 A 若需要获取行级排他锁，
则引擎自动为事务获取该表的一个意向排他锁（IX 锁），然后事务 B 想要获取表级锁，引擎只需要检测到存在表级的意向锁即可，
所以很快能检测到冲突，从而阻止事务 B 获取表级锁。

**简单记忆**

意向锁是引擎为事务自动加的，用以行级锁与表级锁之间的协同工作，是一种辅助锁。

#### 意向共享锁

意向共享锁（IS 锁，Intent Shared Lock），当一个事务打算在某些行上加共享锁时，引擎会自动**为事务**先在表上加一个意向共享锁。

例如， `SELECT ... LOCK IN SHARE MODE` 设置 IS 锁。

#### 意向排它锁

意向排他锁（IX 锁，Intent Exclusive Lock），当一个事务打算在某些行上加排他锁时，引擎会自动**为事务**先在表上加一个意向排他锁。

例如， `SELECT ... FOR UPDATE` 设置 IX 锁。

#### 验证

todo

#### 冲突性

意向锁之间全兼容，但与行级锁之间存在互斥关系：

| /       | 意向共享锁 (IS) | 意向排他锁 (IX) |
|---------|------------|------------|
| 共享锁 (S) | 兼容         | 互斥         |
| 排他锁 (X) | 互斥         | 互斥         |

#### 参考

- [掘金：详解 MySql InnoDB 中意向锁的作用](https://juejin.cn/post/6844903666332368909)
- [MySQL 官文：Intention Locks](https://dev.mysql.com/doc/refman/5.7/en/innodb-locking.html#innodb-intention-locks)

### 行级锁算法

行级锁（不论 S 还是 X 锁）这里有三种算法，即记录锁、间隙锁和临键锁，粒度顺序从小到大。在实际情况中由 MySQL 自动管理并使用。

#### 获取锁

获得行级锁，不论何种算法，都只有两种方式，对于读，是通过`for update`和`lock in share mode`获取（显式加锁）；
对于增删改是自动获取（隐式加锁）。

#### 记录锁（Record Lock）

记录锁是索引记录上的锁，也是粒度最小的锁，冲突概率最低，并发度最高。即使定义的表没有索引，InnoDB
创建一个隐藏的聚簇索引，并使用该索引进行记录锁定全部记录，相当于表锁但原理不同。

**支持**

RU 不支持，RC 以上支持。

**退化**

当尝试获取记录锁时，如果记录不存在，则自动退化为**间隙锁**。

#### 间隙锁（Gap Lock）

锁定一个范围，但是不包含记录本身，即**开区间**，不包括双端端点，如`(1,5)`。冲突概率大于记录锁。只存在于（RR）隔离级别，
目的是为了解决 RR 隔离级别下幻读的现象。此外，间隙锁只应用于非唯一索引列。唯一索引列（含主键）适用于记录锁。

间隙锁也包含 S 锁和 X 锁，由于相互兼容，所以不需要特别区分。间隙锁和行级排他锁互斥，它的唯一目的是防止幻读，即避免其他事务在间隙中插入数据，
造成 RR 及以上级别中一个事务多次读取时结果不一致的情况。

**支持**

RU,RC 不支持，RR 以上支持。

#### 临键锁（Next-Key Lock）

记录所 + 间隙锁的组合，锁定一个范围，并且锁定记录本身，即**左开右闭区间**，如`(1,5]`。是默认的行锁算法。

**支持**

RU,RC 不支持，RR 以上支持。

#### 插入意向锁（Insert Intention Lock）

TODO

#### 要点

- 临键锁是粒度最大的行级锁算法，也是默认的行级锁算法。
- 临键锁是间隙锁和记录锁的组合。
- 不管哪种行级锁算法，都应用在索引记录上，它可以是普通索引或主键索引。
- 唯一索引上的等值查询才能使用纯记录锁，只锁住条件匹配且确实存在的索引记录。

#### 实战验证

如果你想要了解行锁的具体工作细节，可以阅读笔者亲笔的文章 [验证行锁的三种算法](verify_rowlock.md)。

#### 参考

- [MySQL Innodb 锁][0]

## 关于死锁

## 默认原则

- 数据库的**增删改**操作默认都会加**排他锁**，而查询不会加任何锁。

[0]: https://dev.mysql.com/doc/refman/5.7/en/innodb-locking.html#innodb-record-locks