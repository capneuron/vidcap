package timestamps_test

import (
	"testing"

	"vidcap/internal/timestamps"
)

func TestCalcEvery(t *testing.T) {
	ts, err := timestamps.CalcEvery(100.0, 30.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// expect 30, 60, 90
	want := []float64{30.0, 60.0, 90.0}
	if len(ts) != len(want) {
		t.Fatalf("got %d timestamps, want %d: %v", len(ts), len(want), ts)
	}
	for i, v := range ts {
		if v != want[i] {
			t.Errorf("ts[%d] = %.3f, want %.3f", i, v, want[i])
		}
	}
}

func TestCalcEvery_ExactBoundary(t *testing.T) {
	// interval divides duration exactly — last point should be excluded
	ts, err := timestamps.CalcEvery(90.0, 30.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 30, 60 — 90 equals duration, excluded
	if len(ts) != 2 {
		t.Fatalf("got %d timestamps, want 2: %v", len(ts), ts)
	}
}

func TestCalcEvery_InvalidInterval(t *testing.T) {
	_, err := timestamps.CalcEvery(100.0, 0)
	if err == nil {
		t.Fatal("expected error for zero interval")
	}
}

func TestCalcCount(t *testing.T) {
	ts, err := timestamps.CalcCount(100.0, 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// duration=100, count=3: 100*1/4=25, 100*2/4=50, 100*3/4=75
	want := []float64{25.0, 50.0, 75.0}
	if len(ts) != len(want) {
		t.Fatalf("got %d timestamps, want %d: %v", len(ts), len(want), ts)
	}
	for i, v := range ts {
		if v != want[i] {
			t.Errorf("ts[%d] = %.3f, want %.3f", i, v, want[i])
		}
	}
}

func TestCalcCount_InvalidCount(t *testing.T) {
	_, err := timestamps.CalcCount(100.0, 0)
	if err == nil {
		t.Fatal("expected error for zero count")
	}
}
