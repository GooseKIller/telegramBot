package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)
// go run ./*.go
func chatGPT(userId int, message string) string {
	app := "/usr/bin/python3.8"

	cmd := exec.Command(app, "/home/anton/TeleBotGo/pythonScripts/chatBot.py")

	input := fmt.Sprintf("%d;%s", userId, message)

	cmd.Stdin = strings.NewReader(input)

	output, err := cmd.Output()

	if err != nil {
		log.Fatal(err)
	}

	return string(output)
}

func clearHistory(userId int){

}
