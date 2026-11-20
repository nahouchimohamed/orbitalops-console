package domain

import "testing"

func TestWrapLonDeg(t *testing.T) {
	if got := WrapLonDeg(190); got < -170.1 || got > -169.9 {
		t.Fatalf("got %v", got)
	}
}

func TestClampLatDeg(t *testing.T) {
	if ClampLatDeg(120) != 90 {
		t.Fatal("expected clamp")
	}
}
