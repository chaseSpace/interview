## 验证Explain结果中的 Using index condition

## 理论

在MySQL中，Explain结果中Extra列显示`Using index condition`，
表示Server层将Where条件中的范围查询列下推到引擎层进行过滤，减少了引擎层返回的记录数。官方叫做**索引下推**。

无法使用索引下推的情况：

- Where 条件中包含非索引列
- Where 条件中对索引列使用了函数
- Where 条件中对索引列使用了范围查询，例如 `>`，`<`，`>=`，`<=`，`between`，`in`
- Where 条件中对索引列使用了OR查询
- Where 条件中对索引列使用了模糊查询，例如 `like`，`rlike`，`regexp`

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

TODO