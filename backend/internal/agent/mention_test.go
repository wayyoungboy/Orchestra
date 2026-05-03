package agent

import (
	"testing"

	"github.com/orchestra/backend/internal/models"
)

func TestParseMentions_SingleMention(t *testing.T) {
	members := []models.Member{
		{ID: "alice_id", Name: "alice"},
		{ID: "bob_id", Name: "bob"},
	}
	result := ParseMentions("@alice 帮忙", members)
	if len(result) != 1 || result[0] != "alice_id" {
		t.Errorf("expected [alice_id], got %v", result)
	}
}

func TestParseMentions_MultipleMentions(t *testing.T) {
	members := []models.Member{
		{ID: "alice_id", Name: "alice"},
		{ID: "bob_id", Name: "bob"},
	}
	result := ParseMentions("@alice @bob 看看", members)
	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}
	if result[0] != "alice_id" || result[1] != "bob_id" {
		t.Errorf("expected [alice_id, bob_id], got %v", result)
	}
}

func TestParseMentions_AllMention(t *testing.T) {
	members := []models.Member{
		{ID: "a_id", Name: "alice"},
		{ID: "b_id", Name: "bob"},
		{ID: "c_id", Name: "carol"},
	}
	result := ParseMentions("@all 通知", members)
	if len(result) != 3 {
		t.Errorf("expected 3 results for @all, got %d", len(result))
	}
}

func TestParseMentions_NoMention(t *testing.T) {
	members := []models.Member{
		{ID: "alice_id", Name: "alice"},
	}
	result := ParseMentions("普通消息", members)
	if len(result) != 0 {
		t.Errorf("expected empty, got %v", result)
	}
}

func TestParseMentions_UnknownMention(t *testing.T) {
	members := []models.Member{
		{ID: "alice_id", Name: "alice"},
	}
	result := ParseMentions("@unknown 帮忙", members)
	if len(result) != 0 {
		t.Errorf("expected empty for unknown mention, got %v", result)
	}
}

func TestParseMentions_CaseInsensitive(t *testing.T) {
	members := []models.Member{
		{ID: "alice_id", Name: "Alice"},
	}
	result := ParseMentions("@alice 帮忙", members)
	if len(result) != 1 || result[0] != "alice_id" {
		t.Errorf("expected case-insensitive match, got %v", result)
	}
}

func TestParseMentions_Dedup(t *testing.T) {
	members := []models.Member{
		{ID: "alice_id", Name: "alice"},
	}
	result := ParseMentions("@alice @alice 重复", members)
	if len(result) != 1 {
		t.Errorf("expected dedup, got %d results", len(result))
	}
}
