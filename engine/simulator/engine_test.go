package simulator_test

import (
	"testing"
	"time"

	"github.com/samirkhairati/backtesting/engine/broker"
	"github.com/samirkhairati/backtesting/engine/data"
	"github.com/samirkhairati/backtesting/engine/simulator"
)

type mockIterator struct {
	data []map[string]data.Bar
	idx  int
}

func (m *mockIterator) Next() bool {
	if m.idx >= len(m.data) {
		return false
	}
	m.idx++
	return true
}

func (m *mockIterator) Bars() map[string]data.Bar {
	return m.data[m.idx-1]
}

func (m *mockIterator) Timestamp() time.Time {
	// pick any bar's timestamp, they should all be identical
	for _, b := range m.data[m.idx-1] {
		return b.Timestamp
	}
	return time.Time{}
}

func (m *mockIterator) Err() error {
	return nil
}

type mockStrategy struct {
	receivedBars []map[string]data.Bar
}

func (s *mockStrategy) OnBars(b broker.Broker, bars map[string]data.Bar) {
	s.receivedBars = append(s.receivedBars, bars)
}

func TestEngine_Run(t *testing.T) {

	t1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)

	data := []map[string]data.Bar{
		{
			"AAPL": {Symbol: "AAPL", Timestamp: t1, Close: 150.0},
			"MSFT": {Symbol: "MSFT", Timestamp: t1, Close: 250.0},
		},
		{
			"AAPL": {Symbol: "AAPL", Timestamp: t2, Close: 152.0},
			"MSFT": {Symbol: "MSFT", Timestamp: t2, Close: 248.0},
		},
	}

	iterator := &mockIterator{data: data}
	strat := &mockStrategy{}

	engine := simulator.NewEngine(iterator, strat)

	err := engine.Run()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(strat.receivedBars) != 2 {
		t.Fatalf("Expected strategy to receive 2 iterations, got %d", len(strat.receivedBars))
	}

	if strat.receivedBars[0]["AAPL"].Timestamp != t1 || strat.receivedBars[1]["AAPL"].Timestamp != t2 {
		t.Errorf("Bars were not received in the correct chronological order")
	}

	if engine.CurrentTime != t2 {
		t.Errorf("Expected engine clock to be %v, got %v", t2, engine.CurrentTime)
	}
}

