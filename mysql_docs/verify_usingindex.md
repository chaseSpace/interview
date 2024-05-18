## 验证Explain结果中的 Using index

## 理论

在MySQL中，Explain结果中Extra列显示`Using index`，表示查询SQL涉及到的所有列都可以使用索引，**涉及**
指的是查询字段、排序字段、分组字段、连接字段，都要使用索引，缺一不可！

> [!IMPORTANT]
> **所有列都可以使用索引**不是指所有列都建立了索引，而是所有列都被SQL选择的索引所覆盖。

## 准备环境

参照[教程][0]使用Docker启动MySQL实例。

[0]: https://github.com/chaseSpace/go-common-pkg-exmaples/blob/master/_dockerfile/mysql/light.md


进入mysql shell：

```
mysql> select version();
+-----------+
| version() |
+-----------+
| 5.7.44    |
+-----------+
```

执行下面的sql创建测试表：

```
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

## 正例1：常用关键字仅含索引列

SQL中涉及的所有字段都使用索引时，Extra列会显示`Using index`。

```mysql
mysql>
EXPLAIN
SELECT department, count(1)
FROM employees
WHERE department = 'Engineering'
  and salary = 7000
group by department
order by department;
+----+-------------+-----------+------------+------+-----------------------+-----------------------+---------+-------------+------+----------+-------------+
| id | select_type | table     | partitions | type | possible_keys         | key                   | key_len | ref         | rows | filtered | Extra       |
+----+-------------+-----------+------------+------+-----------------------+-----------------------+---------+-------------+------+----------+-------------+
|  1 | SIMPLE      | employees | NULL       | ref  | idx_department_salary | idx_department_salary | 59      | const,const |    1 |   100.00 | Using index |
+----+-------------+-----------+------------+------+-----------------------+-----------------------+---------+-------------+------+----------+-------------+
```

## 正例2：JOIN 仅含索引列

```mysql
mysql>
EXPLAIN
SELECT 1
FROM employees as t1
         JOIN employees t2 ON t1.department = t2.department
WHERE t1.department = 'Engineering';
+----+-------------+-------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+-------------+
| id | select_type | table | partitions | type | possible_keys         | key                   | key_len | ref   | rows | filtered | Extra       |
+----+-------------+-------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+-------------+
|  1 | SIMPLE      | t1    | NULL       | ref  | idx_department_salary | idx_department_salary | 53      | const |    1 |   100.00 | Using index |
|  1 | SIMPLE      | t2    | NULL       | ref  | idx_department_salary | idx_department_salary | 53      | const |    1 |   100.00 | Using index |
+----+-------------+-------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+-------------+
```

## 反例1：SELECT 包含非索引列

### SELECT 仅含非索引列

下面的sql中，select部分选择了非索引列`name`，所以`Using index`字段没有显示。但这不影响使用索引来筛选数据，只是要进行回表查询
name 字段。

```mysql
mysql>
EXPLAIN
SELECT name
FROM employees
WHERE department = 'Engineering';
+----+-------------+-----------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+-------+
| id | select_type | table     | partitions | type | possible_keys         | key                   | key_len | ref   | rows | filtered | Extra |
+----+-------------+-----------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+-------+
|  1 | SIMPLE      | employees | NULL       | ref  | idx_department_salary | idx_department_salary | 53      | const |    1 |   100.00 | NULL  |
+----+-------------+-----------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+-------+
```

### SELECT 包含索引列和非索引列

下面的sql中，select部分包含了索引列和非索引列`name`，所以`Using index`字段没有显示。

```mysql
mysql>
EXPLAIN
SELECT department, name
FROM employees
WHERE department = 'Engineering';
+----+-------------+-----------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+-------+
| id | select_type | table     | partitions | type | possible_keys         | key                   | key_len | ref   | rows | filtered | Extra |
+----+-------------+-----------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+-------+
|  1 | SIMPLE      | employees | NULL       | ref  | idx_department_salary | idx_department_salary | 53      | const |    1 |   100.00 | NULL  |
+----+-------------+-----------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+-------+
```

## 反例2：ORDER_BY 包含非索引列

### ORDER_BY 仅含非索引列

SQL中的OrderBy语句中只包含非索引列，所以`Using index`字段没有显示。所以需要使用`filesort`，二者是互斥的。

```mysql
mysql>
EXPLAIN
SELECT 1
FROM employees
WHERE department = 'Engineering'
order by name;
+----+-------------+-----------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+---------------------------------------+
| id | select_type | table     | partitions | type | possible_keys         | key                   | key_len | ref   | rows | filtered | Extra                                 |
+----+-------------+-----------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+---------------------------------------+
|  1 | SIMPLE      | employees | NULL       | ref  | idx_department_salary | idx_department_salary | 53      | const |    1 |   100.00 | Using index condition;
Using filesort |
+----+-------------+-----------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+---------------------------------------+
```

### ORDER_BY 包含索引列和非索引列

```mysql
mysql>
EXPLAIN
SELECT 1
FROM employees
WHERE department = 'Engineering'
order by salary, name;
+----+-------------+-----------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+---------------------------------------+
| id | select_type | table     | partitions | type | possible_keys         | key                   | key_len | ref   | rows | filtered | Extra                                 |
+----+-------------+-----------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+---------------------------------------+
|  1 | SIMPLE      | employees | NULL       | ref  | idx_department_salary | idx_department_salary | 53      | const |    1 |   100.00 | Using index condition;
Using filesort |
+----+-------------+-----------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+---------------------------------------+
```

## 反例3：GROUP_BY 包含非索引列

### GROUP_BY 仅含非索引列

```mysql
mysql>
EXPLAIN
SELECT 1
FROM employees
WHERE department = 'Engineering'
group by name;
+----+-------------+-----------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+--------------------------------------------------------+
| id | select_type | table     | partitions | type | possible_keys         | key                   | key_len | ref   | rows | filtered | Extra                                                  |
+----+-------------+-----------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+--------------------------------------------------------+
|  1 | SIMPLE      | employees | NULL       | ref  | idx_department_salary | idx_department_salary | 53      | const |    1 |   100.00 | Using index condition;
Using temporary;
Using filesort |
+----+-------------+-----------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+--------------------------------------------------------+
```

### GROUP_BY 包含索引列和非索引列

```mysql
mysql>
EXPLAIN
SELECT 1
FROM employees
WHERE department = 'Engineering'
group by department, name;
+----+-------------+-----------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+--------------------------------------------------------+
| id | select_type | table     | partitions | type | possible_keys         | key                   | key_len | ref   | rows | filtered | Extra                                                  |
+----+-------------+-----------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+--------------------------------------------------------+
|  1 | SIMPLE      | employees | NULL       | ref  | idx_department_salary | idx_department_salary | 53      | const |    1 |   100.00 | Using index condition;
Using temporary;
Using filesort |
+----+-------------+-----------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+--------------------------------------------------------+
```

## 反例4：DISTINCT 包含非索引列

### DISTINCT 仅含非索引列

```mysql
mysql>
EXPLAIN
SELECT DISTINCT name
FROM employees
WHERE department = 'Engineering';
+----+-------------+-----------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+----------------------------------------+
| id | select_type | table     | partitions | type | possible_keys         | key                   | key_len | ref   | rows | filtered | Extra                                  |
+----+-------------+-----------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+----------------------------------------+
|  1 | SIMPLE      | employees | NULL       | ref  | idx_department_salary | idx_department_salary | 53      | const |    1 |   100.00 | Using index condition;
Using temporary |
+----+-------------+-----------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+----------------------------------------+
```

### DISTINCT 包含索引列和非索引列

```mysql
mysql>
EXPLAIN
SELECT DISTINCT department, name
FROM employees
WHERE department = 'Engineering';
+----+-------------+-----------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+----------------------------------------+
| id | select_type | table     | partitions | type | possible_keys         | key                   | key_len | ref   | rows | filtered | Extra                                  |
+----+-------------+-----------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+----------------------------------------+
|  1 | SIMPLE      | employees | NULL       | ref  | idx_department_salary | idx_department_salary | 53      | const |    1 |   100.00 | Using index condition;
Using temporary |
+----+-------------+-----------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+----------------------------------------+
```

## 反例5：HAVING 包含非索引列

由于`HAVING`语句是用于筛选分组后的`SELECT`中的数据，语法上不能单独引用无关的列，所以此处不涉及。

## 反例6：JOIN 包含非索引列

### JOIN 仅含非索引列

下面是一个自连接SQL，关注第一行的`Extra`列即可。

```mysql
mysql>
EXPLAIN
SELECT 1
FROM employees as t1
         JOIN employees t2 ON t1.name = t2.name
WHERE t1.department = 'Engineering';
+----+-------------+-------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+----------------------------------------------------+
| id | select_type | table | partitions | type | possible_keys         | key                   | key_len | ref   | rows | filtered | Extra                                              |
+----+-------------+-------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+----------------------------------------------------+
|  1 | SIMPLE      | t1    | NULL       | ref  | idx_department_salary | idx_department_salary | 53      | const |    1 |   100.00 | NULL                                               |
|  1 | SIMPLE      | t2    | NULL       | ALL  | NULL                  | NULL                  | NULL    | NULL  |    1 |   100.00 | Using where;
Using join buffer (Block Nested Loop) |
+----+-------------+-------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+----------------------------------------------------+
```

### JOIN 包含索引列和非索引列

```mysql
mysql>
EXPLAIN
SELECT 1
FROM employees as t1
         JOIN employees t2 ON t1.department = t2.department and t1.name = t2.name
WHERE t1.department = 'Engineer
+----+-------------+-------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+-------------+
| id | select_type | table | partitions | type | possible_keys         | key                   | key_len | ref   | rows | filtered | Extra       |
+----+-------------+-------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+-------------+
|  1 | SIMPLE      | t1    | NULL       | ref  | idx_department_salary | idx_department_salary | 53      | const |    1 |   100.00 | NULL        |
|  1 | SIMPLE      | t2    | NULL       | ref  | idx_department_salary | idx_department_salary | 53      | const |    1 |   100.00 | Using where |
+----+-------------+-------+------------+------+-----------------------+-----------------------+---------+-------+------+----------+-------------+
```
