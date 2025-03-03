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
	helloResource.SetList(func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		v := ctx.Value(resweave.KeyRequestID)
		msg := "Hello, World!"
		if reqID, ok := v.(string); ok {
			msg = fmt.Sprintf("%s\nRequest: '%s'\n", msg, reqID)
		}
		if bw, err := w.Write([]byte(msg)); err != nil {
			logger.Errorw("Main", "Write Error", err, "Bytes Written", bw)
		}
	})
	translateResource := resweave.NewAPI("translate")
	translateResource.SetList(func(ctx context.Context, w http.ResponseWriter, req *http.Request) {
		v := ctx.Value(resweave.KeyRequestID)
		msg := "Bonjour, Toute le monde!"
		if reqID, ok := v.(string); ok {
			msg = fmt.Sprintf("%s\nRequest: '%s'\n", msg, reqID)
		}
		if bw, err := w.Write([]byte(msg)); err != nil {
			logger.Errorw("Main", "Write Error", err, "Bytes Written", bw)
		}
	})
	if err := helloResource.AddResource(translateResource); err != nil {
		fmt.Println(err.Error())
		panic(err)
	}

	if err := server.AddResource(helloResource); err == nil {
		fmt.Println(server.Run())
	} else {
		fmt.Println(err.Error())
	}
}
