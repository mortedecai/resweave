package main

import (
	"fmt"

	"github.com/mortedecai/resweave"
)

func main() {
	fmt.Println("Running Server for HTML Integration Test: Hello")
	server := resweave.NewServer(8080)
	htmlResource := resweave.NewHTML("", "./html")
	if err := server.AddResource(htmlResource); err == nil {
		fmt.Println(server.Run())
	} else {
		fmt.Println(err.Error())
	}
}
