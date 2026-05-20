package gsinf

import "database/sql"

type MysqlRow struct {
	Values map[string]string
}
type MysqlResult struct {
	Fields []string
	Rows   []MysqlRow
}

type IMysql interface {
	DB(dbname string) IMysqlClient
}

type IMysqlClient interface {
	DBName() string
	TableExists(table_name string) bool
	Exec(sql string, args ...interface{}) (res sql.Result, err error)
	Get(data interface{}, query string, args ...interface{}) (err error, has bool) //查询单行
	Select(arr interface{}, query string, args ...interface{}) (err error)         //查询多行
	Delete(query string, args ...interface{}) (rowsAffected int64, err error)
	Update(query string, args ...interface{}) (rowsAffected int64, err error)
	Insert(query string, args ...interface{}) (lastInsertID int64, err error)
	Query(query string, args ...interface{}) (results []*MysqlResult, err error)
}
