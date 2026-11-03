package domain

import (
	"testing"
	"time"
)

func TestWindowsOverlap(t *testing.T) {
	a0 := time.Unix(0, 0)
	a1 := time.Unix(100, 0)
	b0 := time.Unix(50, 0)
	b1 := time.Unix(150, 0)
	if !WindowsOverlap(a0, a1, b0, b1) {
		t.Fatal("expected overlap")
	}
	if WindowsOverlap(a0, a1, a1, time.Unix(200, 0)) {
		t.Fatal("touching should not overlap")
	}
}
