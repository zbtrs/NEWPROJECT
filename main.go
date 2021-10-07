package main

import (
	"NETWORKPROJECT/config"
	"NETWORKPROJECT/server"
	"fmt"
	"sync"
)

func main() {
	conf, err := config.Solve()
	if err != nil {
		fmt.Println("json error!", err)
	}
	mainServer1 := server.NewServer(conf[0]) //访问哪个网页
	mainServer2 := server.NewServer(conf[1])
	var cnt sync.WaitGroup //计数器,防止主goroutine先退出
	cnt.Add(1)
	go mainServer1.Solve(&cnt) //处理接收到的报文
	cnt.Add(1)
	go mainServer2.Solve(&cnt)
	cnt.Wait()
	//读取配置,知道要把报文发送到哪个上面去
	//对报文进行修改
	//将报文发送到对应服务器并接收
}
