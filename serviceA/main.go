package main

import (
	"github.com/gin-gonic/gin"
	"time"
	log "github.com/cihub/seelog"
	"fmt"
	"os"
	"net/http"
	"sync"
	"runtime"
	"os/signal"
	"math/rand"
)

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

//存放注册服务
var registerService = []string{}

func main() {
	//设置日志
	defer log.Flush()
	logger, err := log.LoggerFromConfigAsFile("webserviceLogCfg.xml")
	if err != nil {
		fmt.Println("log config not found")
		os.Exit(0)
	}
	log.ReplaceLogger(logger) //将默认logger换成现在这个

	//设置并发线程
	runtime.GOMAXPROCS(1)
	var wg sync.WaitGroup
	wg.Add(2)

	//开启自己的网络路由
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(GinLogger())
	srv := &http.Server{ //server是HTTP包下面的一个结构体
		Addr: "127.0.0.1:6666",
		Handler: router,
	}

	go func() {
		defer wg.Done()
		if err := srv.ListenAndServe(); err != nil {
			log.Critical("ListenAndServe quit, error ", err)
			fmt.Printf("ListenAndServe quit, %v\n", err)
		}
	}()

	//检查服务是否还活着
	go func() {
		defer wg.Done()
		for {
			for k, ip := range registerService {
				_, err := http.Get("http://" + ip + "/aliveCheck")
				//说明该服务断开，则删除该IP
				if err != nil {
					kk := k + 1
					log.Errorf("%s 断开", ip)
					fmt.Println(ip, " 已经断开")
					registerService = append(registerService[:k], registerService[kk:]...)
				}
			}
			time.Sleep(5 * time.Second)
		}
	}()

	//接受B的注册请求
	router.POST("/registerList", registerHandle)
	//客户端问地址，给B的地址
	rand.Seed(time.Now().UnixNano()) //一个程序里面只需要seed一次
	router.GET("/serviceAddr", addrHandle)

	fmt.Println("Starting ServerA Successful")
	log.Warn("Starting ServerA Successful")

	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)
	<-quit //blocking ....
	fmt.Println("Shutdown Server ...")
	log.Warn("Shutdown Server ...")
	wg.Wait()

}

func registerHandle(c *gin.Context) {
	ip := c.Query("ip")

	//判断是否存在，存在则不做事情
	for _, sip := range registerService {
		if sip == ip {
			c.String(200, "%s registered", ip)
			return
		}
	}
	registerService = append(registerService, ip)
	c.String(200, "%s registered", ip)
}

func addrHandle(c *gin.Context) {
	queryParam := c.Query("queryParam")
	if queryParam == "add" {
		if len(registerService) <= 0 {
			log.Errorf("addrHandle || no service availble")
			c.String(http.StatusBadRequest, "当前没有可用服务")
			return
		}
		//随机抛一个ip
		i := rand.Intn(len(registerService))
		c.String(200, "%s", registerService[i])
	}
}
