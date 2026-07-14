package models

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMemberDoesNotSerializeA2AAuthToken(t *testing.T) {
	token := "sensitive-a2a-token"
	data, err := json.Marshal(Member{ID: "member-1", A2AAuthToken: &token})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	if strings.Contains(string(data), token) || strings.Contains(string(data), "a2aAuthToken") {
		t.Fatalf("serialized member exposed A2A credentials: %s", data)
	}
}
