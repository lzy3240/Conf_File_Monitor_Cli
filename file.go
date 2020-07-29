package main

import (
	"bufio"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

//获取文件配置
func getFileconf() {
	apiURL := fmt.Sprintf("http://%s/fileconf", cfg.Server.Address)
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
	hc := hostFileconf{}
	err = json.Unmarshal(b, &hc)
	if err != nil {
		log.Error("err:%v\n", err)
	}
	tmp := len(hc.Fileconf)
	confs = confs[0:0] //清空切片中的历史配置
	if tmp > 0 {
		for _, v := range strings.Split(hc.Fileconf, "|") {
			flobj := new(fileConf)
			err = json.Unmarshal([]byte(v), &flobj)
			confs = append(confs, flobj)
		}
		log.Info("get file conf success,%s", hc.Fileconf)
	} else {
		log.Info("get file conf null")
	}
}

//调用接口，发送结果
func postFileresult(data string) {
	apiURL := fmt.Sprintf("http://%s/fileresult", cfg.Server.Address)
	contentType := "application/json"
	resp, err := http.Post(apiURL, contentType, strings.NewReader(data))
	if err != nil {
		log.Error("post failed, err:%v\n", err)
		return
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error("get resp failed, err:%v", err)
		return
	}
	log.Info("send file result success,%v", string(b))
}

//复制文件
func copyFile(srcFileName, dstFileName string) (written int64, err error) {
	srcFile, err := os.Open(srcFileName)

	if err != nil {
		log.Error("open src file err: %v", err)
		return
	}

	defer srcFile.Close()

	//通过srcFile，获取到Reader
	reader := bufio.NewReader(srcFile)

	//打开dstFileName
	dstFile, err := os.OpenFile(dstFileName, os.O_WRONLY|os.O_CREATE, 0666)
	if err != nil {
		log.Error("open dst file err = %v", err)
		return
	}

	writer := bufio.NewWriter(dstFile)
	defer func() {
		writer.Flush() //把缓冲区的内容写入到文件
		dstFile.Close()

	}()

	return io.Copy(writer, reader)
}

//判断string是否在slice内
func isIn(t string, tr []*fileConf) bool {
	for _, v := range tr {
		if t == v.Filename {
			return true
		}
	}
	return false
}

//判断文件夹是否存在
func pathExists(path string) bool {
	_, err := os.Stat(path) //os.Stat获取文件信息
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}

//备份配置文件
func bakFileConf() {
	exist := pathExists("./tmp/")
	if !exist {
		err := os.Mkdir("./tmp/", os.ModePerm)
		checkErr("create dir faild,err:", err)
	}
	if len(confs) > 0 {
		//初始化待监控文件map
		for _, v := range confs {
			_, ok := tmppath[v.Filename]
			if !ok {
				t := strings.Split(v.Filename, "/")
				dstFileName := fmt.Sprintf("./tmp/%s%s", t[len(t)-1], time.Now().Format("20060102150405"))
				_, err := copyFile(v.Filename, dstFileName)
				if err != nil {
					log.Error("backup conf file faild,err:%v", err)
				} else {
					tmppath[v.Filename] = dstFileName
					log.Info("backup file success,backfilename:%v", dstFileName)
				}
			}
		}
		//去除不监控文件map
		for k := range tmppath {
			bool := isIn(k, confs)
			if !bool {
				delete(tmppath, k)
				log.Info("remove monitor conf success,filename:%v", k)
			}
		}
	} else {
		log.Info("confs not found")
	}
}

//获取文件MD5
func getMD5SumString(f *os.File) (string, error) {
	file1Sum := md5.New()
	_, err := io.Copy(file1Sum, f)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%X", file1Sum.Sum(nil)), nil
}

//主监听
func watchfile() {
	if len(confs) > 0 {
		for _, v := range confs {
			file1, err := os.Open(v.Filename)
			checkErr("open src file faild", err)
			defer file1.Close()
			file2, err := os.Open(tmppath[v.Filename])
			checkErr("open bak file faild", err)
			defer file2.Close()
			sum1, err := getMD5SumString(file1)
			checkErr("get src md5 faild", err)
			sum2, err := getMD5SumString(file2)
			checkErr("get bak md5 faild", err)

			if sum1 != sum2 {
				//比较文件不同
				file1.Seek(0, 0)
				file2.Seek(0, 0)
				sc1 := bufio.NewScanner(file1)
				sc2 := bufio.NewScanner(file2)
				for {
					sc1Bool := sc1.Scan()
					sc2Bool := sc2.Scan()
					if !sc1Bool && !sc2Bool {
						break
					}
					if sc1.Text() != sc2.Text() {
						var fres fileResult
						fres.Filename = v.Filename
						fres.Flag = v.Flag
						fres.Result = fmt.Sprintf("new string:%v , old string:%v", sc1.Text(), sc2.Text())
						data, err := json.Marshal(fres)
						if err != nil {
							log.Error("marshal faild,err:%v", err)
						}
						postFileresult(string(data))
						log.Info("file %v was changed,new string:%v , old string:%v", v.Filename, sc1.Text(), sc2.Text())
					}
				}

				//再备份
				t := strings.Split(v.Filename, "/")
				dstFileName := fmt.Sprintf("./tmp/%s%s", t[len(t)-1], time.Now().Format("20060102150405"))
				_, err := copyFile(v.Filename, dstFileName)
				if err != nil {
					log.Error("backup conf file faild,err:%v", err)
				} else {
					tmppath[v.Filename] = dstFileName
					log.Info("backup conf file success,bakfilename:%v", dstFileName)
				}
			} else {
				log.Info("conf file %v no change", v.Filename)
			}
		}
	}
}
