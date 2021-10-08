package main

import (
	"encoding/json"
)

func parse(input string) map[string]interface{} {
	var parsedLog map[string]interface{}
	json.Unmarshal(([]byte(input)), &parsedLog)
	return parsedLog
}
