package main

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mortedecai/resweave"
	"go.uber.org/zap"
)

func main() {
	fmt.Println("Running Server for API Integration Test: Hello")

	var logger *zap.SugaredLogger
	server := resweave.NewServer(8080)
	if l, err := zap.NewDevelopment(); err != nil {
		fmt.Println("******** COULD NOT CREATE A LOGGER!!!!!!! ************")
	} else {
		logger = l.Sugar()
		server.SetLogger(logger, true)
	}

	helloResource := resweave.NewAPI("hello")
	helloResource.SetList(func(_ context.Context, w http.ResponseWriter, req *http.Request) {
		if bw, err := w.Write([]byte("Hello, World!")); err != nil {
			logger.Errorw("Main", "Write Error", err, "Bytes Written", bw)
		}
	})

	if err := server.AddResource(helloResource); err == nil {
		fmt.Println(server.Run())
	} else {
		fmt.Println(err.Error())
	}
}
