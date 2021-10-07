package config

import (
	"encoding/json"
	"io/ioutil"
)

type Rule struct {
	Root           string `json:"root"`
	Index          string `json:"index"`
	Location       string `json:"location"`
	ProxyPass      string `json:"proxy_pass"`
	MatchLocation  string `json:"matchLocation"`
	ProxySetHeader string `json:"proxy_set_header"`
}
type JsonConf struct {
	Port              string `json:"port"`
	ErrorLog          string `json:"error_log"`
	AccessLog         string `json:"access_log"`
	ServerName        string `json:"server_name"`
	IsLoadBalance     string `json:"isLoadBalance"`
	LoadBalanceMethod string `json:"loadBalanceMethod"`
	Rules             []Rule `json:"rules"`
}

func Solve() ([]JsonConf, error) {
	var Res = make([]JsonConf, 5)
	fileData, err := ioutil.ReadFile("D:\\NETWORKPROJECT\\config\\conf.json")
	if err != nil {
		return []JsonConf{}, err
	}
	err = json.Unmarshal([]byte(fileData), &Res)
	if err != nil {
		return []JsonConf{}, err
	}
	return Res, nil
}
