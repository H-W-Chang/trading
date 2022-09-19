package database_test

import (
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"trading/pkg/database"
	"trading/pkg/entity"
)

var repo entity.PendingOrderRepository

func TestMain(m *testing.M) {
	repo = database.NewRepository("memory")
	os.Exit(m.Run())
}
func TestGetDatabase(t *testing.T) {
	newrepo := database.NewRepository("unknown")
	if newrepo != nil {
		t.Error("expected nil got", repo)
	}
}

func TestQueryEmpty(t *testing.T) {
	price := 100.0
	condition := entity.QueryCondition{Op: 0, Price: 100.0}
	orders := repo.Query(condition)
	if len(orders) != 0 {
		t.Errorf("expected 0 got %d", len(orders))
	}
	t.Cleanup(func() { repo.Delete(entity.QueryCondition{Op: entity.All, Price: price}) })
}

func TestUpdateEmpty(t *testing.T) {
	id := 1
	price := 100.0
	condition := entity.QueryCondition{Op: 0, Price: 100.0}
	var orders []entity.Order
	for i := 0; i < 10; i, id = i+1, id+1 {
		orders = append(orders, entity.Order{OrderID: strconv.Itoa(id), UserID: strconv.Itoa(id), Item: "gold", Op: 0, Volume: 10, Price: price, MatchRule: "partial"})
	}
	repo.Update(condition, orders)
	orders = repo.Query(condition)
	t.Logf("orders: %+v", orders)
	if len(orders) != 10 {
		t.Errorf("expected 10 got %d", len(orders))
	}
	orders = []entity.Order{}
	for i := 0; i < 5; i, id = i+1, id+1 {
		orders = append(orders, entity.Order{OrderID: strconv.Itoa(id), UserID: strconv.Itoa(id), Item: "gold", Op: 0, Volume: 10, Price: price, MatchRule: "partial"})
	}
	repo.Update(condition, orders)
	orders = repo.Query(condition)
	t.Logf("orders: %+v", orders)
	if len(orders) != 5 {
		t.Errorf("expected 5 got %d", len(orders))
	}
	t.Cleanup(func() { repo.Delete(entity.QueryCondition{Op: entity.All, Price: price}) })
}

func TestInsert(t *testing.T) {
	id := 1
	price := 100.0
	condition := entity.QueryCondition{Op: 0, Price: price}
	order := entity.Order{OrderID: strconv.Itoa(id), UserID: strconv.Itoa(id), Item: "gold", Op: 0, Volume: 100, Price: price, MatchRule: "partial"}
	id++
	err := repo.Insert(condition, order)
	if err != nil {
		t.Error(err)
	}
	order = entity.Order{OrderID: strconv.Itoa(id), UserID: strconv.Itoa(id), Item: "gold", Op: 0, Volume: 100, Price: price, MatchRule: "partial"}
	err = repo.Insert(condition, order)
	if err != nil {
		t.Error("expected error got nil")
	}
	orders := repo.Query(condition)
	t.Logf("orders: %+v", orders)
	if len(orders) != 2 {
		t.Errorf("expected 2 got %d", len(orders))
	}
	t.Cleanup(func() { repo.Delete(entity.QueryCondition{Op: entity.All, Price: price}) })
}

func TestDelete(t *testing.T) {
	id := 1
	price := 100.0
	buyCondition := entity.QueryCondition{Op: 0, Price: price}
	sellCondition := entity.QueryCondition{Op: 1, Price: price}
	//insert buy
	order := entity.Order{OrderID: strconv.Itoa(id), UserID: strconv.Itoa(id), Item: "gold", Op: 0, Volume: 100, Price: price, MatchRule: "partial"}
	id++
	err := repo.Insert(buyCondition, order)
	if err != nil {
		t.Error(err)
	}
	//insert sell
	order = entity.Order{OrderID: strconv.Itoa(id), UserID: strconv.Itoa(id), Item: "gold", Op: 1, Volume: 100, Price: price, MatchRule: "partial"}
	id++
	err = repo.Insert(sellCondition, order)
	if err != nil {
		t.Error(err)
	}
	//delete buy
	repo.Delete(buyCondition)
	orders := repo.Query(buyCondition)
	if len(orders) != 0 {
		t.Errorf("expected 0 got %d", len(orders))
	}
	//detete sell
	repo.Delete(sellCondition)
	orders = repo.Query(sellCondition)
	if len(orders) != 0 {
		t.Errorf("expected 0 got %d", len(orders))
	}
}

// use case
// 3 order in sell, 2 buy order coming
func TestTransaction(t *testing.T) {
	var wg sync.WaitGroup
	var id int32 = 0
	price := 100.0
	//insert 3 sell order, volume 10
	for i := 0; i < 3; i++ {
		atomic.AddInt32(&id, 1)
		order := entity.Order{OrderID: strconv.Itoa(int(id)), UserID: strconv.Itoa(int(id)), Item: "gold", Op: 1, Volume: 10, Price: price, MatchRule: "partial"}
		repo.Insert(entity.QueryCondition{Op: 1, Price: price}, order)
	}
	wg.Add(2)
	// first buy order coming
	go func() {
		atomic.AddInt32(&id, 1)
		condition := entity.QueryCondition{Op: 1, Price: price}
		condition.OrderID = strconv.Itoa(int(id))
		lockId := repo.Lock(condition)
		condition.LockId = lockId
		condition.OrderID = ""
		orders := repo.Query(condition)
		if len(orders) != 3 {
			t.Errorf("expected 3 got %d", len(orders))
		}
		repo.Delete(condition)
		orders = repo.Query(condition)
		if len(orders) != 0 {
			t.Errorf("expected 0 got %d", len(orders))
		}
		//insert 3 sell order
		for i := 0; i < 3; i++ {
			atomic.AddInt32(&id, 1)
			order := entity.Order{OrderID: strconv.Itoa(int(id)), UserID: strconv.Itoa(int(id)), Item: "gold", Op: 1, Volume: 10, Price: price, MatchRule: "partial"}
			repo.Insert(condition, order)
			time.Sleep(time.Second)
		}
		orders = repo.Query(condition)
		if len(orders) != 3 {
			t.Errorf("expected 3 got %d", len(orders))
		}
		repo.Unlock(condition)
		wg.Done()
	}()
	time.Sleep(1 * time.Second)
	go func() {
		atomic.AddInt32(&id, 1)
		condition.OrderID = strconv.Itoa(int(id))
		lockId := repo.Lock(condition)
		condition.LockId = lockId
		condition.OrderID = ""
		var orders []entity.Order
		for i := 0; i < 5; i++ {
			atomic.AddInt32(&id, 1)
			order := entity.Order{OrderID: strconv.Itoa(int(id)), UserID: strconv.Itoa(int(id)), Item: "gold", Op: 1, Volume: 10, Price: price, MatchRule: "partial"}
			orders = append(orders, order)
		}
		repo.Update(condition, orders)
		orders = repo.Query(condition)
		if len(orders) != 5 {
			t.Errorf("expected 5 got %d", len(orders))
		}
		repo.Unlock(condition)
		wg.Done()
	}()
	wg.Wait()
	orders := repo.Query(condition)
	t.Logf("orders: %+v", orders)
	if len(orders) != 5 {
		t.Errorf("expected 5 got %d", len(orders))
	}
}
