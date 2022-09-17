package server

import (
	"encoding/json"
	"log"
	"net/http"
	"trading/pkg/matcher"
)

type HttpServer struct {
}

func (h *HttpServer) Serve() {
	mux := http.NewServeMux()
	mux.HandleFunc("/order", h.OrderReqHandler)
	http.ListenAndServe(":8080", mux)
}

func (h *HttpServer) OrderReqHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("OrderReqHandler")
	if r.Method == "POST" {
		var newOrder matcher.Order
		err := json.NewDecoder(r.Body).Decode(&newOrder)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		err = newOrder.Validate()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		var newMatcher matcher.Matcher = GetMatcher(newOrder.MatchRule)
		result, err := newMatcher.Match(newOrder)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Write([]byte(result))
	}
}
