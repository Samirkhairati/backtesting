package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/samirkhairati/backtesting/core/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockStore struct {
	mock.Mock
}

func (m *mockStore) GetBars(ctx context.Context, symbol string) (data.BarIterator, error) {
	args := m.Called(ctx, symbol)
	return args.Get(0).(data.BarIterator), args.Error(1)
}

func (m *mockStore) GetMultiBars(ctx context.Context, symbols []string) (data.MultiBarIterator, error) {
	args := m.Called(ctx, symbols)
	return args.Get(0).(data.MultiBarIterator), args.Error(1)
}

func (m *mockStore) GetMultiBarsWithTimeframe(ctx context.Context, symbols []string, start, end string) (data.MultiBarIterator, error) {
	args := m.Called(ctx, symbols, start, end)
	return args.Get(0).(data.MultiBarIterator), args.Error(1)
}

func TestRedisStore_GetMultiBarsWithTimeframe(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	mockDB := new(mockStore)
	store := NewRedisStore(client, mockDB)

	ctx := context.Background()

	// Mock data
	ts1, _ := time.Parse("2006-01-02", "2023-01-01")
	ts2, _ := time.Parse("2006-01-02", "2023-01-02")
	ts3, _ := time.Parse("2006-01-02", "2023-01-03")

	allBars := []data.Bar{
		{Symbol: "AAPL", Timestamp: ts1, Close: 150},
		{Symbol: "AAPL", Timestamp: ts2, Close: 155},
		{Symbol: "AAPL", Timestamp: ts3, Close: 160},
	}
	memIter := data.NewMemoryBarIterator(allBars)

	// First call should hit the mock DB (cache miss)
	mockDB.On("GetBars", ctx, "AAPL").Return(memIter, nil).Once()

	// Ask for only ts2
	iter, err := store.GetMultiBarsWithTimeframe(ctx, []string{"AAPL"}, "2023-01-02", "2023-01-02")
	assert.NoError(t, err)

	// Assert only ts2 is returned
	assert.True(t, iter.Next())
	bars := iter.Bars()
	assert.Equal(t, 1, len(bars))
	assert.Equal(t, float64(155), bars["AAPL"].Close)
	assert.Equal(t, ts2, iter.Timestamp())

	assert.False(t, iter.Next())
	assert.NoError(t, iter.Err())

	// Verify mock was called
	mockDB.AssertExpectations(t)

	// Second call should NOT hit the mock DB (cache hit)
	// Even though we ask for a different timeframe
	iter2, err := store.GetMultiBarsWithTimeframe(ctx, []string{"AAPL"}, "2023-01-01", "2023-01-03")
	assert.NoError(t, err)

	// Assert all 3 returned
	assert.True(t, iter2.Next())
	assert.Equal(t, float64(150), iter2.Bars()["AAPL"].Close)

	assert.True(t, iter2.Next())
	assert.Equal(t, float64(155), iter2.Bars()["AAPL"].Close)

	assert.True(t, iter2.Next())
	assert.Equal(t, float64(160), iter2.Bars()["AAPL"].Close)

	assert.False(t, iter2.Next())

	// mockDB should NOT have been called again (the `.Once()` expectation on the mock guarantees it would fail if called)
	mockDB.AssertExpectations(t)
}
