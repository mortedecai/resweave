package main

import (
	"fmt"

	"github.com/mortedecai/resweave"
)

func main() {
	fmt.Println("Running Server for HTML Integration Test: Multi Host Hello")
	server := resweave.NewServer(80)
	htmlResource := resweave.NewHTML("", "./html/default")
	caHtmlResource := resweave.NewHTML("", "./html/caHost")
	if err := server.AddResource(htmlResource); err != nil {
		fmt.Println("Could not add default resource because: ", err.Error())
	}
	if h, err := server.AddHost("mortedecai-ca"); err == nil {
		if err := h.AddResource(caHtmlResource); err != nil {
			fmt.Println("Could not add caHtmlResource resource because: ", err.Error())
		}
	} else {
		fmt.Println("Could not add caHost because: ", err.Error())
	}
	fmt.Println(server.Run())
}
