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

func Analyse(conn net.Conn) (Request, error) {
	scanner := bufio.NewReader(conn)
	request := new(Request)
	request.Body = make([]byte, 0)
	request.Headers = make(headers)
	request.Contents = ""

	textt, _, _ := scanner.ReadLine()
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
			Linetext := strings.Split(text, ": ")
			A := Linetext[0]
			B := Linetext[1]
			C := strings.Split(B, ", ")
			request.Headers[A] = append(request.Headers[A], C...)
		} else {
			break
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

func Modify(sta Request, r config.Rule) Request {
	var s = strings.Split(r.ProxySetHeader, ":")
	a := make([]string, 0)
	a = append(a, s[1])
	sta.Headers[s[0]] = a
	sta.Url = r.Location
	return sta
}

func Http2String(sta Request) string {
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

func GetStatic(sta Request, r config.Rule) string {
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

func GetResponse(s []byte) string {

	responseText := "HTTP/1.1 200 OK\r\n"
	//responseText += "Content-Type: text/html; charset=UTF-8\r\n"
	//responseText += "text/html;application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9\r\n"
	responseText += "Connection: keep-alive\r\n"
	//responseText += "Content-Encoding: gzip\r\n"
	responseText += "\r\n"
	responseText += string(s)
	return responseText
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

func CheckNot200(text string) bool {
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

func GetFirstLine(text string) string {
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

func MarchLeft(s1 string, s2 string) bool {
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

func MarchRE(s1 string, s2 string) bool {
	if s2 == "/" {
		return false
	} //特殊符号问题,根本不可能用到正则匹配
	//fmt.Println("MarchRe ",s1,s2)
	flag, _ := regexp.MatchString(s2, s1)
	return flag
}
