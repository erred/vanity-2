package main

import (
	"net/http"

	"go.seankhliao.com/svcrunner"
	"go.seankhliao.com/vanity/server"
)

func main() {
	hs := &http.Server{}
	svr := server.New(hs)
	svcrunner.Options{}.Run(
		svcrunner.NewHTTP(hs, svr.Register, svr.Init),
	)
}
