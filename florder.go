package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"
)

//获取文件夹配置
func getFlorderconf() {
	apiURL := fmt.Sprintf("http://%s/florderconf", cfg.Server.Address)
	resp, err := http.Get(apiURL)
	if err != nil {
		log.Error("get failed, err:%v\n", err)
		return
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("get resp failed, err:%v\n", err)
		return
	}
	//fmt.Println(string(b))
	//分解为对象
	hc := hostFlorderconf{}
	err = json.Unmarshal(b, &hc)
	if err != nil {
		log.Error("err:%v\n", err)
	}
	tmp := len(hc.Florderconf)
	fconfs = fconfs[0:0] //清空切片中的历史配置
	if tmp > 0 {
		for _, v := range strings.Split(hc.Florderconf, "|") {
			flobj := new(florderConf)
			err = json.Unmarshal([]byte(v), &flobj)
			fconfs = append(fconfs, flobj)
		}
		log.Info("get florder conf success,%s", hc.Florderconf)
	} else {
		log.Info("get florder conf null")
	}
}

//调用接口发送结果数据
func postFlorderresult(data string) {
	apiURL := fmt.Sprintf("http://%s/florderresult", cfg.Server.Address)
	contentType := "application/json"
	resp, err := http.Post(apiURL, contentType, strings.NewReader(data))
	if err != nil {
		log.Error("post failed, err:%v\n", err)
		return
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("get resp failed, err:%v\n", err)
		return
	}
	log.Info("send florder result success,%v", string(b))
}

//列文件夹数量
func listdir(fl, fg, st string) (int, error) {
	var num int
	switch fg {
	case "florder":
		num = decideType(fl, st)
	case "dateflorder":
		dateStr := time.Now().Format("2006-01-02")
		tmpfl := fl + dateStr + "/"
		num = decideType(tmpfl, st)
	default:
		err := errors.New("unknown flag")
		return num, err
	}
	return num, nil
}

//过滤指定配置类型
func decideType(fl, st string) int {
	var files []string
	dir, err := ioutil.ReadDir(fl)
	checkErr("path error:", err)

	for _, fi := range dir {
		if fi.IsDir() {
			//如果是文件夹，直接记录该文件夹
			files = append(files, fi.Name())
		} else {
			//不是文件夹的，判断指定格式
			for _, k := range strings.Split(st, ",") {
				tmp := strings.Split(fi.Name(), ".")
				ok := isSliceIn(k, tmp)
				if ok {
					files = append(files, fi.Name())
				}
			}
		}
	}
	return len(files)
}

//判断string是否在slice内
func isSliceIn(t string, tr []string) bool {
	for _, v := range tr {
		if t == v {
			return true
		}
	}
	return false
}

//主监听
func watchflorder() {
	if len(fconfs) > 0 {
		for _, v := range fconfs {
			num, err := listdir(v.Florder, v.Flag, v.Stage)
			if err != nil {
				log.Error("list dir faild,err:%v", err)
			}
			var fres florderResult
			fres.Florder = v.Florder
			fres.Result = strconv.Itoa(num)
			data, err := json.Marshal(fres)
			if err != nil {
				log.Error("marshal faild,err:%v", err)
			}
			postFlorderresult(string(data))
			log.Info("florder %v has item:%v", v.Florder, num)
		}
	} else {
		log.Info("conf not found")
	}
}
