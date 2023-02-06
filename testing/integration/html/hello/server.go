package main

import (
	"fmt"

	"github.com/mortedecai/resweave"
	"go.uber.org/zap"
)

func main() {
	fmt.Println("Running Server for HTML Integration Test: Hello")

	server := resweave.NewServer(8080)
	if l, err := zap.NewDevelopment(); err != nil {
		fmt.Println("******** COULD NOT CREATE A LOGGER!!!!!!! ************")
	} else {
		server.SetLogger(l.Sugar(), true)
	}

	htmlResource := resweave.NewHTML("", "./html")
	if err := server.AddResource(htmlResource); err == nil {
		fmt.Println(server.Run())
	} else {
		fmt.Println(err.Error())
	}
}
