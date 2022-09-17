package database_test

import (
	"os"
	"testing"
	"trading/pkg/database"
	"trading/pkg/matcher"

	"github.com/google/uuid"
)

var repo matcher.OrderRepository

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

func TestMemoryRepository(t *testing.T) {
	id := uuid.New()
	order := matcher.Order{OrderID: id.String(), UserID: "1", Item: "gold", Op: 0, Volume: 100, Price: 100.0, MatchRule: "partial"}
	condition := matcher.QueryCondition{Op: 0, Price: 100.0}
	err := repo.Insert(condition, order)
	if err != nil {
		t.Error(err)
	}
	orders := repo.Query(condition)
	if len(orders) != 1 {
		t.Errorf("expected 1 got %d", len(orders))
	}
	t.Logf("orders: %+v", orders)
	t.Cleanup(func() { repo.Delete(condition) })
}

func TestUpdateOrders(t *testing.T) {
	price := 100.0
	condition := matcher.QueryCondition{Op: 0, Price: price}
	repo.Update(condition, []matcher.Order{})
	orders := repo.Query(condition)
	if len(orders) != 0 {
		t.Errorf("expected 0 got %d", len(orders))
	}
	id := uuid.New()
	order := matcher.Order{OrderID: id.String(), UserID: "1", Item: "gold", Op: 0, Volume: 100, Price: 100.0, MatchRule: "partial"}
	repo.Update(condition, []matcher.Order{order})
	orders = repo.Query(condition)
	if len(orders) != 1 {
		t.Errorf("expected 1 got %d", len(orders))
	}
	t.Logf("orders: %+v", orders)
	t.Cleanup(func() { repo.Delete(condition) })
}
func TestFindByPrice(t *testing.T) {
	price := 100.0
	condition := matcher.QueryCondition{Op: 0, Price: price}
	orders := repo.Query(condition)
	if len(orders) != 0 {
		t.Errorf("expected 0 got %d", len(orders))
	}
	id := uuid.New()
	order := matcher.Order{OrderID: id.String(), UserID: "1", Item: "gold", Op: 0, Volume: 100, Price: 100.0, MatchRule: "partial"}
	err := repo.Insert(condition, order)
	if err != nil {
		t.Error(err)
	}
	orders = repo.Query(condition)
	if len(orders) != 1 {
		t.Errorf("expected 1 got %d", len(orders))
	}
	t.Logf("orders: %+v", orders)
	t.Cleanup(func() { repo.Delete(condition) })
}
