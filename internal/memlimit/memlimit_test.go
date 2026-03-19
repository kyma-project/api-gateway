package memlimit

import (
	"runtime/debug"
	"testing"

	"github.com/go-logr/logr"
)

func TestSetGoMemLimitFromCgroup_InvalidPct(t *testing.T) {
	tests := []struct {
		name string
		pct  float64
	}{
		{"zero", 0},
		{"negative", -0.5},
		{"above one", 1.1},
	}
	log := logr.Logger{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := SetGoMemLimitFromCgroup(tt.pct, log); err == nil {
				t.Error("expected error for invalid percentage")
			}
		})
	}
}

func TestParseCgroup_ValidValue(t *testing.T) {
	val, err := parseCgroup("1073741824\n")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 1073741824 {
		t.Fatalf("expected 1073741824, got %d", val)
	}
}

func TestParseCgroup_Max(t *testing.T) {
	_, err := parseCgroup("max\n")
	if err == nil {
		t.Error("expected error for 'max' value")
	}
}

func TestParseCgroup_InvalidValue(t *testing.T) {
	_, err := parseCgroup("notanumber\n")
	if err == nil {
		t.Error("expected error for non-numeric value")
	}
}

func TestSetMemoryLimit_Integration(t *testing.T) {
	prev := debug.SetMemoryLimit(-1)
	debug.SetMemoryLimit(prev)
}
