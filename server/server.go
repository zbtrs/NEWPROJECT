package server

import (
	"NETWORKPROJECT/config"
	"fmt"
	"net"
	"os"
	"strings"
)

var ErrorLogLocation string

type Server struct {
	config config.JsonConf
}

func NewServer(conf config.JsonConf) Server {
	return Server{
		config: conf,
	}
}

func (s *Server) SentMessage(text string, r config.Rule) string {
	conn, err := net.Dial("tcp", r.ProxyPass) //TODO 改为配置中的地址
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

func CheckStatic(Url string, Location string) bool {
	var flag = false
	SplitString := strings.Split(Url, "/")
	for _, i := range SplitString {
		if i == "static" {
			flag = true
			break
		}
	}
	if flag == true && Location == "/static" {
		return true
	}
	return false
}

func (s *Server) matchURL(conn net.Conn) {
	defer conn.Close()
	sta, err := Analyse(conn)
	if err != nil {
		AddErrorLog(ErrorLogLocation, err)
		fmt.Println("Analyse err: ", err)
		return
	}
	for _, v := range s.config.Rules {
		if CheckStatic(sta.Url, v.Location) == true {
			//处理静态页面
			tempText := GetStatic(sta, v)
			responseText, _ := os.ReadFile(tempText)
			responseText2 := GetResponse(responseText)
			s.SentResponse(responseText2, conn)
		} else if sta.Url == v.Location { //TODO other match rules
			sta = Modify(sta, v)
			statext := Http2String(sta)
			responseText := s.SentMessage(statext, v)
			s.SentResponse(responseText, conn)
			AddAccessLog(sta, s.config.AccessLog, responseText)
		}
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
		go s.matchURL(conn)
	}
}
