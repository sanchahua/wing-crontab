package mysql

import (
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	"fmt"
)

var ErrNoRows = sql.ErrNoRows

type MysqlBase struct {
	db *sql.DB
	dataSourceName string
	maxOpenConns int;
	maxIdleConns int
}

func (m *MysqlBase)init(dataSourceName string, maxOpenConns, maxIdleConns int) error {
	m.dataSourceName = dataSourceName
	m.maxOpenConns = maxOpenConns
	m.maxIdleConns = maxIdleConns

	var err error
	if m.db, err = sql.Open("mysql", m.dataSourceName); err != nil {
		return err
	}

	if err = m.db.Ping(); err != nil {
		return err
	}

	m.db.SetMaxOpenConns(m.maxOpenConns)
	m.db.SetMaxIdleConns(m.maxIdleConns)
	return nil
}

func (m *MysqlBase)QueryAll(strSql string, args ...interface{}) ([]map[string]*string, error) {

	rows, err := m.db.Query(strSql, args ...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if len(cols) == 0 {
		return nil, fmt.Errorf("Columns 0")
	}

	table := make([]map[string]*string, 0)
	rawResult := make([][]byte, len(cols))
	dest := make([]interface{}, len(cols))
	for i, _ := range rawResult {
		dest[i] = &rawResult[i] // Put pointers to each string in the interface slice
	}

	for rows.Next() {
		if err = rows.Scan(dest...); err != nil {
			return nil, err
		}

		row := make(map[string]*string)
		for i, raw := range rawResult {
			key := cols[i]
			if raw == nil {
				row[key] = nil
			} else {
				s := string(raw)
				row[key] = &s
			}
		}
		table = append(table, row)
	}

	if len(table) == 0 {
		return nil, ErrNoRows
	}

	return table, nil
}

func (m *MysqlBase)QueryAll2Slice(strSql string, args ...interface{}) ([][]*string, error) {

	rows, err := m.db.Query(strSql, args ...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if len(cols) == 0 {
		return nil, fmt.Errorf("Columns 0")
	}

	table := make([][]*string, 0)
	rawResult := make([][]byte, len(cols))
	dest := make([]interface{}, len(cols))
	for i, _ := range rawResult {
		dest[i] = &rawResult[i] // Put pointers to each string in the interface slice
	}

	for rows.Next() {
		if err = rows.Scan(dest...); err != nil {
			return nil, err
		}

		row := make([]*string, 0)
		for _, raw := range rawResult {
			if raw == nil {
				row = append(row, nil)
			} else {
				s := string(raw)
				row = append(row, &s)
			}
		}
		table = append(table, row)
	}

	if len(table) == 0 {
		return nil, ErrNoRows
	}

	return table, nil
}

func (m *MysqlBase)QueryRow(strSql string, args []interface{}, dest ...interface{}) error {

	rows, err := m.db.Query(strSql, args ...)
	if err != nil {
		return err
	}
	defer rows.Close()
	if false == rows.Next() {
		return ErrNoRows
	}

	if err = rows.Scan(dest...); err != nil {
		return err
	}

	return nil
}

func (m *MysqlBase)QueryExist(strSql string, args ...interface{}) error {

	rows, err := m.db.Query(strSql, args ...)
	if err != nil {
		return err
	}
	defer rows.Close()
	if false == rows.Next() {
		return ErrNoRows
	}

	return nil
}

func (m *MysqlBase)QueryRow2Mapss(strSql string, args ...interface{}) (map[string]*string, error) {
	rows, err := m.db.Query(strSql, args ...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	if len(cols) == 0 {
		return nil, fmt.Errorf("Columns 0")
	}

	rawResult := make([][]byte, len(cols))
	dest := make([]interface{}, len(cols))
	for i, _ := range rawResult {
		dest[i] = &rawResult[i] // Put pointers to each string in the interface slice
	}

	if false == rows.Next() {
		return nil, ErrNoRows
	}

	if err = rows.Scan(dest...); err != nil {
		return nil, err
	}

	row := make(map[string]*string)
	for i, raw := range rawResult {
		key := cols[i]
		if raw == nil {
			row[key] = nil
		} else {
			s := string(raw)
			row[key] = &s
		}
	}
	return row, nil
}

func (m *MysqlBase)QueryRow2Slice(strSql string, args ...interface{}) ([]*string, error) {
	rows, err := m.db.Query(strSql, args ...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	if len(cols) == 0 {
		return nil, fmt.Errorf("Columns 0")
	}

	rawResult := make([][]byte, len(cols))
	dest := make([]interface{}, len(cols))
	for i, _ := range rawResult {
		dest[i] = &rawResult[i] // Put pointers to each string in the interface slice
	}
	if rows.Next() {
		if err = rows.Scan(dest...); err != nil {
			return nil, err
		}

		row := make([]*string, 0)
		for _, raw := range rawResult {
			if raw == nil {
				row = append(row, nil)
			} else {
				s := string(raw)
				row = append(row, &s)
			}
		}
		return row, nil
	}

	return nil, ErrNoRows
}

/// rowsAffected -1 表示 >= 0 | -2 表示 >= 1 | -3 表示 忽略 | 其他代表具体的数量
func (m *MysqlBase)Exec(rowsAffected int64, query string, args ...interface{}) error {
	if result, err := m.db.Exec(query, args ...); err != nil {
		return err
	} else if rows, err := result.RowsAffected(); err != nil {
		return err
	} else if rowsAffected >= 0 && rows == rowsAffected{
			return nil
	} else if rowsAffected == -1 && rows >= 0 {
		return nil
	} else if rowsAffected == -2 && rows >= 1 {
		return nil
	} else if rowsAffected == -3 {
		return nil
	} else {
		return fmt.Errorf("RowsAffected invalid, rowsAffected='%d' rows='%d'", rowsAffected, rows)
	}
}

func (m *MysqlBase)ExecInsert(query string, args ...interface{}) (int64, error) {
	if result, err := m.db.Exec(query, args ...); err != nil {
		return 0, err
	} else if rows, err := result.RowsAffected(); err != nil {
		return 0, err
	} else if rows != 1 {
		return 0, fmt.Errorf("rows invalid, rows='%d'", rows)
	} else if insertid, err := result.LastInsertId(); err != nil {
		return 0, err
	} else {
		return insertid, nil
	}
}