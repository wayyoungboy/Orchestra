package repository

import (
	"testing"
)

func TestBatchGetUnreadCounts_Empty(t *testing.T) {
	r := &ConversationReadRepository{}
	result, err := r.BatchGetUnreadCounts([]string{}, "member-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty map, got %d entries", len(result))
	}
}

func TestJoinPlaceholders(t *testing.T) {
	tests := []struct {
		n      int
		expect string
	}{
		{1, "?"},
		{2, "?,?"},
		{3, "?,?,?"},
		{5, "?,?,?,?,?"},
	}
	for _, tc := range tests {
		got := joinPlaceholders(tc.n)
		if got != tc.expect {
			t.Errorf("joinPlaceholders(%d) = %q, want %q", tc.n, got, tc.expect)
		}
	}
}
