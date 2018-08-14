// Handy executor for mysql executor.
// Based on database/sql.
//
//    // Init your database instance.
//    db := xxxx
//
//    // Create Executor with db
//    executor = executor.New(db)
//
//    // Multiple query
//    err := executor.Query("select * from test")
//    if err==nil {
//    	for executor.Next() {
//	    	executor.GetFieldInt32("id");
//	      	executor.GetFieldString("name");
//    	}
//    }
//	  executor.Clear()
//
//    // Single query
//    ok, err = executor.Find("select * from test where id=?", 1)
//    if err==nil && ok {
//    	 executor.GetFieldInt32("id")
//    	 executor.GetFieldString("name")
//	  }
//	  executor.Clear()
//
//    // insert
//    err = executor.Exec("insert into test set name=?", "xxx")
//    if err==nil {
//        executor.LastInsertId()
//    }
//	  executor.Clear()
//
//    // update
//    err = executor.Exec("update test set name=? where id=?", "yyy", 1)
//    if err==nil {
//        executor.AffectedRows()
//    }
//	  executor.Clear()
//

package executor

import (
	"database/sql"
	"strconv"
)

var Test = 0

// An executor is stateful, can't be used concurrently.
type Executor struct {
	db *sql.DB

	err  error
	rows *sql.Rows

	cols []string
	vals []interface{}

	columnIndex map[string]int

	lastid       int64
	affectedRows int64
}

func New(db *sql.DB) *Executor {
	return &Executor{
		db:          db,
		columnIndex: make(map[string]int),
	}
}

// Rows are closed here.
// Always Clear after query or exec.
func (r *Executor) Clear() {
	if r.rows != nil {
		r.rows.Close()
		r.rows = nil
	}
}

func (r *Executor) Query(query string, args ...interface{}) error {

	r.rows, r.err = r.db.Query(query, args...)
	if r.err != nil {
		return r.err
	}

	r.cols, _ = r.rows.Columns()
	r.vals = make([]interface{}, len(r.cols))
	for i, column := range r.cols {
		r.vals[i] = new(sql.RawBytes)
		r.columnIndex[column] = i
	}

	return nil
}

func (r *Executor) Find(query string, args ...interface{}) (bool, error) {

	err := r.Query(query, args...)
	if err != nil {
		return false, r.err
	}
	return r.Next(), nil
}

func (r *Executor) Exec(query string, args ...interface{}) error {

	var stmt *sql.Stmt
	var res sql.Result

	stmt, r.err = r.db.Prepare(query)
	if r.err != nil {
		return r.err
	}
	res, r.err = stmt.Exec(args...)
	defer stmt.Close()
	if r.err != nil {
		return r.err
	}
	r.lastid, _ = res.LastInsertId()
	r.affectedRows, _ = res.RowsAffected()
	return nil
}

func (r *Executor) LastInsertId() int64 {
	return r.lastid
}

func (r *Executor) AffectedRows() int64 {
	return r.affectedRows
}

func (r *Executor) Next() bool {
	if r.rows.Next() == false {
		return false
	}
	r.rows.Scan(r.vals...)
	return true
}

func (r *Executor) GetFieldString(name string) string {
	index, ok := r.columnIndex[name]
	if !ok {
		return ""
	}
	v := r.GetFieldByIndex(index)
	return v
}

func (r *Executor) GetFieldInt32(name string) int32 {
	index, ok := r.columnIndex[name]
	if !ok {
		return 0
	}
	v := r.GetFieldByIndex(index)
	n, err := strconv.ParseInt(v, 10, 32)
	if err != nil {
		return 0
	}
	return int32(n)
}
func (r *Executor) GetFieldInt64(name string) int64 {
	index, ok := r.columnIndex[name]
	if !ok {
		return 0
	}
	v := r.GetFieldByIndex(index)
	n, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0
	}
	return n
}
func (r *Executor) GetFieldUint32(name string) uint32 {
	index, ok := r.columnIndex[name]
	if !ok {
		return 0
	}
	v := r.GetFieldByIndex(index)
	n, err := strconv.ParseUint(v, 10, 32)
	if err != nil {
		return 0
	}
	return uint32(n)
}
func (r *Executor) GetFieldUint64(name string) uint64 {
	index, ok := r.columnIndex[name]
	if !ok {
		return 0
	}
	v := r.GetFieldByIndex(index)
	n, err := strconv.ParseUint(v, 10, 64)
	if err != nil {
		return 0
	}
	return n
}
func (r *Executor) GetFieldFloat(name string) float64 {
	index, ok := r.columnIndex[name]
	if !ok {
		return 0
	}
	v := r.GetFieldByIndex(index)
	n, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0
	}
	return n
}

func (r *Executor) GetFieldByIndex(i int) string {
	if i < 0 || i >= len(r.vals) {
		return ""
	}

	var f interface{}
	f = r.vals[i]
	v := string(*(f.(*sql.RawBytes)))
	return v
}
