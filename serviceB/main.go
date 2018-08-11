package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
	"github.com/go-ini/ini"
	log "github.com/cihub/seelog"
	"github.com/gin-gonic/gin"
	"net/http"
	"strings"
	"os"
	"strconv"
	"os/signal"
)

func add(c *gin.Context) {
	a := c.Param("a")
	b := c.Param("b")

	x, err1 := strconv.Atoi(a)
	if err1 != nil {
		log.Errorf("add || a convert failed")
		c.Status(http.StatusBadRequest)
		return
	}
	y, err2 := strconv.Atoi(b)
	if err2 != nil {
		log.Errorf("add || b convert failed")
		c.Status(http.StatusBadRequest)
		return
	}
	sum := x + y
	c.String(200, "%d", sum)
}

func checkHandle(c *gin.Context) {
	c.Status(http.StatusOK)
	return
}

func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()
		c.Next()

		// after request
		latency := time.Since(t)
		// access the status we are sending
		status := c.Writer.Status()

		log.Debug("GinLogger URL %s, Method %s, perf(hs) %.02f, status %d",
			c.Request.URL.String(),
			c.Request.Method,
			latency.Seconds()*1000.0,
			status)
	}
}

func main() {
	//配置日志
	defer log.Flush()
	logger, err := log.LoggerFromConfigAsFile("webserviceLogCfg.xml")
	if err != nil {
		fmt.Println("log config not found,", err)
		os.Exit(0)
	}
	log.ReplaceLogger(logger) //将默认logger换成现在这个

	//从命令行得到端口号
	if len(os.Args) != 2 {
		log.Errorf("参数个数有误，应输入一个端口号")
		fmt.Println("参数个数有误，应输入一个端口号")
		return
	}
	port := os.Args[1]

	runtime.GOMAXPROCS(1)
	var wg sync.WaitGroup
	wg.Add(2)

	file, err := ini.Load("service.cfg")
	if err != nil {
		log.Errorf("无法解析配置文件#%s", err.Error())
		return
	}
	section := file.Section("")
	aIp := section.Key("httpAddrA").String()
	bIp := section.Key("httpAddrB").String()
	bIp = bIp + ":" + port

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(GinLogger())
	router.GET("/Add/:a/:b", add) //一个接口，给客户端调用
	router.GET("/aliveCheck", checkHandle)

	//设置网络配置
	srv := &http.Server{ //server是HTTP包下面的一个结构体
		Addr: bIp,
		Handler: router,
	}
	go func() {
		defer wg.Done()
		if err := srv.ListenAndServe(); err != nil {
			log.Critical("ListenAndServe1 quit, error ", err)
			fmt.Printf("ListenAndServe1 quit, %v\n", err)
		}
	}()

	go func() {
		defer wg.Done()
		strArray := "http://" + aIp + "/registerList?ip=" + bIp
		//给注册中心发送心跳
		for {
			http.Post(strArray, "", strings.NewReader(""))
			time.Sleep(5 * time.Second)
		}
	}()

	fmt.Println("Starting Server Successful")
	log.Warn("Starting Server Successful")

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit //blocking ....
	fmt.Println("Shutdown Server ...")
	log.Warn("Shutdown Server ...")
	wg.Wait()

}
