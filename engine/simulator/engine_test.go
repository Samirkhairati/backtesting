package simulator_test

import (
	"testing"
	"time"

	"github.com/samirkhairati/backtesting/engine/data"
	"github.com/samirkhairati/backtesting/engine/simulator"
)

type mockIterator struct {
	bars []data.Bar
	idx  int
}

func (m *mockIterator) Next() bool {
	if m.idx >= len(m.bars) {
		return false
	}
	m.idx++
	return true
}

func (m *mockIterator) Bar() data.Bar {
	return m.bars[m.idx-1]
}

func (m *mockIterator) Err() error {
	return nil
}

type mockStrategy struct {
	receivedBars []data.Bar
}

func (s *mockStrategy) OnBar(bar data.Bar) {
	s.receivedBars = append(s.receivedBars, bar)
}

func TestEngine_Run(t *testing.T) {

	t1 := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)

	bars := []data.Bar{
		{Symbol: "AAPL", Timestamp: t1, Close: 150.0},
		{Symbol: "AAPL", Timestamp: t2, Close: 152.0},
	}

	iterator := &mockIterator{bars: bars}
	strat := &mockStrategy{}

	engine := simulator.NewEngine(iterator, strat)

	err := engine.Run()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(strat.receivedBars) != 2 {
		t.Fatalf("Expected strategy to receive 2 bars, got %d", len(strat.receivedBars))
	}

	if strat.receivedBars[0].Timestamp != t1 || strat.receivedBars[1].Timestamp != t2 {
		t.Errorf("Bars were not received in the correct chronological order")
	}

	if engine.CurrentTime != t2 {
		t.Errorf("Expected engine clock to be %v, got %v", t2, engine.CurrentTime)
	}
}
