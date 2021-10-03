package main

import (
	"NETWORKPROJECT/config"
	"NETWORKPROJECT/server"
	"fmt"
)

func main() {
	conf, err := config.Solve()
	if err != nil {
		fmt.Println("json error!", err)
	}
	mainServer := server.NewServer(conf)
	mainServer.Solve() //处理接收到的报文
	//读取配置,知道要把报文发送到哪个上面去
	//对报文进行修改
	//将报文发送到对应服务器并接收
}
