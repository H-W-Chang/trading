package server_test

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"trading/pkg/database"
	"trading/pkg/matcher"
	"trading/pkg/server"

	"github.com/google/uuid"
)

// var httpTestServer server.Server
var httpTestServer *server.HttpServer

func TestMain(m *testing.M) {
	httpTestServer = server.NewServer("http", &database.MemoryRepository{}).(*server.HttpServer)
	m.Run()
}

func TestHttpAddOrder(t *testing.T) {
	id := uuid.New()
	jsonOrder := matcher.Order{OrderID: id.String(), UserID: "1", Item: "gold", Op: 0, Volume: 100, Price: 100, MatchRule: "partial"}
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
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}
	if string(data) != matcher.Pending {
		t.Errorf("expected \"%s\" got \"%v\"", matcher.Pending, string(data))
	}
}
