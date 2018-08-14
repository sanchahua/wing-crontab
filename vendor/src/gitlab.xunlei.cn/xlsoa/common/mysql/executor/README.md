Mysql Executor
==========================


# Run test

You need to set environment for mysql connection.
If parameter is not set, default value will be used.

Example:
```
export MYSQL_TEST_HOST=localhost
```

Parameter List:
* ```MYSQL_TEST_HOST```     (Default 'localhost')
* ```MYSQL_TEST_PORT```     (Default '3306')
* ```MYSQL_TEST_USER```     (Default 'root')
* ```MYSQL_TEST_PASSWORD``` (Default '')
* ```MYSQL_TEST_DBNAME```   (Default 'test')


1. Set parameters in ```runtest.sh```
```
export MYSQL_TEST_HOST=localhost
export MYSQL_TEST_PORT=3306
export MYSQL_TEST_USER=root
export MYSQL_TEST_PASSWORD=""
export MYSQL_TEST_DBNAME=test
```

2. Run test
```./runtest.sh```


