package server

import (
	"log"
	"os"
)

func getAccessLog(sta Request, responseText string) string {
	s := sta.Method + " " + sta.Url + " " + sta.Version + "\n"
	s += GetFirstLine(responseText)
	s += "User-Agent: "
	for i := 0; i < len(sta.Headers["User-Agent"]); i++ {
		if i != 0 {
			s += ","
		}
		s += sta.Headers["User-Agent"][i]
	}
	s += "\n"
	s += "Host: "
	for i := 0; i < len(sta.Headers["Host"]); i++ {
		if i != 0 {
			s += ","
		}
		s += sta.Headers["Host"][i]
	}
	s += "\n"
	return s
}

func AddAccessLog(sta Request, position string, responseText string) {
	file, err := os.OpenFile(position, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	log.SetOutput(file)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	text := getAccessLog(sta, responseText)
	log.Print(text)
}

func AddErrorLog(position string, errText error) {
	file, err := os.OpenFile(position, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	log.SetOutput(file)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	log.Print(errText)
}
