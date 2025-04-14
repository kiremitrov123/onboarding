package cache

import (
	"testing"
	"time"

	"github.com/kiremitrov123/onboarding/stockprice/model"
)

func TestLocalCacheRacePrevention(t *testing.T) {
	lc := NewLocalCache(5*time.Minute, 10*time.Minute)

	// Set placeholder before fetching
	lc.SetPlaceholder("AAPL")

	// Simulate a `late response` scenario by setting actual data before invalidation
	lc.Set("AAPL", mockPrice(123.45))

	// Now simulate invalidation arriving after the stale set
	lc.Invalidate("AAPL")

	// The value should have been skipped OR cleaned up by invalidation
	if _, found := lc.Get("AAPL"); found {
		t.Error("Expected cache to skip or invalidate stale value after race")
	}

	// Test that a valid fetch scenario works
	lc.SetPlaceholder("AAPL")
	lc.Set("AAPL", mockPrice(133.00)) // no invalidation
	price, found := lc.Get("AAPL")
	if !found {
		t.Error("Expected valid price to be cached")
		return
	}
	if price.Price != 133.00 {
		t.Errorf("Expected 133.00, got %.2f", price.Price)
	}
}

func mockPrice(val float64) model.Price {
	return model.Price{
		Symbol:    "AAPL",
		Price:     val,
		Timestamp: time.Now(),
	}
}
