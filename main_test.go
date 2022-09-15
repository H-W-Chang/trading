package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/google/uuid"
)

func TestMain(m *testing.M) {
	BuyOrderList = make(map[float64]*OrderList)
	SellOrderList = make(map[float64]*OrderList)
	os.Exit(m.Run())
}

func SendOrderPost(order Order) *httptest.ResponseRecorder {
	id := uuid.New()
	order.OrderId = id.String()
	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(order)
	req := httptest.NewRequest(http.MethodPost, "/order", payloadBuf)
	w := httptest.NewRecorder()
	OrderPost(w, req)
	return w
}

func TestAddOrder(t *testing.T) {
	ol := OrderList{}
	id := uuid.New()
	ol.AddOrder(Order{id.String(), "1", "gold", 0, 100, 100})
	o := ol.PopFront()
	if o.UserID != "1" || o.Op != 0 || o.Volume != 100 || o.Price != 100 {
		t.Errorf("AddOrder failed")
	}
}
func TestTransaction(t *testing.T) {
	ol := OrderList{}
	ol.TransactionStart()
	id := uuid.New()
	ol.AddOrder(Order{id.String(), "1", "gold", 0, 100, 100})
	id = uuid.New()
	ol.AddOrder(Order{id.String(), "2", "gold", 0, 100, 100})
	o := ol.PopFront()
	if o.UserID != "1" {
		t.Errorf("Transaction failed")
	}
	o = ol.PopFront()
	if o.UserID != "2" {
		t.Errorf("Transaction failed")
	}
	ol.TransactionStop()
}

func TestOrderRequest(t *testing.T) {
	id := uuid.New()
	o := Order{id.String(), "1", "gold", 0, 100, 100}
	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(o)
	req := httptest.NewRequest(http.MethodPost, "/order", payloadBuf)
	w := httptest.NewRecorder()
	OrderPost(w, req)
	res := w.Result()
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}
	if string(data) != "add buy order to pending list" {
		t.Errorf("expected add buy order to pending list got %v", string(data))
	}
}

func TestBuyOrderDealTrade(t *testing.T) {
	price := 100.10
	id := uuid.New()
	SendOrderPost(Order{id.String(), "1", "gold", 1, 2, price})
	id = uuid.New()
	SendOrderPost(Order{id.String(), "2", "gold", 1, 8, price})
	id = uuid.New()
	w := SendOrderPost(Order{id.String(), "3", "gold", 0, 3, price})
	res := w.Result()
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}
	t.Logf("response: %v", string(data))
	if string(data) != "buy order dealed" {
		t.Errorf("expected buy order dealed got %v", string(data))
	}
	length := SellOrderList[price].GetLength()
	if length != 1 {
		t.Errorf("expected sell order list length 1 got %v", length)
	}
	for i := 0; i < length; i++ {
		o := SellOrderList[price].GetOrder(0)
		t.Logf("sell order: %+v", o)
		if o.UserID != "2" {
			t.Errorf("expected 2 got %v", o.UserID)
		}
		if o.Volume != 7 {
			t.Errorf("expected 7 got %v", o.Volume)
		}
	}
}
func TestBuyOrderNotDealTrade(t *testing.T) {
	price := 100.10
	SendOrderPost(Order{"", "1", "gold", 1, 2, price})
	SendOrderPost(Order{"", "2", "gold", 1, 8, price})
	w := SendOrderPost(Order{"", "3", "gold", 0, 11, price})
	res := w.Result()
	defer res.Body.Close()
	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		t.Errorf("expected error to be nil got %v", err)
	}
	t.Logf("response: %v", string(data))
	if string(data) != "add buy order to pending list" {
		t.Errorf("expected add buy order to pending list got %v", string(data))
	}
	sellLength := SellOrderList[price].GetLength()
	if sellLength != 0 {
		t.Errorf("expected sell order list length 1 got %v", sellLength)
	}

	buyLength := BuyOrderList[price].GetLength()
	if buyLength != 1 {
		t.Errorf("expected sell order list length 1 got %v", sellLength)
	}
	for i := 0; i < buyLength; i++ {
		o := BuyOrderList[price].GetOrder(0)
		t.Logf("buy order: %+v", o)
		if o.UserID != "3" {
			t.Errorf("expected 2 got %v", o.UserID)
		}
		if o.Volume != 1 {
			t.Errorf("expected 1 got %v", o.Volume)
		}
	}
}
