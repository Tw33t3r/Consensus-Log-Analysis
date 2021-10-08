package main

import (
	"encoding/json"
)

type logData struct {
	Level   string
	Port    string
	Ip      string
	logData interface{}
	caller  string
	Time    string
	Message string
}

func parse(input string) logData {
	var parsedLog logData
	json.Unmarshal(([]byte(input)), &parsedLog)
	return parsedLog
}
