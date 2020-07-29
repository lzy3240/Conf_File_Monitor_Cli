package conf

//AppConf 配置文件
type AppConf struct {
	Server  `ini:"server"`
	LogConf `ini:"log"`
}

//LogConf 配置文件
type LogConf struct {
	Level        string `ini:"level"`
	Florder      string `ini:"florder"`
	Perfix       string `ini:"perfix"`
	CutParameter string `ini:"cutparameter"`
}

//Server 配置文件
type Server struct {
	Address string `ini:"address"`
	Time    int    `ini:"time"`
}
