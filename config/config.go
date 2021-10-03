package config

import (
	"encoding/json"
	"io/ioutil"
)

type Rule struct {
	Location       string `json:"location"`
	MatchLocation  string `json:"matchLocation"`
	ProxySetHeader string `json:"proxy_set_header"`
	ProxyPass      string `json:"proxy_pass"`
	Root           string `json:"root"`
	Index          string `json:"index"`
}
type JsonConf struct {
	Port       string `json:"port"`
	ServerName string `json:"server_name"`
	ErrorLog   string `json:"error_log"`
	AccessLog  string `json:"access_log"`
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
	return Res, nil
}
