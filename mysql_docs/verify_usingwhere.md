## 验证 Explain 结果中的 Using where

## 理论

在 MySQL 中，Explain 结果中 Extra 列显示`Using where`，表示 Server 层对引擎层返回的结果进行了二次过滤。
这通常是因为 WHERE 语句中包含了无法使用索引的列，或者<u>优化器决定不使用索引</u>。

Where 子句包含了无法使用索引的列的情况：

- 条件列没有被索引覆盖
- 条件列被索引覆盖，但不满足最左匹配规则
- 条件列被索引覆盖，但使用了函数
- 条件列被索引覆盖，但进行了隐式类型转换
- 条件列被索引覆盖，但使用的模糊查询规则不合理
- 条件列被索引覆盖，无特殊情况，但 SELECT 子句仅含常数
- 条件列被索引覆盖，但优化器决定不使用索引（效率更高）

**官方文档**

在[MySQL 5.7 官方文档 Extra](https://dev.mysql.com/doc/refman/5.7/en/explain-output.html#explain-extra-information)
的页面中精确搜索`Using where`关键字。

## 准备环境

参照[教程][0]使用 Docker 启动 MySQL 实例。

[0]: https://github.com/chaseSpace/go-common-pkg-exmaples/blob/master/_dockerfile/mysql/light.md


进入 mysql shell：

```plain
mysql> select version();
+-----------+
| version() |
+-----------+
| 5.7.44    |
+-----------+
```

执行下面的 sql 创建测试表：

```plain
CREATE DATABASE testdb;
USE testdb;

CREATE TABLE employees
(
    id         INT AUTO_INCREMENT PRIMARY KEY,
    name       VARCHAR(50),
    department VARCHAR(50),
    salary     DECIMAL(10, 2),
    key idx_department_salary (department, salary)
);

INSERT INTO employees (name, department, salary) VALUES
                                             ('Alice', 'HR', 5000.00),
                                             ('Bob', 'Engineering', 7000.00),
                                             ('Charlie', 'HR', 5500.00),
                                             ('David', 'Engineering', 7200.00),
                                             ('Eva', 'Marketing', 6000.00);

```

## 正例 1：没有被索引覆盖

第一种，全部条件列都没有被索引覆盖。

```mysql
mysql>
EXPLAIN
SELECT 1
FROM employees as t1
WHERE name = 'x';
+----+-------------+-------+------------+------+---------------+------+---------+------+------+----------+-------------+
| id | select_type | table | partitions | type | possible_keys | key  | key_len | ref  | rows | filtered | Extra       |
+----+-------------+-------+------------+------+---------------+------+---------+------+------+----------+-------------+
|  1 | SIMPLE      | t1    | NULL       | ALL  | NULL          | NULL | NULL    | NULL |    1 |   100.00 | Using where |
+----+-------------+-------+------------+------+---------------+------+---------+------+------+----------+-------------+
```

第二种，部分条件列没有被索引覆盖。

```mysql
mysql>
EXPLAIN
SELECT *
FROM employees as t1
WHERE department = 'x'
  and name = 'x';
+----+-------------+-------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+-------------+
| id | select_type | table | partitions | type | possible_keys         | key                   | key_len | ref   | rows | filtered | Extra       |
+----+-------------+-------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+-------------+
|  1 | SIMPLE      | t1    | NULL       | ref  | idx_department_salary | idx_department_salary | 53      | const |    1 |   100.00 | Using where |
+----+-------------+-------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+-------------+
```

## 正例 2：条件列被索引覆盖，但不满足最左匹配规则

```mysql
mysql>
EXPLAIN
SELECT *
FROM employees as t1
WHERE salary > 1;
+----+-------------+-------+------------+------+---------------+------+---------+------+------+----------+-------------+
| id | select_type | table | partitions | type | possible_keys | key  | key_len | ref  | rows | filtered | Extra       |
+----+-------------+-------+------------+------+---------------+------+---------+------+------+----------+-------------+
|  1 | SIMPLE      | t1    | NULL       | ALL  | NULL          | NULL | NULL    | NULL |    1 |   100.00 | Using where |
+----+-------------+-------+------------+------+---------------+------+---------+------+------+----------+-------------+
```

> [!NOTE]
> 题外话，这里有个特殊情况，将上述 SQL 中的 SELECT 部分改为`SELECT 1 ...`时，SQL 仍然会走索引，这是最左匹配规则的特例。

## 正例 3：条件列被索引覆盖，但使用了函数

```mysql
mysql>
EXPLAIN
SELECT *
FROM employees as t1
WHERE length(department) > 1;
+----+-------------+-------+------------+------+---------------+------+---------+------+------+----------+-------------+
| id | select_type | table | partitions | type | possible_keys | key  | key_len | ref  | rows | filtered | Extra       |
+----+-------------+-------+------------+------+---------------+------+---------+------+------+----------+-------------+
|  1 | SIMPLE      | t1    | NULL       | ALL  | NULL          | NULL | NULL    | NULL |    1 |   100.00 | Using where |
+----+-------------+-------+------------+------+---------------+------+---------+------+------+----------+-------------+
1 row in set
, 1 warning (0.01 sec)
```

## 正例 4：条件列被索引覆盖，但进行了隐式类型转换

```mysql
mysql>
EXPLAIN
SELECT *
FROM employees as t1
WHERE department = 1;
+----+-------------+-------+------------+------+-----------------------+------+---------+------+------+----------+-------------+
| id | select_type | table | partitions | type | possible_keys         | key  | key_len | ref  | rows | filtered | Extra       |
+----+-------------+-------+------------+------+-----------------------+------+---------+------+------+----------+-------------+
|  1 | SIMPLE      | t1    | NULL       | ALL  | idx_department_salary | NULL | NULL    | NULL |    1 |   100.00 | Using where |
+----+-------------+-------+------------+------+-----------------------+------+---------+------+------+----------+-------------+
```

## 正例 5：条件列被索引覆盖，但使用的模糊查询规则不合理

```mysql
mysql>
EXPLAIN
SELECT *
FROM employees as t1
WHERE department like '%xx';
+----+-------------+-------+------------+------+---------------+------+---------+------+------+----------+-------------+
| id | select_type | table | partitions | type | possible_keys | key  | key_len | ref  | rows | filtered | Extra       |
+----+-------------+-------+------------+------+---------------+------+---------+------+------+----------+-------------+
|  1 | SIMPLE      | t1    | NULL       | ALL  | NULL          | NULL | NULL    | NULL |    1 |   100.00 | Using where |
+----+-------------+-------+------------+------+---------------+------+---------+------+------+----------+-------------+
```

## 正例 6：条件列被索引覆盖，无特殊情况，但 SELECT 子句仅含常数

比如当部分覆盖的索引列是范围查询时，本应执行索引下推，但由于 SELECT 仅含常数，影响了下推策略执行，转而在 Server 层执行了范围过滤。

> [!NOTE]
> 通过理论推演，这里进行索引下推仍然是能够提高效率的，引擎层只要返回满足条件的行数，由 Server 层填充对应数量的行即可。
> 但经过测试，在 8.0 版本中也是这个分析结果。此外，笔者使用过十万条数据进行测试，结果无二，所以此处留下一个疑问。

```mysql
mysql>
EXPLAIN
SELECT 1
FROM employees as t1
WHERE department = 'HR'
  and salary > 1;
+----+-------------+-------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+--------------------------+
| id | select_type | table | partitions | type | possible_keys         | key                   | key_len | ref   | rows | filtered | Extra                    |
+----+-------------+-------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+--------------------------+
|  1 | SIMPLE      | t1    | NULL       | ref  | idx_department_salary | idx_department_salary | 53      | const |    1 |   100.00 | Using where;
Using index |
+----+-------------+-------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+--------------------------+
```

## 正例 7：条件列被索引覆盖，但优化器决定不使用索引

通常出现在 Where 子句条件能通过索引过滤的行数太少（体现为`filtered`列分值较低），不如走全表扫描。此时引擎层会直接返回全表给
Server 层，
后者再使用 Where 子句过滤。

```mysql
mysql>
EXPLAIN
SELECT *
FROM employees as t1
WHERE department > '';
+----+-------------+-------+------------+------+-----------------------+------+---------+------+-------+----------+-------------+
| id | select_type | table | partitions | type | possible_keys         | key  | key_len | ref  | rows  | filtered | Extra       |
+----+-------------+-------+------------+------+-----------------------+------+---------+------+-------+----------+-------------+
|  1 | SIMPLE      | t1    | NULL       | ALL  | idx_department_salary | NULL | NULL    | NULL | 97465 |    50.00 | Using where |
+----+-------------+-------+------------+------+-----------------------+------+---------+------+-------+----------+-------------+
```

## 正例 X：其他优化器不使用索引的情况

比如：

- 数据分布不均匀：某个值在表中占据了大部分行，MySQL 可能会认为使用索引不划算，选择进行全表扫描。
- 数据量很小：不如全表扫描快。

## 反例 1：所有条件列都被索引覆盖且无特殊情况

注意，是所有条件列，不是所有涉及的列。

```mysql
mysql>
EXPLAIN
SELECT *
FROM employees as t1
WHERE department = 'x'
  and salary > 1;
+----+-------------+-------+------------+-------+-----------------------+-----------------------+---------+------+------+----------+-----------------------+
| id | select_type | table | partitions | type  | possible_keys         | key                   | key_len | ref  | rows | filtered | Extra                 |
+----+-------------+-------+------------+-------+-----------------------+-----------------------+---------+------+------+----------+-----------------------+
|  1 | SIMPLE      | t1    | NULL       | range | idx_department_salary | idx_department_salary | 59      | NULL |    1 |   100.00 | Using index condition |
+----+-------------+-------+------------+-------+-----------------------+-----------------------+---------+------+------+----------+-----------------------+
```

## 反例 2：Where 子句空

第一种，SELECT 常数。

```mysql
mysql>
EXPLAIN
SELECT 1
FROM employees as t1;
+----+-------------+-------+------------+-------+---------------+-----------------------+---------+------+------+----------+-------------+
| id | select_type | table | partitions | type  | possible_keys | key                   | key_len | ref  | rows | filtered | Extra       |
+----+-------------+-------+------------+-------+---------------+-----------------------+---------+------+------+----------+-------------+
|  1 | SIMPLE      | t1    | NULL       | index | NULL          | idx_department_salary | 59      | NULL |    5 |   100.00 | Using index |
+----+-------------+-------+------------+-------+---------------+-----------------------+---------+------+------+----------+-------------+
```

第二种，SELECT 任意列。

```mysql
mysql>
EXPLAIN
SELECT *
FROM employees as t1;
+----+-------------+-------+------------+------+---------------+------+---------+------+------+----------+-------+
| id | select_type | table | partitions | type | possible_keys | key  | key_len | ref  | rows | filtered | Extra |
+----+-------------+-------+------------+------+---------------+------+---------+------+------+----------+-------+
|  1 | SIMPLE      | t1    | NULL       | ALL  | NULL          | NULL | NULL    | NULL |    5 |   100.00 | NULL  |
+----+-------------+-------+------------+------+---------------+------+---------+------+------+----------+-------+
```