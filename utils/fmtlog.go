package utils

import "fmt"

func LogJson(data interface{}) {
	if data == nil {
		return
	}
	fmt.Println(data)
}
