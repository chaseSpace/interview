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
CREATE DATABASE IF NOT EXISTS testdb;
USE testdb;

CREATE TABLE employees
(
    id         INT AUTO_INCREMENT PRIMARY KEY,
    name       VARCHAR(50),
    age        INT,
    department VARCHAR(50),
    salary     DECIMAL(10, 2),
    key idx_name (name),
    key idx_department_salary (department, salary)
);
```

插入测试数据100w：

```
DROP PROCEDURE IF EXISTS GenerateEmployees;
CREATE PROCEDURE GenerateEmployees(IN N INT)
BEGIN
    DECLARE i INT DEFAULT 1;
    DECLARE random_department VARCHAR(50);

    WHILE i <= N
        DO
            -- 随机生成部门
            SET random_department = ELT(FLOOR(RAND() * 5) + 1, 'HR', 'Finance', 'IT', 'Marketing', 'Sales');

            -- 插入数据
            INSERT INTO employees (name, age, department, salary)
            VALUES (substring(uuid(), 1, 8), rand(50) * 100, random_department, ROUND(RAND() * 1000 + 5000, 2)),
                   (substring(uuid(), 1, 8), rand(50) * 100, random_department, ROUND(RAND() * 1000 + 5000, 2)),
                   (substring(uuid(), 1, 8), rand(50) * 100, random_department, ROUND(RAND() * 1000 + 5000, 2)),
                   (substring(uuid(), 1, 8), rand(50) * 100, random_department, ROUND(RAND() * 1000 + 5000, 2)),
                   (substring(uuid(), 1, 8), rand(50) * 100, random_department, ROUND(RAND() * 1000 + 5000, 2)),
                   (substring(uuid(), 1, 8), rand(50) * 100, random_department, ROUND(RAND() * 1000 + 5000, 2)),
                   (substring(uuid(), 1, 8), rand(50) * 100, random_department, ROUND(RAND() * 1000 + 5000, 2)),
                   (substring(uuid(), 1, 8), rand(50) * 100, random_department, ROUND(RAND() * 1000 + 5000, 2)),
                   (substring(uuid(), 1, 8), rand(50) * 100, random_department, ROUND(RAND() * 1000 + 5000, 2)),
                   (substring(uuid(), 1, 8), rand(50) * 100, random_department, ROUND(RAND() * 1000 + 5000, 2));

            -- 计数器加 1
            SET i = i + 1;
        END WHILE;
END;

# TRUNCATE TABLE employees;
CALL GenerateEmployees(100000); # 运行需要1min+
```