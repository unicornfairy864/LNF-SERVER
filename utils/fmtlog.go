package utils

import (
	"encoding/json"
	"fmt"
)

func LogJson(data interface{}) {
	if data == nil {
		return
	}
	str, _ := json.Marshal(data)
	fmt.Println(string(str))
}
