package gsframe

import (
	"github.com/tanenking/gsframe/gsinf"
	"github.com/tanenking/gsframe/internal/mysqlx"
)

func InitMysqlHelper(configs map[string]*gsinf.MysqlConfig) error {
	return mysqlx.InitMysqlHelper(configs)
}
func Mysql(dbname string) gsinf.IMysqlClient {
	return mysqlx.Mysql.DB(dbname)
}
