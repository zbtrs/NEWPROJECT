package server

import (
	"NETWORKPROJECT/config"
	"fmt"
	"net"
	"os"
	"strings"
)

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
		fmt.Println("Dial err:", err)
		return ""
	}
	defer conn.Close()
	_, err = conn.Write([]byte(text))
	if err != nil {
		fmt.Println("SentMessage error:", err)
		return ""
	}
	buf := make([]byte, 100010)
	n, err := conn.Read(buf[:])
	n1, _ := conn.Read(buf[n:])
	n += n1
	if err != nil {
		fmt.Println("recv failed,err:", err)
		return ""
	}
	return string(buf[:n])
}

func (s *Server) SentResponse(responseText string, conn net.Conn) {
	_, err := conn.Write([]byte(responseText))
	if err != nil {
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
		}
	}
	if err != nil {
		return
	}

}

func (s *Server) Solve() {
	Sta, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("listen error!", err)
		return
	}
	defer Sta.Close()
	for {
		conn, err := Sta.Accept()
		if err != nil {
			return
		}
		go s.matchURL(conn)
	}
}
