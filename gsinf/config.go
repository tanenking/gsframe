package gsinf

type ServiceConfig struct {
	ProjectName  string
	Etcd         *EtcdConfig
	Mysql        map[string]*MysqlConfig
	RedisCluster *RedisClusterConfig
}

type EtcdConfig struct {
	EtcdCenters []string
}
type MysqlConfig struct {
	Name     string
	UserName string
	Password string
	Host     string
	Port     uint16
	Database string
	Charset  string
}
type RedisClusterConfig struct {
	UserName  string
	Password  string
	KeyPrefix string
	Redis     []*RedisConfig
}
type RedisConfig struct {
	Host string
	Port uint16
}
