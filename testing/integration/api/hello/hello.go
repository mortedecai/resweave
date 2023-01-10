package main

import (
	"fmt"

	"github.com/mortedecai/resweave"
	"go.uber.org/zap"
)

func main() {
	fmt.Println("Running Server for API Integration Test: Hello")

	server := resweave.NewServer(8080)
	if l, err := zap.NewDevelopment(); err != nil {
		fmt.Println("******** COULD NOT CREATE A LOGGER!!!!!!! ************")
	} else {
		server.SetLogger(l.Sugar(), true)
	}

	helloResource := resweave.NewAPI("hello")
	if err := server.AddResource(helloResource); err == nil {
		fmt.Println(server.Run())
	} else {
		fmt.Println(err.Error())
	}
}
