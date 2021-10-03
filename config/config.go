package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

type Rule struct {
	Location  string `json:"location"`
	Host      string `json:"Host"`
	ProxyPass string `json:"proxy_pass"`
	Root      string `json:"root"`
	Index     string `json:"index"`
}
type JsonConf struct {
	Port       string `json:"port"`
	ServerName string `json:"server_name"`
	Rules      []Rule `json:"rules"`
}

func Solve() ([]JsonConf, error) {
	var Res = make([]JsonConf, 5)                                             //TODO 重构
	fileData, err := ioutil.ReadFile("D:\\NETWORKPROJECT\\config\\conf.json") // TODO cmd flag
	if err != nil {
		return []JsonConf{}, err
	}
	err = json.Unmarshal([]byte(fileData), &Res)
	if err != nil {
		return []JsonConf{}, err
	}
	fmt.Println(Res[0].Port, Res[0].ServerName, Res[1].Port, Res[1].ServerName, Res[0].Rules[0].Host)
	return Res, nil
}
