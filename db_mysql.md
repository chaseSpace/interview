# MySQL 相关

本文内容部分摘自网络。

## 什么是关系型数据库

一种建立在关系模型的基础上的数据库。关系模型表明了数据库中所存储的数据之间的联系（一对一、一对多、多对多）。
大部分关系型数据库都使用 SQL 来操作数据库中的数据。并且，大部分关系型数据库都支持事务的四大特性(ACID)。

### 常见的关系型数据库

MySQL、PostgreSQL、Oracle、SQL Server、SQLite（微信本地的聊天记录的存储就是用的 SQLite）。

## MySQL 字段类型

MySQL 字段类型可以简单分为三大类：

- 数值类型：整型（TINYINT、SMALLINT、MEDIUMINT、INT 和 BIGINT）浮点型（FLOAT 和 DOUBLE）、定点型（DECIMAL）
- 字符串类型：CHAR、VARCHAR、TINYTEXT、TEXT、MEDIUMTEXT、LONGTEXT、TINYBLOB、BLOB、MEDIUMBLOB 和 LONGBLOB 等，最常用的是 CHAR 和
  VARCHAR。
- 日期时间类型：YEAR、TIME、DATE、DATETIME 和 TIMESTAMP 等。

### DECIMAL 和 FLOAT/DOUBLE 的区别是什么

DECIMAL 是定点数，FLOAT/DOUBLE 是浮点数。DECIMAL 可以存储精确的小数值，FLOAT/DOUBLE 只能存储近似的小数值。

### DATETIME 和 TIMESTAMP 的区别是什么

DATETIME 类型没有时区信息，TIMESTAMP 和时区有关。

TIMESTAMP 只需要使用 4 个字节的存储空间，但是 DATETIME 需要耗费 8 个字节的存储空间。但是，这样造成了一个问题，Timestamp
表示的时间范围更小。

- DATETIME：`1000-01-01 00:00:00` ~ `9999-12-31 23:59:59`
- Timestamp：`1970-01-01 00:00:01` ~ `2037-12-31 23:59:59`

### 为什么不建议使用 NULL 作为列默认值

- NULL与空字符串不同，NULL 需要更多的存储空间。
- 查询NULL值需要使用专门的SQL语句，比如 `IS NULL`和`IS NOT NULL`，而查询空字符串只需要使用 `=` 或者 `<>` 即可。
- NULL会影响聚合函数的查询结果，例如，SUM、AVG、MIN、MAX 等聚合函数会忽略 NULL 值。
    - COUNT(*) 会包含NULL值所在的行，但 COUNT(col) 不会。
- 查询不便：在查询中使用`NOT IN`或`!=`等反向条件时，查询结果不会包含NULL值所在的行，需要加上`ISNULL(col)`。

注意：DISTINCT 会将多个 NULL 值算作一个 NULL。

> 对于不需要做聚合的字段，可以允许NULL值。

## SQL语句的执行过程

连接器： 身份认证和权限相关(登录 MySQL 的时候)。
查询缓存： 执行查询语句的时候，会先查询缓存（MySQL 8.0 版本后移除，因为这个功能不太实用）。
分析器： 没有命中缓存的话，SQL 语句就会经过分析器，进行词法分析和语法分析，生成语法树；同时检查对应的表和字段是否存在。
优化器： 负责将语法树转化成执行计划。包括选择不同的索引、决定是否使用子查询或连接操作。
执行器： 由存储引擎执行语句，然后返回数据。

## MySQL架构分层是怎样的

TODO

## 为什么不建议单表超过2000w数据

TODO

## 参考

- https://javaguide.cn/database/mysql/mysql-questions-01.html
- 