package executor

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"os"
	"testing"
)

var (
	host     string
	port     string
	user     string
	password string
	dbname   string

	maxRecordsAdd = 3
)

var createTableSql = `CREATE TABLE IF NOT EXISTS test (
id int(11) unsigned NOT NULL AUTO_INCREMENT,
name varchar(64) NOT NULL DEFAULT '',
ts timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
PRIMARY KEY (id)
);`

var db *sql.DB

func env(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func TestStart(t *testing.T) {

	// Init
	{
		host = env("MYSQL_TEST_HOST", "localhost")
		port = env("MYSQL_TEST_PORT", "3306")
		user = env("MYSQL_TEST_USER", "root")
		password = env("MYSQL_TEST_PASSWORD", "")
		dbname = env("MYSQL_TEST_DBNAME", "test")

		var err error
		dns := fmt.Sprintf("%v:%v@tcp(%v:%v)/%v?charset=utf8", user, password, host, port, dbname)
		db, err = sql.Open("mysql", dns)
		if err != nil {
			t.Fatalf("sql.Open fail. [%v]", err)
		}
	}

	t.Run("CREATE", TestCreateTable)
	t.Run("ADD", TestAdd)
	t.Run("Query", TestQuery)
	t.Run("Find", TestFind)
	t.Run("Delete", TestDelete)
	t.Run("CheckDelete", TestCheckDelete)
	t.Run("DROP", TestDropTable)
}

func TestCreateTable(t *testing.T) {

	executor := New(db)
	defer executor.Clear()

	err := executor.Exec("CREATE DATABASE IF NOT EXISTS test")
	if err != nil {
		t.Fatalf("Create database fail. [%v]", err)
	}
	err = executor.Exec(createTableSql)
	if err != nil {
		t.Fatalf("Create table fail. [%v]", err)
	}
}

func TestDropTable(t *testing.T) {
	executor := New(db)
	defer executor.Clear()

	err := executor.Exec("DROP TABLE IF EXISTS test")
	if err != nil {
		t.Fatalf("Drop table fail. [%v]", err)
	}
}

func TestAdd(t *testing.T) {

	for i := 1; i <= maxRecordsAdd; i++ {

		var err error
		var name = fmt.Sprintf("name_%v", i)

		executor := New(db)
		defer executor.Clear()

		err = executor.Exec("insert into test set name=?", name)
		if err != nil {
			t.Fatalf("Insert row fail. [%v]", err)
			return
		}

		var lastId = executor.LastInsertId()
		if int(lastId) != i {
			t.Fatalf("LastInsertId %v!=%v", lastId, i)
		}

	}

}

func TestQuery(t *testing.T) {
	executor := New(db)
	defer executor.Clear()

	err := executor.Query("select id, name from test")
	if err != nil {
		t.Fatalf("Select fail. [%v]", err)
	}

	var nextId int32 = 0
	var trueName string
	for executor.Next() {
		nextId++
		trueName = fmt.Sprintf("name_%v", nextId)

		idGet := executor.GetFieldInt32("id")
		nameGet := executor.GetFieldString("name")

		if idGet != nextId {
			t.Fatalf("Select id '%v'!=True id '%v'", idGet, nextId)
		}
		if nameGet != trueName {
			t.Fatalf("Select name '%v'!=True name '%v'", nameGet, trueName)
		}

	}

	if int(nextId) != maxRecordsAdd {
		t.Fatalf("Query row number %v!=maxRecordsAdd %v", nextId, maxRecordsAdd)
	}
}

func TestFind(t *testing.T) {
	executor := New(db)
	defer executor.Clear()

	var findId = maxRecordsAdd
	ok, err := executor.Find("select name from test where id=?", findId)
	if err != nil {
		t.Fatalf("Select fail. [%v]", err)
	}
	if !ok {
		t.Fatalf("Id %v not found", findId)
	}
}

func TestDelete(t *testing.T) {

	executor := New(db)
	defer executor.Clear()

	var deleteId = maxRecordsAdd
	err := executor.Exec("delete from test where id=?", deleteId)
	if err != nil {
		t.Fatalf("Delete fail. [%v]", err)
	}

	if executor.AffectedRows() != 1 {
		t.Fatalf("Delete fail, affectd rows %v", executor.AffectedRows())
	}
}

func TestCheckDelete(t *testing.T) {
	executor := New(db)
	defer executor.Clear()

	var findId = maxRecordsAdd
	ok, err := executor.Find("select name from test where id=?", findId)
	if err != nil {
		t.Fatalf("Select fail. [%v]", err)
	}
	if ok {
		t.Fatalf("Id %v still exists, which should be Delete before", findId)
	}
}
