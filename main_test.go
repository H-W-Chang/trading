package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"trading/pkg/database"
	"trading/pkg/entity"
	"trading/pkg/matcher"
	"trading/pkg/server"

	"github.com/google/uuid"
)

var httpTestServer *server.HttpServer

func TestMain(m *testing.M) {
	repo := database.NewRepository("memory")
	httpTestServer = server.NewServer("http", repo).(*server.HttpServer)
	os.Exit(m.Run())
}
func TestHttpAddOrder(t *testing.T) {
	id := uuid.New()
	jsonOrder := entity.Order{OrderID: id.String(), UserID: "1", Item: "gold", Op: 0, Volume: 100, Price: 100, MatchRule: "partial"}
	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(jsonOrder)
	req := httptest.NewRequest(http.MethodPost, "/order", payloadBuf)
	w := httptest.NewRecorder()
	httpTestServer.OrderReqHandler(w, req)
	res := w.Result()
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Errorf("AddOrder failed")
	}
	data, err := io.ReadAll(res.Body)
	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}
	if string(data) != matcher.Pending {
		t.Errorf("expected \"%s\" got \"%v\"", matcher.Pending, string(data))
	}
}
