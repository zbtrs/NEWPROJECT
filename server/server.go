package server

import (
	"NETWORKPROJECT/config"
	"fmt"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"
)

var ErrorLogLocation string
var cnt int = 0

type Server struct {
	config config.JsonConf
}

func NewServer(conf config.JsonConf) Server {
	return Server{
		config: conf,
	}
}

func (s *Server) SentMessage(text string, r config.Rule) string {
	conn, err := net.Dial("tcp", r.ProxyPass) //建立TCP连接
	if err != nil {
		AddErrorLog(ErrorLogLocation, err)
		fmt.Println("Dial err:", err)
		return ""
	}
	defer conn.Close()
	_, err = conn.Write([]byte(text)) //发送信息
	if err != nil {
		AddErrorLog(ErrorLogLocation, err)
		fmt.Println("SentMessage error:", err)
		return ""
	}
	buf := make([]byte, 100010) //读取报文,不能用ioutil.ReadAll
	n, err := conn.Read(buf[:])
	n1, _ := conn.Read(buf[n:]) //一次可能读不完
	n += n1
	if err != nil {
		AddErrorLog(ErrorLogLocation, err)
		fmt.Println("recv failed,err:", err)
		return ""
	}
	return string(buf[:n])
}

func (s *Server) SentResponse(responseText string, conn net.Conn) {
	_, err := conn.Write([]byte(responseText))
	if err != nil {
		AddErrorLog(ErrorLogLocation, err)
		fmt.Println("SentResponse error: ", err)
		return
	}
}

func Response404(conn net.Conn) {
	s := "HTTP/1.1 404 Not Found\r\n"
	s += "\r\n"
	_, _ = conn.Write([]byte(s))
}

func (s *Server) LinkSolve(sta Request, conn net.Conn, v config.Rule) {
	sta = Modify(sta, v)                      //根据配置文件更改报文
	statext := Http2String(sta)               //将报文变成string类型便于发送
	responseText := s.SentMessage(statext, v) //得到回复报文
	if CheckNot200(responseText) {            //如果状态不是正常的,输出到错误日志中
		AddErrorLog(ErrorLogLocation, fmt.Errorf(GetFirstLine(responseText)))
	}
	s.SentResponse(responseText, conn)                  //发送报文
	AddAccessLog(sta, s.config.AccessLog, responseText) //添加到连接日志中
}

func (s *Server) RequestOthers(sta Request, conn net.Conn, v config.Rule) {
	sta = Modify2(sta, v)
	statext := Http2String(sta)
	responseText := s.SentMessage(statext, v)
	if CheckNot200(responseText) {
		AddErrorLog(ErrorLogLocation, fmt.Errorf(GetFirstLine(responseText)))
	}
	s.SentResponse(responseText, conn)
	AddAccessLog(sta, s.config.AccessLog, responseText)
}

func (s *Server) matchURL(conn net.Conn, cnt *sync.WaitGroup) {
	defer conn.Close()
	defer cnt.Done()
	sta, err := Analyse(conn)
	if err != nil {
		AddErrorLog(ErrorLogLocation, err)
		fmt.Println("Analyse err: ", err)
		return
	}
	var flag bool = false //是否有匹配上
	for _, v := range s.config.Rules {
		if CheckStatic(sta.Url, v.MatchLocation) == true { //静态文件服务
			tempText := GetStatic(sta, v)              //要打开的文件位置
			responseText, _ := os.ReadFile(tempText)   //打开文件,返回一个[]byte
			responseText2 := GetResponse(responseText) //手动构建报文
			s.SentResponse(responseText2, conn)
			flag = true
			break
		} else if sta.Url == v.MatchLocation {
			s.LinkSolve(sta, conn, v)
			flag = true
			break
		}
	}
	if flag {
		return
	}
	for _, v := range s.config.Rules { //左通配符匹配
		if MarchLeft(sta.Url, v.MatchLocation) {
			s.LinkSolve(sta, conn, v)
			flag = true
			break
		}
	}
	if flag {
		return
	}
	for _, v := range s.config.Rules {
		if MarchRight(sta.Url, v.MatchLocation) {
			s.LinkSolve(sta, conn, v)
			flag = true
			break
		}
	}
	if flag {
		return
	}
	for _, v := range s.config.Rules {
		if MarchRE(sta.Url, v.MatchLocation) {
			s.LinkSolve(sta, conn, v)
			flag = true
			break
		}
	}
	if flag == false {
		s.RequestOthers(sta, conn, s.config.Rules[len(s.config.Rules)-1])
		return
	}
}

func (s *Server) MatchThread(x int, conn net.Conn) {
	sta, err := Analyse(conn)
	if err != nil {
		AddErrorLog(ErrorLogLocation, err)
		fmt.Println("Analyse err: ", err)
		return
	}
	s.LinkSolve(sta, conn, s.config.Rules[x])
}

func (s *Server) RandomLoad(conn net.Conn) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	x := r.Intn(10000) % 3
	s.MatchThread(x, conn)
}

func (s *Server) PollingLoad(conn net.Conn) {
	cnt++
	if cnt == 3 {
		cnt = 0
	}
	s.MatchThread(cnt, conn)
}

func (s *Server) WeightedRandomLoad(conn net.Conn) {
	var l = []int{0, 0, 0, 0, 1, 1, 2}
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	x := r.Intn(10000) % 7
	s.MatchThread(l[x], conn)
}

func (s *Server) LoadBalance(conn net.Conn, cnt *sync.WaitGroup) { //选择什么策略
	defer cnt.Done()
	if s.config.LoadBalanceMethod == "randomload" {
		s.RandomLoad(conn)
	}
	if s.config.LoadBalanceMethod == "pollingload" {
		s.PollingLoad(conn)
	}
	if s.config.LoadBalanceMethod == "weightedrandomload" {
		s.WeightedRandomLoad(conn)
	}
}

func (s *Server) Solve(cnt *sync.WaitGroup) {
	defer cnt.Done() //每一个goroutine都要确保计数器-1
	ErrorLogLocation = s.config.ErrorLog
	Sta, err := net.Listen("tcp", ":"+s.config.Port) //监听
	if err != nil {
		AddErrorLog(ErrorLogLocation, err)
		fmt.Println("listen error!", err)
		return
	}
	defer Sta.Close()
	for {
		conn, err := Sta.Accept()
		if err != nil {
			AddErrorLog(ErrorLogLocation, err)
			fmt.Println("Accept error: ", err)
			return
		}
		if s.config.IsLoadBalance == "false" { //是否负载均衡
			cnt.Add(1)
			go s.matchURL(conn, cnt)
		} else {
			cnt.Add(1)
			go s.LoadBalance(conn, cnt)
		}
	}
}
