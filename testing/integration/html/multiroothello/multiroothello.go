package main

import (
	"fmt"

	"github.com/mortedecai/resweave"
	"go.uber.org/zap"
)

func main() {
	fmt.Println("Running Server for HTML Integration Test: Multi Host Hello")
	server := resweave.NewServer(80)
	if l, err := zap.NewDevelopment(); err != nil {
		fmt.Println("******** COULD NOT CREATE A LOGGER!!!!!!! ************")
	} else {
		server.SetLogger(l.Sugar(), true)
	}

	if err := server.AddResource(resweave.NewHTML("", "./html/folderOne")); err != nil {
		fmt.Println("Could not add default resource because: ", err.Error())
	}
	if err := server.AddResource(resweave.NewHTML("two", "./html/folderTwo")); err != nil {
		fmt.Println("Could not add resource named 'two' because: ", err.Error())
	}
	fmt.Println(server.Run())
}
