package matcher

import (
	"log"
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"trading/pkg/database"
	"trading/pkg/entity"

	fuzz "github.com/google/gofuzz"
)

var repo entity.PendingOrderRepository
var partialMatcher PartialMatcher

func TestMain(m *testing.M) {
	// setup
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	repo = database.NewRepository("memory")
	partialMatcher = PartialMatcher{Or: repo}
	m.Run()
	// teardown
}

func TestMatchOne(t *testing.T) {
	price := 100.0
	buyOrder := entity.Order{OrderID: "1", UserID: "1", Item: "gold", Op: 0, Volume: 10, Price: price, MatchRule: "partial"}
	sellOrder := entity.Order{OrderID: "2", UserID: "2", Item: "gold", Op: 1, Volume: 10, Price: price, MatchRule: "partial"}
	repo.Insert(entity.QueryCondition{Op: entity.Buy, Price: price}, buyOrder)
	result, err := partialMatcher.Match(sellOrder)
	if err != nil || result != Deal {
		t.Errorf("expected deal got %v, err: %v", result, err)
	}
	t.Cleanup(func() { repo.Delete(entity.QueryCondition{Op: entity.All, Price: price}) })
}
func TestMatchMany(t *testing.T) {
	price := 100.0
	buyOrder := []entity.Order{{OrderID: "1", UserID: "1", Item: "gold", Op: 0, Volume: 10, Price: price, MatchRule: "partial"},
		{OrderID: "2", UserID: "2", Item: "gold", Op: 0, Volume: 10, Price: price, MatchRule: "partial"}}
	sellOrder := entity.Order{OrderID: "2", UserID: "3", Item: "gold", Op: 1, Volume: 15, Price: price, MatchRule: "partial"}
	for i := range buyOrder {
		repo.Insert(entity.QueryCondition{Op: entity.Buy, Price: price}, buyOrder[i])
	}
	result, err := partialMatcher.Match(sellOrder)
	if err != nil || result != Deal {
		t.Errorf("expected deal got %v, err: %v", result, err)
	}
	buyPendingOrder := repo.Query(entity.QueryCondition{Op: entity.Buy, Price: price})
	if len(buyPendingOrder) != 1 {
		t.Errorf("expected 1 got %d", len(buyPendingOrder))
	}
	if buyPendingOrder[0].Volume != 5 {
		t.Errorf("expected 5 got %d", buyPendingOrder[0].Volume)
	}
	t.Logf("buyPendingOrder: %+v", buyPendingOrder)
	t.Cleanup(func() { repo.Delete(entity.QueryCondition{Op: entity.All, Price: price}) })
}

func FuzzMatchOne(f *testing.F) {
	f.Add("1", "1", int8(0), 10, 100.0)
	f.Add("2", "2", int8(1), 1, 91.2)
	f.Fuzz(func(t *testing.T, orderId string, userId string, op int8, volume int, price float64) {
		pendingOrder := entity.Order{OrderID: orderId, UserID: userId, Item: "gold", Op: op, Volume: volume, Price: price, MatchRule: "partial"}
		t.Logf("pendingOrder: %+v", pendingOrder)
		err := pendingOrder.Validate()
		if err != nil {
			return
		}
		repo.Insert(entity.QueryCondition{Op: pendingOrder.Op, Price: price}, pendingOrder)
		newOrder := entity.Order{OrderID: orderId, UserID: userId, Item: "gold", Op: entity.GetOppositeOp(pendingOrder.Op), Volume: volume, Price: price, MatchRule: "partial"}
		t.Logf("newOrder: %+v", newOrder)
		result, err := partialMatcher.Match(newOrder)
		if err != nil || result != Deal {
			t.Errorf("expected deal got %v, err: %v", result, err)
		}
		t.Cleanup(func() { repo.Delete(entity.QueryCondition{Op: entity.All, Price: price}) })
	})
}
func FuzzMatchMany(f *testing.F) {
	f.Add("1", "1", int8(0), 100.0)
	f.Add("2", "2", int8(1), 91.2)
	f.Fuzz(func(t *testing.T, orderId string, userId string, op int8, price float64) {
		countFuzz := fuzz.New().Funcs(func(i *int, c fuzz.Continue) {
			*i = c.Intn(10) + 1
		})
		volumeFuzz := fuzz.New().Funcs(func(i *int, c fuzz.Continue) {
			*i = c.Intn(100) + 1
		})
		var count int
		countFuzz.Fuzz(&count)
		t.Logf("pending count: %d", count)
		var totalPendingVolume, totalNewVolume int = 0, 0
		for i := 0; i < count; i++ {
			var volume int
			volumeFuzz.Fuzz(&volume)
			pendingOrder := entity.Order{OrderID: strconv.Itoa(i + 1), UserID: strconv.Itoa(i + 1), Item: "gold", Op: op, Volume: volume, Price: price, MatchRule: "partial"}
			t.Logf("pendingOrder: %+v", pendingOrder)
			err := pendingOrder.Validate()
			if err != nil {
				return
			}
			totalPendingVolume += volume
			repo.Insert(entity.QueryCondition{Op: pendingOrder.Op, Price: price}, pendingOrder)
		}
		t.Logf("totalPendingVolume %v", totalPendingVolume)
		oppositeOp := entity.GetOppositeOp(op)
		countFuzz.Fuzz(&count)
		t.Logf("new count: %d", count)
		for i := 0; i < count; i++ {
			var volume int
			volumeFuzz.Fuzz(&volume)
			newOrder := entity.Order{OrderID: strconv.Itoa(i + 1), UserID: strconv.Itoa(i + 1), Item: "gold", Op: oppositeOp, Volume: volume, Price: price, MatchRule: "partial"}
			t.Logf("newOrder: %+v", newOrder)
			err := newOrder.Validate()
			if err != nil {
				return
			}
			totalNewVolume += volume
			result, err := partialMatcher.Match(newOrder)
			t.Logf("result: %v, err: %v", result, err)
		}
		t.Logf("totalNewVolume %v", totalNewVolume)
		if totalNewVolume > totalPendingVolume {
			pendingOrder := repo.Query(entity.QueryCondition{Op: oppositeOp, Price: price})
			volumeLeft := 0
			for i := range pendingOrder {
				volumeLeft += pendingOrder[i].Volume
				t.Logf("id: %v, volume: %v", pendingOrder[i].OrderID, pendingOrder[i].Volume)
			}
			if volumeLeft != (totalNewVolume - totalPendingVolume) {
				t.Errorf("expected %d got %d", totalNewVolume-totalPendingVolume, volumeLeft)
			}
		} else if totalNewVolume < totalPendingVolume {
			pendingOrder := repo.Query(entity.QueryCondition{Op: op, Price: price})
			volumeLeft := 0
			for i := range pendingOrder {
				volumeLeft += pendingOrder[i].Volume
				t.Logf("id: %v, volume: %v", pendingOrder[i].OrderID, pendingOrder[i].Volume)
			}
			if volumeLeft != (totalPendingVolume - totalNewVolume) {
				t.Errorf("expected %d got %d", totalPendingVolume-totalNewVolume, volumeLeft)
			}
		} else {
			pendingOrder := repo.Query(entity.QueryCondition{Op: op, Price: price})
			if len(pendingOrder) != 0 {
				t.Errorf("expected 0 got %d", len(pendingOrder))
			}
			pendingOrder = repo.Query(entity.QueryCondition{Op: oppositeOp, Price: price})
			if len(pendingOrder) != 0 {
				t.Errorf("expected 0 got %d", len(pendingOrder))
			}
		}
		t.Cleanup(func() { repo.Delete(entity.QueryCondition{Op: entity.All, Price: price}) })
	})
}

func FuzzMatchManyAsync(f *testing.F) {
	f.Add(100.0)
	// f.Add(2.4)
	// f.Add(1341.0)
	f.Fuzz(func(t *testing.T, price float64) {
		var wg sync.WaitGroup
		var totalBuyVolume, totalSellVolume int32 = 0, 0
		var orderId int32 = 0
		rand.Seed(time.Now().UnixNano())
		userCount := rand.Intn(10) + 1
		t.Logf("buy user count: %d", userCount)
		//buy order
		wg.Add(userCount)
		for i := 0; i < userCount; i++ {
			go func(count int) {
				rand.Seed(time.Now().UnixNano())
				orderCount := rand.Intn(10) + 1
				t.Logf("user %v buy order count: %d", count, orderCount)
				for j := 0; j < orderCount; j++ {
					rand.Seed(time.Now().UnixNano())
					volume := rand.Intn(100) + 1
					atomic.AddInt32(&orderId, 1)
					newOrder := entity.Order{OrderID: strconv.Itoa(int(atomic.LoadInt32(&orderId))), UserID: strconv.Itoa(count), Item: "gold", Op: 0, Volume: volume, Price: price, MatchRule: "partial"}
					t.Logf("user %v new buy order: %+v", count, newOrder)
					err := newOrder.Validate()
					if err != nil {
						return
					}
					atomic.AddInt32(&totalBuyVolume, int32(volume))
					// totalBuyVolume += volume
					result, err := partialMatcher.Match(newOrder)
					t.Logf("result: %v, err: %v", result, err)
				}
				wg.Done()
			}(i + 1)
		}
		//sell order
		rand.Seed(time.Now().UnixNano())
		userCount = rand.Intn(10) + 1
		t.Logf("sell user count: %d", userCount)
		wg.Add(userCount)
		for i := 0; i < userCount; i++ {
			go func(count int) {
				rand.Seed(time.Now().UnixNano())
				orderCount := rand.Intn(10) + 1
				t.Logf("user %v sell order count: %d", count, orderCount)
				for j := 0; j < orderCount; j++ {
					rand.Seed(time.Now().UnixNano())
					volume := rand.Intn(100) + 1
					atomic.AddInt32(&orderId, 1)
					newOrder := entity.Order{OrderID: strconv.Itoa(int(atomic.LoadInt32(&orderId))), UserID: strconv.Itoa(count), Item: "gold", Op: 1, Volume: volume, Price: price, MatchRule: "partial"}
					t.Logf("user %v new sell order: %+v", count, newOrder)
					err := newOrder.Validate()
					if err != nil {
						return
					}
					atomic.AddInt32(&totalSellVolume, int32(volume))
					// totalSellVolume += volume
					result, err := partialMatcher.Match(newOrder)
					t.Logf("result: %v, err: %v", result, err)
				}
				wg.Done()
			}(i + 1)
		}
		wg.Wait()
		t.Logf("totalBuyVolume %v, totalSellVolume %v", totalBuyVolume, totalSellVolume)
		if totalBuyVolume > totalSellVolume {
			pendingOrder := repo.Query(entity.QueryCondition{Op: 0, Price: price})
			volumeLeft := 0
			for i := range pendingOrder {
				volumeLeft += pendingOrder[i].Volume
				t.Logf("buy orderId: %v, volume: %v", pendingOrder[i].OrderID, pendingOrder[i].Volume)
			}
			if volumeLeft != int(totalBuyVolume-totalSellVolume) {
				t.Errorf("expected %d got %d", totalBuyVolume-totalSellVolume, volumeLeft)
			}
		} else if totalBuyVolume < totalSellVolume {
			pendingOrder := repo.Query(entity.QueryCondition{Op: 1, Price: price})
			volumeLeft := 0
			for i := range pendingOrder {
				volumeLeft += pendingOrder[i].Volume
				t.Logf("sell orderId: %v, volume: %v", pendingOrder[i].OrderID, pendingOrder[i].Volume)
			}
			if volumeLeft != int(totalSellVolume-totalBuyVolume) {
				t.Errorf("expected %d got %d", totalSellVolume-totalBuyVolume, volumeLeft)
			}
		} else {
			pendingOrder := repo.Query(entity.QueryCondition{Op: 0, Price: price})
			if len(pendingOrder) != 0 {
				t.Errorf("expected 0 got %d", len(pendingOrder))
			}
			pendingOrder = repo.Query(entity.QueryCondition{Op: 1, Price: price})
			if len(pendingOrder) != 0 {
				t.Errorf("expected 0 got %d", len(pendingOrder))
			}
		}
		t.Cleanup(func() { repo.Delete(entity.QueryCondition{Op: entity.All, Price: price}) })
	})
}
