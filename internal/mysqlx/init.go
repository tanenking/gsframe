package mysqlx

import (
	"time"

	"github.com/tanenking/gsframe/gsinf"
	"github.com/tanenking/gsframe/internal/logx"

	_ "github.com/go-sql-driver/mysql"
)

const (
	//最大连接数
	mysql_max_open_conns = 100
	//闲置连接数
	mysql_max_idle_conns = 20
	//最大连接周期
	mysql_max_lifttime = 60 * time.Second
)

type mysql_t struct {
	mysqls map[string]*mysqlClient_t
}

var (
	Mysql *mysql_t
)

func init() {
	Mysql = &mysql_t{
		mysqls: make(map[string]*mysqlClient_t),
	}
}

func InitMysqlHelper(configs map[string]*gsinf.MysqlConfig) error {
	if len(configs) <= 0 {
		return nil
	}

	for name, cfg := range configs {
		if len(cfg.Charset) <= 0 {
			cfg.Charset = "utf8mb4"
		}
		cfg.Name = name
		cli, err := makeMysql(cfg)
		if err != nil {
			logx.ErrorF("makeMysql err -> %v", err)
			return err
		}
		Mysql.mysqls[cfg.Name] = cli
		logx.InfoF("mysql helper [ %s ] create success", cfg.Name)
	}

	logx.InfoF("InitMysqlHelper success")
	return nil
}

func (r *mysql_t) DB(dbname string) gsinf.IMysqlClient {
	cli, ok := Mysql.mysqls[dbname]
	if !ok {
		return nil //&mysqlClient_t{MysqlConfig: &gsinf.MysqlConfig{Name: dbname}}
	}
	return cli
}
