package server_test

import (
	"os"
	"testing"
	"trading/pkg/database"
	"trading/pkg/server"
)

var httpTestServer *server.HttpServer

func TestMain(m *testing.M) {
	repo := database.NewRepository("memory")
	httpTestServer = server.NewServer("http", repo).(*server.HttpServer)
	os.Exit(m.Run())
}
