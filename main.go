package main

import (
	"fmt"
	"strings"
	"time"
	"zproject/Conf_File_Monitor_Cli/conf"

	"github.com/lzy3240/mlog"
	"github.com/lzy3240/mtools"
	"gopkg.in/ini.v1"
)

type hostFileconf struct {
	Fileconf string `json:"fileconf"`
}

type hostFlorderconf struct {
	Florderconf string `json:"florderconf"`
}

// type hostResult struct {
// 	Res string `json:"res"`
// }

type fileConf struct {
	Filename string `json:"filename"`
	Flag     string `json:"flag"`
	Stage    string `json:"stage"`
}

type fileResult struct {
	Filename string `json:"filename"`
	Flag     string `json:"flag"`
	Result   string `json:"result"`
}

type florderConf struct {
	Florder string `json:"florder"`
	Flag    string `json:"flag"`
	Stage   string `json:"stage"`
}

type florderResult struct {
	Florder string `json:"florder"`
	Result  string `json:"result"`
}

var (
	cfg     = new(conf.AppConf)
	log     *mlog.Logger
	confs   []*fileConf
	fconfs  []*florderConf
	ip      string
	tmppath = make(map[string]string)
)

//msg
func msg() {
	fmt.Println("*************************")
	fmt.Println(" 文件&文件夹变化监听服务")
	fmt.Println(" Author : Lzy")
	fmt.Println(" Version :1.0.0")
	fmt.Println("*************************")
	fmt.Println("client:", ip)
	fmt.Println()
}

//checkErr 检查错误
func checkErr(str string, err error) {
	if err != nil {
		log.Error("%v,%v", str, err)
	}
}

func init() {
	//1.1加载配置文件
	err := ini.MapTo(cfg, "./conf/config.ini")
	if err != nil {
		fmt.Printf("init config faild,err:%v", err)
		return
	}

	//1.2初始化日志
	tmp := strings.Split(cfg.LogConf.CutParameter, "|")
	s := strings.Join(tmp, "=")
	log = mlog.Newlog(cfg.LogConf.Level, cfg.LogConf.Florder, cfg.LogConf.Perfix, s)

	//1.3获取本机IP
	ip, err = mtools.GetOutboundIP()
	if err != nil {
		log.Error("get outbound ip faild,err:%v", err)
	} else {
		log.Info("get outbound ip success:%s", ip)
	}
}

func main() {
	msg()
	go func() {
		for {
			getFileconf()
			bakFileConf()
			watchfile()
			time.Sleep(time.Second * time.Duration(cfg.Server.Time))
		}
	}()
	go func() {
		for {
			getFlorderconf()
			watchflorder()
			time.Sleep(time.Second * time.Duration(cfg.Server.Time))
		}
	}()

	select {}
}
