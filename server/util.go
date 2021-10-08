package server

import (
	"NETWORKPROJECT/config"
	"bufio"
	"fmt"
	"net"
	"regexp"
	"strings"
)

type headers map[string][]string

type Request struct {
	Version  string
	Url      string
	Method   string
	Headers  map[string][]string
	Body     []byte
	Contents string
}

func Analyse(conn net.Conn) (Request, error) { //提取报文各成分
	scanner := bufio.NewReader(conn)
	request := new(Request)
	request.Body = make([]byte, 0)
	request.Headers = make(headers)
	request.Contents = ""

	textt, _, _ := scanner.ReadLine() //单独处理第一行
	text := string(textt)
	Linetext := strings.Split(text, " ")
	if len(Linetext) != 3 {
		return *request, nil
	}
	request.Method = Linetext[0]
	request.Url = Linetext[1]
	request.Version = Linetext[2]

	for {
		textt, _, _ := scanner.ReadLine()
		text := string(textt)
		request.Contents += text + "\r\n"
		if text != "" {
			Linetext := strings.Split(text, ": ") //记录header
			A := Linetext[0]
			B := Linetext[1]
			C := strings.Split(B, ", ")
			request.Headers[A] = append(request.Headers[A], C...)
		} else {
			break //读到空行了
		}
	}
	_, err := scanner.Read(request.Body)
	if err != nil {
		AddErrorLog(ErrorLogLocation, err)
		fmt.Println("httpparse error", err)
		return Request{}, err
	}
	return *request, nil
}

func Modify(sta Request, r config.Rule) Request { //根据配置文件修改报文
	var s = strings.Split(r.ProxySetHeader, ":")
	a := make([]string, 0)
	a = append(a, s[1])
	sta.Headers[s[0]] = a
	a = make([]string, 0)
	a = append(a, "close") //如果是keep-alive会出现问题
	sta.Headers["Connection"] = a
	sta.Url = r.Location
	return sta
}

func Modify2(sta Request, r config.Rule) Request {
	var s = strings.Split(r.ProxySetHeader, ":")
	a := make([]string, 0)
	a = append(a, s[1])
	sta.Headers[s[0]] = a
	return sta
}

func Http2String(sta Request) string { //将报文变成string,一行行处理
	var res string = ""
	res += sta.Method + " " + sta.Url + " " + sta.Version + "\r\n"
	for A, B := range sta.Headers {
		res += A + ": "
		for i := 0; i < len(B); i++ {
			res += B[i]
			if i != len(B)-1 {
				res += ", "
			}
		}
		res += "\r\n"
	}
	res += "\r\n" + string(sta.Body)
	return res
}

func GetStatic(sta Request, r config.Rule) string { //得到静态文件位置
	SplitString := strings.Split(sta.Url, "/")
	StaticString := ""
	pos := 0
	for k, v := range SplitString {
		if v == "static" {
			pos = k
			break
		}
	}
	if pos == len(SplitString)-1 { //打开默认文件
		StaticString = r.Root + "\\" + r.Index
		return StaticString
	}
	StaticString = r.Root
	for i := pos + 1; i < len(SplitString); i++ {
		StaticString += "\\" + SplitString[i]
	}
	return StaticString
	//否则就是要拼接路径
}

func GetResponse(s []byte) string { //手动构建回复报文
	responseText := "HTTP/1.1 200 OK\r\n"
	responseText += "Connection: close\r\n"
	responseText += "\r\n"
	responseText += string(s)
	return responseText
}

func CheckStatic(Url string, Location string) bool { //检查是否是静态文件服务
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

func CheckNot200(text string) bool { //状态是否正常
	s1 := []byte(text)
	var flag = true
	for i := 0; i < len(s1)-1; i++ {
		if s1[i] == 'O' && s1[i+1] == 'K' {
			flag = false
			break
		}
	}
	return flag
}

func GetFirstLine(text string) string { //得到string的第一行,以便提取信息
	s1 := []byte(text)
	s2 := make([]byte, 0)
	for i := 0; i < len(s1); i++ {
		if s1[i] == '\r' {
			break
		}
		s2 = append(s2, s1[i])
	}
	return string(s2) + "\n"
}

func MarchLeft(s1 string, s2 string) bool { //是否左匹配
	b1, b2 := []byte(s1), []byte(s2)
	if len(b2) == 0 || b2[0] != '*' || len(b1) < len(b2)-1 {
		return false
	}
	for i, j := len(b1)-1, len(b2)-1; j > 0; {
		if b1[i] != b2[j] {
			return false
		}
		i--
		j--
	}
	return true
}

func MarchRight(s1 string, s2 string) bool {
	b1, b2 := []byte(s1), []byte(s2)
	//fmt.Println("check2 ",s1,s2,b1,b2,len(b1),len(b2),b2[len(b2) - 1] == '*')
	if len(b2) == 0 || b2[len(b2)-1] != '*' || len(b1) < len(b2)-1 {
		//fmt.Println("check3")
		return false
	}
	for i, j := 0, 0; j < len(b2)-1; {
		if b1[i] != b2[j] {
			return false
		}
		i++
		j++
	}
	return true
}

func MarchRE(s1 string, s2 string) bool { //是否是正则匹配
	if s2 == "/" {
		return false
	} //特殊符号问题,根本不可能用到正则匹配
	flag, _ := regexp.MatchString(s2, s1)
	return flag
}
