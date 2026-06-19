package mysqlx

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/tanenking/gsframe/gsinf"
	"github.com/tanenking/gsframe/internal/logger"

	"github.com/jmoiron/sqlx"
)

type mysqlClient_t struct {
	*gsinf.MysqlConfig
	mysqlDB *sqlx.DB
}

func makeMysql(cfg *gsinf.MysqlConfig) (cli *mysqlClient_t, err error) {
	dbDSN := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s", cfg.UserName, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.Charset)

	cli = &mysqlClient_t{
		MysqlConfig: cfg,
	}
	cli.mysqlDB, err = sqlx.Open("mysql", dbDSN)
	if err != nil {
		return
	}
	cli.mysqlDB.SetMaxOpenConns(mysql_max_open_conns)
	cli.mysqlDB.SetMaxIdleConns(mysql_max_idle_conns)
	cli.mysqlDB.SetConnMaxLifetime(mysql_max_lifttime)

	if err = cli.mysqlDB.Ping(); err != nil {
		return
	}

	logger.Log().Info("mysql conn success -> %s", dbDSN)
	return
}

func (m *mysqlClient_t) _update_or_delete(query string, args ...interface{}) (rowsAffected int64, err error) {
	if m.mysqlDB == nil {
		err = fmt.Errorf("mysqlDB %s not init", m.Name)
		logger.Log().Error("sql = %s, err = %v", query, err)
		return
	}
	ts := strings.TrimSpace(query)
	flag := strings.ToLower(ts[0:6])
	if strings.Compare(flag, "update") != 0 && strings.Compare(flag, "delete") != 0 {
		err = fmt.Errorf("this function just run update or delete")
		logger.Log().Error("sql = %s, err = %v", query, err)
		return
	}
	var ret sql.Result
	ret, err = m.mysqlDB.Exec(query, args...)
	if err != nil {
		logger.Log().Error("sql = %s, err = %v", query, err)
		return
	}
	rowsAffected, err = ret.RowsAffected()
	if err != nil {
		logger.Log().Error("sql = %s, err = %v", query, err)
		return
	}
	return
}

func (m *mysqlClient_t) DBName() string {
	return m.Name
}

func (m *mysqlClient_t) TableExists(table_name string) bool {
	var count int32 = 0
	sql := fmt.Sprintf(`select count(*) from information_schema.tables where TABLE_SCHEMA = '%s' and TABLE_NAME = '%s' limit 1`, m.Name, table_name)
	err, has := m.Get(&count, sql)
	if err != nil {
		logger.Log().Error("sql = %s, err = %v", sql, err)
		return false
	}
	if !has {
		return false
	}

	return count > 0
}

func (m *mysqlClient_t) Exec(sql string, args ...interface{}) (res sql.Result, err error) {
	if m.mysqlDB == nil {
		err = fmt.Errorf("mysqlDB %s not init", m.Name)
		logger.Log().Error("sql = %s, err = %v", sql, err)
		return
	}
	res, err = m.mysqlDB.Exec(sql, args...)
	if err != nil {
		logger.Log().Error("sql = %s, err = %v", sql, err)
	}
	return
}

func (m *mysqlClient_t) Get(data interface{}, query string, args ...interface{}) (err error, has bool) {
	if m.mysqlDB == nil {
		err = fmt.Errorf("mysqlDB %s not init", m.Name)
		logger.Log().Error("sql = %s, err = %v", query, err)
		return
	}
	err = m.mysqlDB.Get(data, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = nil
			has = false
			return
		}
		logger.Log().Error("sql = %s, err = %v", query, err)
		return
	}
	has = true
	return
}

func (m *mysqlClient_t) Select(arr interface{}, query string, args ...interface{}) (err error) {
	if m.mysqlDB == nil {
		err = fmt.Errorf("mysqlDB %s not init", m.Name)
		logger.Log().Error("sql = %s, err = %v", query, err)
		return
	}
	err = m.mysqlDB.Select(arr, query, args...)
	if err != nil {
		logger.Log().Error("sql = %s, err = %v", query, err)
		return
	}
	return
}

func (m *mysqlClient_t) Delete(query string, args ...interface{}) (rowsAffected int64, err error) {
	return m._update_or_delete(query, args...)
}
func (m *mysqlClient_t) Update(query string, args ...interface{}) (rowsAffected int64, err error) {
	return m._update_or_delete(query, args...)
}

func (m *mysqlClient_t) Insert(query string, args ...interface{}) (lastInsertID int64, err error) {
	if m.mysqlDB == nil {
		err = fmt.Errorf("mysqlDB %s not init", m.Name)
		logger.Log().Error("sql = %s, err = %v", query, err)
		return
	}
	ts := strings.TrimSpace(query)
	flag := strings.ToLower(ts[0:6])
	if strings.Compare(flag, "insert") != 0 {
		err = fmt.Errorf("this function just run insert")
		logger.Log().Error("sql = %s, err = %v", query, err)
		return
	}
	var ret sql.Result
	ret, err = m.mysqlDB.Exec(query, args...)
	if err != nil {
		logger.Log().Error("sql = %s, err = %v", query, err)
		return
	}
	lastInsertID, err = ret.LastInsertId()
	if err != nil {
		logger.Log().Error("sql = %s, err = %v", query, err)
		return
	}
	return
}

func (m *mysqlClient_t) Query(query string, args ...interface{}) (results []*gsinf.MysqlResult, err error) {
	if m.mysqlDB == nil {
		err = fmt.Errorf("mysqlDB %s not init", m.Name)
		logger.Log().Error("sql = %s, err = %v", query, err)
		return
	}
	results = []*gsinf.MysqlResult{}
	var rows *sql.Rows
	//查询数据，取所有字段
	rows, err = m.mysqlDB.Query(query, args...)
	if err != nil {
		logger.Log().Error("sql = %s, err = %v", query, err)
		return
	}

	defer rows.Close()

	for {
		res := &gsinf.MysqlResult{}
		//返回所有列
		res.Fields, err = rows.Columns()
		if err != nil {
			return
		}
		//这里表示一行所有列的值，用[]byte表示
		vals := make([][]byte, len(res.Fields))
		//这里表示一行填充数据
		scans := make([]interface{}, len(res.Fields))
		//这里scans引用vals，把数据填充到[]byte里
		for k := range vals {
			scans[k] = &vals[k]
		}
		i := 0
		for rows.Next() {
			//填充数据
			rows.Scan(scans...)
			//每行数据
			r := gsinf.MysqlRow{
				Values: map[string]string{},
			}
			//把vals中的数据复制到row中
			for k, v := range vals {
				key := res.Fields[k]
				//这里把[]byte数据转成string
				r.Values[key] = string(v)
			}
			//放入结果集
			res.Rows = append(res.Rows, r)
			i++
		}

		results = append(results, res)
		if !rows.NextResultSet() {
			break
		}
	}

	if err != nil {
		logger.Log().Error("sql = %s, err = %v", query, err)
	}
	return
}
