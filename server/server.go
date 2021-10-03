package server

import (
	"NETWORKPROJECT/config"
	"fmt"
	"net"
	"os"
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
