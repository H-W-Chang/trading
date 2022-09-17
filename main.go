package main

import (
	"trading/pkg/database"
	"trading/pkg/server"
)

func main() {
	repo := database.NewRepository("memory")
	server := server.NewServer("http", repo)
	server.Serve()
}
