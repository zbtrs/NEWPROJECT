package server

import (
	"NETWORKPROJECT/config"
	"fmt"
	"math/rand"
	"net"
	"os"
	"time"
)

var ErrorLogLocation string
var ch1, ch2, ch3 = make(chan net.Conn), make(chan net.Conn), make(chan net.Conn)
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
	conn, err := net.Dial("tcp", r.ProxyPass)
	if err != nil {
		AddErrorLog(ErrorLogLocation, err)
		fmt.Println("Dial err:", err)
		return ""
	}
	defer conn.Close()
	_, err = conn.Write([]byte(text))
	if err != nil {
		AddErrorLog(ErrorLogLocation, err)
		fmt.Println("SentMessage error:", err)
		return ""
	}
	buf := make([]byte, 100010)
	n, err := conn.Read(buf[:])
	n1, _ := conn.Read(buf[n:])
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
	//fmt.Println("check404")
	s := "HTTP/1.1 404 Not Found\r\n"
	s += "\r\n"
	_, _ = conn.Write([]byte(s))
}

func (s *Server) LinkSolve(sta Request, conn net.Conn, v config.Rule) {
	//fmt.Println("LinkCheck",v.MatchLocation,v.Location)
	sta = Modify(sta, v)
	statext := Http2String(sta)
	responseText := s.SentMessage(statext, v)
	if CheckNot200(responseText) {
		AddErrorLog(ErrorLogLocation, fmt.Errorf(GetFirstLine(responseText)))
	}
	s.SentResponse(responseText, conn)
	AddAccessLog(sta, s.config.AccessLog, responseText)
}

func (s *Server) matchURL(conn net.Conn) {
	defer conn.Close()
	sta, err := Analyse(conn)
	if err != nil {
		AddErrorLog(ErrorLogLocation, err)
		fmt.Println("Analyse err: ", err)
		return
	}
	//fmt.Println("check ",sta.Url)
	var flag bool = false //是否有匹配上
	for _, v := range s.config.Rules {
		if CheckStatic(sta.Url, v.MatchLocation) == true { //静态文件服务
			tempText := GetStatic(sta, v)
			responseText, _ := os.ReadFile(tempText)
			responseText2 := GetResponse(responseText)
			s.SentResponse(responseText2, conn)
			flag = true
			break
		} else if sta.Url == v.MatchLocation { //TODO other match rules
			s.LinkSolve(sta, conn, v)
			flag = true
			break
		}
	}
	//fmt.Println("check1 ",flag)
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
	//fmt.Println("check2 ",flag)
	if flag {
		return
	}
	for _, v := range s.config.Rules {
		if MarchRight(sta.Url, v.MatchLocation) {
			//fmt.Println("sb! ",sta.Url,v.MatchLocation,v.Location)
			s.LinkSolve(sta, conn, v)
			flag = true
			break
		}
	}
	//fmt.Println("check3 ",flag)
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
		Response404(conn)
		AddErrorLog(ErrorLogLocation, fmt.Errorf("Can't March URL"))
		return
	}
}

func (s *Server) Thread1() {
	for {
		conn := <-ch1
		s.matchURL(conn)
	}
}

func (s *Server) Thread2() {
	for {
		conn := <-ch2
		s.matchURL(conn)
	}
}

func (s *Server) Thread3() {
	for {
		conn := <-ch3
		s.matchURL(conn)
	}
}

func (s *Server) MatchThread(x int, conn net.Conn) {
	if x == 0 {
		ch1 <- conn
	}
	if x == 1 {
		ch2 <- conn
	}
	if x == 3 {
		ch3 <- conn
	}
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

func (s *Server) LoadBalance(conn net.Conn) {
	go s.Thread1()
	go s.Thread2()
	go s.Thread3()
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

func (s *Server) Solve() {
	ErrorLogLocation = s.config.ErrorLog
	Sta, err := net.Listen("tcp", ":"+s.config.Port)
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
		s.LoadBalance(conn)
		//go s.matchURL(conn)
	}
}
