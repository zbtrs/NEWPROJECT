package server

import (
	"NETWORKPROJECT/config"
	"bufio"
	"fmt"
	"net"
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
		fmt.Println("httpparse error", err)
		return Request{}, err
	}
	return *request, nil
}

func Modify(sta Request, r config.Rule) Request {
	a := make([]string, 0)
	a = append(a, r.Host)
	sta.Headers["Host"] = a
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
	responseText += "text/html;application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9\r\n"
	responseText += "Connection: keep-alive\r\n"
	//responseText += "Content-Encoding: gzip\r\n"
	responseText += "\r\n"
	responseText += string(s)
	return responseText
}
