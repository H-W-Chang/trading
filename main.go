package main

import (
	// "trading/pkg/Interactor"
	"trading/pkg/server"
)

func main() {
	var server server.Server = &server.HttpServer{}
	server.Serve()
}