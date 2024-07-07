## 验证 Explain 结果中的 Using index condition

## 理论

在 MySQL 中，Explain 结果中 Extra 列显示`Using index condition`，
表示 Server 层将 Where 语句中的部分条件下推到引擎层进行过滤，减少了引擎层回表次数以及返回给 Server 层的记录数，提高了查询性能。
官方名词叫做**索引下推**。

使用索引下推的条件（需同时满足）：

- 仅支持 InnoDB 和 MyISAM 引擎。
- 查询使用了二级索引（包含单列索引和联合索引）或联合唯一索引。
- 表连接 type 为`range`、 `ref_or_null`、`ref`, 以及`eq_ref`方式的查询，以下从效率最差到效率最好的方式排序。
    - `range`查询是指使用 = 、 <> 、 > 、 >= 、 < 、 <= 、 IS (NOT) NULL 、 <=> 、 BETWEEN 、 LIKE 或 IN()
      其中之一的运算符与**常量**比较（不含`=`）。
    - `ref_or_null` 是指 Where 语句中对索引列进行等值比较时，且额外包含为 NULL 值的过滤条件，
例如`WHERE key_column=expr OR key_column IS NULL`。
    - `ref`是指使用了非主键或唯一类型的索引，并且过滤行数占比较高。
    - `eq_ref`是指在多表查询中，对索引包含的所有列使用等值比较运算符，并且使用的是**前表**的主键或唯一非 NULL 索引。
- 其他少见的条件限制，见[官方文档][0]。

**简单记忆**

满足使用二级索引+对索引列进行范围查询/NULL 查询的情况。

[0]: https://dev.mysql.com/doc/refman/5.7/en/index-condition-pushdown-optimization.html

无法使用索引下推的情况：

- 使用了主键索引
- 使用了等值查询
- 其他等。

**官方文档**

在[MySQL 5.7 官方文档 Extra](https://dev.mysql.com/doc/refman/5.7/en/explain-output.html#explain-extra-information)
的页面中精确搜索`Using index condition`关键字。

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
    age        INT,
    department VARCHAR(50),
    salary     DECIMAL(10, 2),
    key idx_department_salary (department, salary),
    key idx_age_salary (age, salary),
    key idx_salary (salary)
);

INSERT INTO employees (name, age, department, salary) VALUES 
                                                ('Alice', 30, 'Sales', 50000.00),
                                                ('Bob', 35, 'HR', 60000.00),
                                                ('Charlie', 40, 'Engineering', 70000.00),
                                                ('David', 45, 'Marketing', 55000.00),
                                                ('Eve', 50, 'Finance', 65000.00);


CREATE TABLE employees2
(
    name       VARCHAR(50),
    age        INT,
    department VARCHAR(50),
    salary     DECIMAL(10, 2),
    primary key idx_name_dep (name, department)
);
```

为了避免数据量过少，影响 explain 结果，使用存储过程造数据：

```plain
DELIMITER //

CREATE PROCEDURE generate_test_data(
    IN num_rows INT
)
BEGIN
    DECLARE i INT DEFAULT 1;
    DECLARE department_list VARCHAR(255);
    DECLARE department VARCHAR(50);
    DECLARE age INT;
    DECLARE salary DECIMAL(10, 2);
    DECLARE salary_range DECIMAL(10, 2);
    
    -- 清空表数据
    TRUNCATE TABLE employees;

    -- 部门列表
    SET department_list = 'HR,IT,Finance,Sales,Marketing,Operations,Research';

    -- 循环插入数据
    WHILE i <= num_rows DO
        -- 随机选择部门
        SET department = SUBSTRING_INDEX(SUBSTRING_INDEX(department_list, ',', 1 + FLOOR(RAND() * (LENGTH(department_list) - LENGTH(REPLACE(department_list, ',', ''))) / LENGTH(','))), ',', -1);
        
        -- 随机生成年龄
        SET age = FLOOR(RAND() * 40) + 20;

        -- 随机生成工资范围
        SET salary_range = CASE
            WHEN age <= 25 THEN RAND() * 2000 + 3000
            WHEN age > 25 AND age <= 35 THEN RAND() * 3000 + 4000
            ELSE RAND() * 5000 + 5000
        END;

        -- 随机生成工资
        SET salary = ROUND(RAND() * salary_range, 2);
        
        INSERT INTO employees (name, age, department, salary)
        VALUES (
            CONCAT('Employee_', i),
            age,
            department,
            salary
        );
        
        SET i = i + 1;
    END WHILE;
    
END//

DELIMITER ;
```

表 2 的存储过程：

```plain
DELIMITER //

CREATE PROCEDURE generate_test_data2(
    IN num_rows INT
)
BEGIN
    DECLARE i INT DEFAULT 1;
    DECLARE department_list VARCHAR(255);
    DECLARE department VARCHAR(50);
    DECLARE age INT;
    DECLARE salary DECIMAL(10, 2);
    DECLARE salary_range DECIMAL(10, 2);
    
    -- 清空表数据
    TRUNCATE TABLE employees;

    -- 部门列表
    SET department_list = 'HR,IT,Finance,Sales,Marketing,Operations,Research';

    -- 循环插入数据
    WHILE i <= num_rows DO
        -- 随机选择部门
        SET department = SUBSTRING_INDEX(SUBSTRING_INDEX(department_list, ',', 1 + FLOOR(RAND() * (LENGTH(department_list) - LENGTH(REPLACE(department_list, ',', ''))) / LENGTH(','))), ',', -1);
        
        -- 随机生成年龄
        SET age = FLOOR(RAND() * 40) + 20;

        -- 随机生成工资范围
        SET salary_range = CASE
            WHEN age <= 25 THEN RAND() * 2000 + 3000
            WHEN age > 25 AND age <= 35 THEN RAND() * 3000 + 4000
            ELSE RAND() * 5000 + 5000
        END;

        -- 随机生成工资
        SET salary = ROUND(RAND() * salary_range, 2);
        
        INSERT INTO employees2 (name, age, department, salary)
        VALUES (
            CONCAT('Employee_', i),
            age,
            department,
            salary
        );
        
        SET i = i + 1;
    END WHILE;
    
END//

DELIMITER ;
```

生成数据：

```sql
call generate_test_data(100000);
call generate_test_data2(100000);
```

## 正例 1：使用联合索引+range 查询

第一种：同时包含等值查询和 range 查询。

```mysql
mysql>
EXPLAIN
SELECT *
FROM employees as t1
WHERE department = 'HR'
  and salary > 1;
+----+-------------+-------+------------+-------+-----------------------+-----------------------+---------+------+------+----------+-----------------------+
| id | select_type | table | partitions | type  | possible_keys         | key                   | key_len | ref  | rows | filtered | Extra                 |
+----+-------------+-------+------------+-------+-----------------------+-----------------------+---------+------+------+----------+-----------------------+
|  1 | SIMPLE      | t1    | NULL       | range | idx_department_salary | idx_department_salary | 59      | NULL |    2 |   100.00 | Using index condition |
+----+-------------+-------+------------+-------+-----------------------+-----------------------+---------+------+------+----------+-----------------------+
```

第二种：使用联合索引，仅包含范围查询。

```mysql
-- 要求索引过滤条数较多，即 filtered 列接近100，否则不使用索引下推。
mysql>
EXPLAIN
SELECT *
FROM employees as t1
WHERE age > 80;
+----+-------------+-------+------------+-------+----------------+----------------+---------+------+------+----------+-----------------------+
| id | select_type | table | partitions | type  | possible_keys  | key            | key_len | ref  | rows | filtered | Extra                 |
+----+-------------+-------+------------+-------+----------------+----------------+---------+------+------+----------+-----------------------+
|  1 | SIMPLE      | t1    | NULL       | range | idx_age_salary | idx_age_salary | 5       | NULL |    1 |   100.00 | Using index condition |
+----+-------------+-------+------------+-------+----------------+----------------+---------+------+------+----------+-----------------------+
```

> [!NOTE]
> range 查询还包括 Between 和 Like 这样的运算符，示例省略。。

第三种：使用联合唯一索引+range 查询

```mysql
mysql>
create unique index idx_name_dep on employees (name, department);
Query OK, 0 rows affected (0.66 sec)

mysql>
EXPLAIN
SELECT *
FROM employees as t1
WHERE name = 'Employee_11'
  and department > 'x';
+----+-------------+-------+------------+-------+------------------------------------+--------------+---------+------+------+----------+-----------------------+
| id | select_type | table | partitions | type  | possible_keys                      | key          | key_len | ref  | rows | filtered | Extra                 |
+----+-------------+-------+------------+-------+------------------------------------+--------------+---------+------+------+----------+-----------------------+
|  1 | SIMPLE      | t1    | NULL       | range | idx_name_dep,idx_department_salary | idx_name_dep | 106     | NULL |    1 |   100.00 | Using index condition |
+----+-------------+-------+------------+-------+------------------------------------+--------------+---------+------+------+----------+-----------------------+
```

## 正例 2：使用单列索引+range 查询

```mysql
mysql>
EXPLAIN
SELECT *
FROM employees as t1
WHERE salary > 10000;
+----+-------------+-------+------------+-------+---------------+------------+---------+------+------+----------+-----------------------+
| id | select_type | table | partitions | type  | possible_keys | key        | key_len | ref  | rows | filtered | Extra                 |
+----+-------------+-------+------------+-------+---------------+------------+---------+------+------+----------+-----------------------+
|  1 | SIMPLE      | t1    | NULL       | range | idx_salary    | idx_salary | 6       | NULL |    1 |   100.00 | Using index condition |
+----+-------------+-------+------------+-------+---------------+------------+---------+------+------+----------+-----------------------+
```

## 正例 3：使用单列索引+null 查询

`IS NULL`查询。

```mysql
mysql>
EXPLAIN
SELECT *
FROM employees as t1
WHERE salary is null;
+----+-------------+-------+------------+------+---------------+------------+---------+-------+------+----------+-----------------------+
| id | select_type | table | partitions | type | possible_keys | key        | key_len | ref   | rows | filtered | Extra                 |
+----+-------------+-------+------------+------+---------------+------------+---------+-------+------+----------+-----------------------+
|  1 | SIMPLE      | t1    | NULL       | ref  | idx_salary    | idx_salary | 6       | const |    1 |   100.00 | Using index condition |
+----+-------------+-------+------------+------+---------------+------------+---------+-------+------+----------+-----------------------+
```

`IS NOT NULL`查询。

```mysql
mysql>
EXPLAIN
SELECT *
FROM employees as t1
WHERE name is not null;
+----+-------------+-------+------------+-------+---------------+--------------+---------+------+------+----------+-----------------------+
| id | select_type | table | partitions | type  | possible_keys | key          | key_len | ref  | rows | filtered | Extra                 |
+----+-------------+-------+------------+-------+---------------+--------------+---------+------+------+----------+-----------------------+
|  1 | SIMPLE      | t1    | NULL       | range | idx_name_dep  | idx_name_dep | 53      | NULL |    1 |   100.00 | Using index condition |
+----+-------------+-------+------------+-------+---------------+--------------+---------+------+------+----------+-----------------------+
```

## 正例 4：使用联合索引+ref_or_null 查询

这里替换为单列索引也一样。`IS NULL`查询。

```mysql
mysql>
EXPLAIN
SELECT *
FROM employees as t1
WHERE age = 1000
   or age is null;
+----+-------------+-------+------------+-------------+----------------+----------------+---------+-------+------+----------+-----------------------+
| id | select_type | table | partitions | type        | possible_keys  | key            | key_len | ref   | rows | filtered | Extra                 |
+----+-------------+-------+------------+-------------+----------------+----------------+---------+-------+------+----------+-----------------------+
|  1 | SIMPLE      | t1    | NULL       | ref_or_null | idx_age_salary | idx_age_salary | 5       | const |    2 |   100.00 | Using index condition |
+----+-------------+-------+------------+-------------+----------------+----------------+---------+-------+------+----------+-----------------------+
```

`IS NOT NULL`查询。

```mysql
mysql>
EXPLAIN
SELECT *
FROM employees as t1
WHERE age = 1000
   or age is not null;
+----+-------------+-------+------------+-------+----------------+----------------+---------+------+------+----------+-----------------------+
| id | select_type | table | partitions | type  | possible_keys  | key            | key_len | ref  | rows | filtered | Extra                 |
+----+-------------+-------+------------+-------+----------------+----------------+---------+------+------+----------+-----------------------+
|  1 | SIMPLE      | t1    | NULL       | range | idx_age_salary | idx_age_salary | 5       | NULL |    1 |   100.00 | Using index condition |
+----+-------------+-------+------------+-------+----------------+----------------+---------+------+------+----------+-----------------------+
```

## 正例 5：使用联合索引+range 查询+非索引列查询

说明索引下推与是否包含非索引列条件没有关系。

```mysql
mysql>
EXPLAIN
SELECT *
FROM employees as t1
WHERE name < 'Employee'
  and salary > 11111;
+----+-------------+-------+------------+-------+---------------+--------------+---------+------+------+----------+------------------------------------+
| id | select_type | table | partitions | type  | possible_keys | key          | key_len | ref  | rows | filtered | Extra                              |
+----+-------------+-------+------------+-------+---------------+--------------+---------+------+------+----------+------------------------------------+
|  1 | SIMPLE      | t1    | NULL       | range | idx_name_dep  | idx_name_dep | 53      | NULL |    1 |    33.33 | Using index condition;
Using where |
+----+-------------+-------+------------+-------+---------------+--------------+---------+------+------+----------+------------------------------------+
```

## 反例 1：使用单列索引+等值查询

```mysql
mysql>
EXPLAIN
SELECT *
FROM employees as t1
WHERE salary = 10000;
+----+-------------+-------+------------+------+---------------+------------+---------+-------+------+----------+-------+
| id | select_type | table | partitions | type | possible_keys | key        | key_len | ref   | rows | filtered | Extra |
+----+-------------+-------+------------+------+---------------+------------+---------+-------+------+----------+-------+
|  1 | SIMPLE      | t1    | NULL       | ref  | idx_salary    | idx_salary | 6       | const |    1 |   100.00 | NULL  |
+----+-------------+-------+------------+------+---------------+------------+---------+-------+------+----------+-------+
```

## 反例 2：使用联合索引+等值查询

```mysql
mysql>
EXPLAIN
SELECT *
FROM employees as t1
WHERE age = 44
  and salary = 10000;
+----+-------------+-------+------------+------+---------------------------+----------------+---------+-------------+------+----------+-------+
| id | select_type | table | partitions | type | possible_keys             | key            | key_len | ref         | rows | filtered | Extra |
+----+-------------+-------+------------+------+---------------------------+----------------+---------+-------------+------+----------+-------+
|  1 | SIMPLE      | t1    | NULL       | ref  | idx_age_salary,idx_salary | idx_age_salary | 11      | const,const |    1 |   100.00 | NULL  |
+----+-------------+-------+------------+------+---------------------------+----------------+---------+-------------+------+----------+-------+
```

## 反例 3：使用主键索引+range 查询

第一种：使用单列主键索引。

```mysql
mysql>
EXPLAIN
SELECT *
FROM employees as t1
WHERE id > 90000;
+----+-------------+-------+------------+-------+---------------+---------+---------+------+-------+----------+-------------+
| id | select_type | table | partitions | type  | possible_keys | key     | key_len | ref  | rows  | filtered | Extra       |
+----+-------------+-------+------------+-------+---------------+---------+---------+------+-------+----------+-------------+
|  1 | SIMPLE      | t1    | NULL       | range | PRIMARY       | PRIMARY | 4       | NULL | 19096 |   100.00 | Using where |
+----+-------------+-------+------------+-------+---------------+---------+---------+------+-------+----------+-------------+
```

第二种：使用联合主键索引。

```mysql
mysql>
mysql>
EXPLAIN
SELECT *
FROM employees2 as t1
WHERE name = 'Employee_10014'
  and department >= 'IT';
+----+-------------+-------+------------+-------+---------------+---------+---------+------+------+----------+-------------+
| id | select_type | table | partitions | type  | possible_keys | key     | key_len | ref  | rows | filtered | Extra       |
+----+-------------+-------+------------+-------+---------------+---------+---------+------+------+----------+-------------+
|  1 | SIMPLE      | t1    | NULL       | range | PRIMARY       | PRIMARY | 104     | NULL |    1 |   100.00 | Using where |
+----+-------------+-------+------------+-------+---------------+---------+---------+------+------+----------+-------------+
```