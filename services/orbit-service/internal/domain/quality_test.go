package domain

import "testing"

func TestPassQualityTier(t *testing.T) {
	if PassQualityTier(45) != "excellent" {
		t.Fatal("expected excellent")
	}
	if PassQualityTier(5) != "marginal" {
		t.Fatal("expected marginal")
	}
}
