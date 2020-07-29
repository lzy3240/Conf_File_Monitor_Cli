package conf

//AppConf 配置文件
type AppConf struct {
	Server `ini:"server"`
}

//Server 配置文件
type Server struct {
	Address string `ini:"address"`
	Time    int    `ini:"time"`
}
