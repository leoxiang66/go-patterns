package utils

import (
	"fmt"
	"log"
	"os"
)

func IsDev() bool {
	env := os.Getenv("env")
	isDev := env == "dev"
	return isDev
}

func LogMessage(message string) {
	if IsDev() {
		fmt.Println(message) // 开发环境使用 fmt.Println
	} else {
		log.Println(message) // 生产环境使用 log
	}
}
