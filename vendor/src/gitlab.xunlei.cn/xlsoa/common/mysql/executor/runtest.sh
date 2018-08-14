#!/bin/bash

export MYSQL_TEST_HOST=10.10.159.17
export MYSQL_TEST_PORT=3306
export MYSQL_TEST_USER=root
export MYSQL_TEST_PASSWORD="sd-9898w"
export MYSQL_TEST_DBNAME=test

go test -run Start -v
