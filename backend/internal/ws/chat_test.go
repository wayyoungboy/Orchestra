package ws

import "testing"

func TestGenerateClientIDUsesUniqueIDs(t *testing.T) {
	ids := make(map[string]struct{}, 128)
	for i := 0; i < 128; i++ {
		id := generateClientID()
		if _, exists := ids[id]; exists {
			t.Fatalf("generateClientID() generated duplicate id %q", id)
		}
		ids[id] = struct{}{}
	}
}
