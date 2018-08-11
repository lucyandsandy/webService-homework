#服务注册与拉取

**服务注册与拉取时序图**

```sequence
服务B->服务A:每隔5秒，发送心跳注册服务，POST(ip)
服务A-->服务B:返回200
服务A->服务B:每隔5秒，检查是否还有心跳，GET()
服务B-->服务A:返回200
客户端->服务A:获取服务B的地址，GET(add)
服务A-->客户端:返回 服务B的地址
客户端->服务B:请求加法运算服务，GET(a,b)
服务B-->客户端:返回运算结果sum
```

## 演示结果图示
**`服务B开启5个节点，分别设置端口号为1111，2222，3333，4444，5555`**

![1](./image/服务B.png)

**`开启服务A`**

![2](./image/服务A.png)

**`开启客户端，输入加法运算的参数，可以看出5个节点随机接受客户端的访问请求`**

![3](./image/客户端.png)

**`杀掉服务B端口号为4444，5555的进程，只剩下1111,2222,3333三个端口开启，由下图可以看出只有这3个端口被访问`**

![捕获](./image/负载均衡.PNG)

**`由此可以得出结论，该服务能够满足作业要求的两个节点负载均衡，一个挂掉，另一个节点仍能满足客户端请求`**

##注册中心A服务
**接口列表**

|接口|说明|
|----|-----|
|/registerList|接受服务B的注册|
|/serviceAddr|接受客户端对地址的查询|

###接受服务B注册

**Request**

- URL

	```http
	/registerList
	```
- Method

	`POST`
	
- 传递参数

	|参数|说明|
	|---|----|
	|ip|服务B的ip地址，如127.0.0.1：1111|
	
**检查服务是否掉线**

​	开启一个go routine，每隔5秒检查服务列表中服务是否断开，发送一个GET请求，如果返回状态码不是200，则服务判定断开，将服务从服务列表中删除
	

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


###接受客户端请求

**Request**

- URL

	```http
	/serviceAddr
	```
- Method

	`GET`
	
- 传递参数

	|参数|说明|
	|---|----|
	|queryParam|请求的服务类型，如add|
	
**负载均衡--随机算法**

​	从服务列表中随机的挑选一个可用服务给客户端，实现负载均衡


	import "math/rand"
	...
	rand.Seed(time.Now().UnixNano()) //一个程序里面只需要seed一次
	i := rand.Intn(len(registerService))
	c.String(200, "%s", registerService[i])

##功能服务B
**接口列表**

|接口|说明|
|----|-----|
|/Add|接受客户端的调用|
|/aliveCheck|接受服务A的心跳检查|

###接受客户端调用服务

**Request**

- URL

	```http
	/Add
	```
- Method

	`GET`
	
- 传递参数

	|参数|说明|
	|---|----|
	|a|加法运算的参数1|
	|b|加法运算的参数2|
	

###接受服务A心跳检查

**Request**

- URL

	```http
	/aliveCheck
	```
- Method

	`GET`
	
**服务B发送心跳到服务A**

​	开启一个`go routine`，每隔5秒钟，服务B向服务A发送一个携带自己地址的`POST`请求，将自己注册到服务A的可用服务列表中


	go func() {
		defer wg.Done()
		strArray := "http://" + aIp + "/registerList?ip=" + bIp
		//给注册中心发送心跳
		for {
			http.Post(strArray, "", strings.NewReader(""))
			time.Sleep(5 * time.Second)
		}
	}()

##客户端

###向服务A要服务B地址

	`GET(add)`

###连接服务B实现加法运算
	`GET(a,b)`

**B地址不可用，重新找服务A索取**
	有可能出现一种情况，当`A`给客户端服务`B`地址时，地址可用，但当客户端去连接`B`时，地址断开，这时需要客户端再去服务`A`获取`B`服务可用地址
	
	//如果拿到的地址不可用，在重新找A要，直到拿到有用的
	for ;err3 != nil;{
		log.Errorf("main || getSum failed ", err3)
		fmt.Println("getSum error:", err3)
		log.Infof("main || 重新寻找可用服务... ")
		fmt.Println("重新寻找可用服务...")
		serviceIp_,err_ := queryFunc(serAIp)
		if err_!= nil{
			log.Infof("queryFunc ||", err_)
			fmt.Println("queryFunc || get B address failed ",err_ )
			return
		}
		if serviceIp_ == "no service availble"{
			log.Infof(" no service availble||", err)
			fmt.Println("no service availble|| get B address failed ",err )
			return
		}
		strGetSum_ := "http://" + serviceIp_ + "/Add/" + a + "/" + b
		resp3, err3 = http.Get(strGetSum_)
	}


