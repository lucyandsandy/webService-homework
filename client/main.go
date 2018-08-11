package main

import (
	"fmt"
	"os"
	"github.com/go-ini/ini"
	"net/http"
	"io/ioutil"
	log "github.com/cihub/seelog"
)

func main() {

	//配置日志信息
	defer log.Flush()
	logger, err := log.LoggerFromConfigAsFile("webserviceLogCfg.xml")
	if err != nil {
		fmt.Println("log config not found")
		os.Exit(0)
	}
	log.ReplaceLogger(logger) //将默认logger换成现在这个

	//从配置文件中读取A的地址
	file, err := ini.Load("service.cfg")
	if err != nil {
		log.Errorf("无法解析配置文件#%s", err.Error())
		return
	}
	section := file.Section("")
	serAIp := section.Key("httpAddrA").String()

	//读取参数
	if len(os.Args) != 3 {
		fmt.Println("参数输入不正确")
		log.Errorf("main || wrong input parameters")
		return
	}
	a := os.Args[1]
	b := os.Args[2]

	serviceIp, err := queryFunc(serAIp)
	if err != nil {
		log.Infof("queryFunc ||", err)
		fmt.Println("queryFunc || get B address failed ", err)
		return
	}
	if serviceIp == "no service availble" {
		log.Infof(" no service availble||", err)
		fmt.Println("no service availble|| get B address failed ", err)
		return
	}

	//路由到B中加法运算，将参数传过去
	strGetSum := "http://" + serviceIp + "/Add/" + a + "/" + b
	resp3, err3 := http.Get(strGetSum)

	//如果拿到的地址不可用，在重新找A要，直到拿到有用的
	for ; err3 != nil; {
		log.Errorf("main || getSum failed ", err3)
		fmt.Println("getSum error:", err3)
		log.Infof("main || 重新寻找可用服务... ")
		fmt.Println("重新寻找可用服务...")
		serviceIp_, err_ := queryFunc(serAIp)
		if err_ != nil {
			log.Infof("queryFunc ||", err_)
			fmt.Println("queryFunc || get B address failed ", err_)
			return
		}
		if serviceIp_ == "no service availble" {
			log.Infof(" no service availble||", err)
			fmt.Println("no service availble|| get B address failed ", err)
			return
		}
		strGetSum_ := "http://" + serviceIp_ + "/Add/" + a + "/" + b
		resp3, err3 = http.Get(strGetSum_)
	}
	defer resp3.Body.Close()
	body3, err2 := ioutil.ReadAll(resp3.Body)
	if err2 != nil {
		log.Errorf("main || read sum body  failed %s", err2)
		fmt.Println("readSum error:", err2)
		return
	}
	sum := string(body3)
	log.Infof("get sum successfully")
	fmt.Println("the sum is:", sum)
}

func queryFunc(serAIp string) (serviceIp string, err error) {
	//给A发送请求，得到B的地址
	strQuery := "http://" + serAIp + "/serviceAddr?queryParam=add"
	resp, err := http.Get(strQuery)
	if err != nil {
		log.Errorf("main || get method failed %s", err)
		fmt.Println("向A发送请求失败:", err)
		return "", err
	}
	defer resp.Body.Close()
	body, err1 := ioutil.ReadAll(resp.Body)
	if err1 != nil {
		log.Errorf("main || read B addr body  failed %s", err1)
		fmt.Println("读取A返回数据失败:", err1)
		return "", err1
	}
	serviceBIp := string(body)
	if resp.StatusCode != 200 {
		log.Infof("get B address failed", serviceBIp)
		fmt.Println("get B address failed || ", serviceBIp)
		return "no service availble", nil
	}
	log.Infof("get B address successfully")
	fmt.Println("B address is:", serviceBIp) //test
	return serviceBIp, nil
}
