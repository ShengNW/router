package model

import "testing"

func TestResolveBalanceCreditExpiresAt(t *testing.T) {
	base := int64(1776073942) // 2026-04-13 15:12:22 UTC+8

	if got := ResolveBalanceCreditExpiresAt(base, 0); got != 0 {
		t.Fatalf("validity=0 should never expire, got %d", got)
	}

	wantThreeDays := base + 3*86400
	if got := ResolveBalanceCreditExpiresAt(base, 3); got != wantThreeDays {
		t.Fatalf("validity=3 expected %d, got %d", wantThreeDays, got)
	}
}
