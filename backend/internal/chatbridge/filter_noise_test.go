package chatbridge

import "testing"

func TestIsPTYNoiseLineForChat_ClaudeFooter(t *testing.T) {
	cases := []string{
		"▶▶ bypass permissions on (shift+tab to cycle) ctrl+g to edit in Vim",
		"bypasspermissionson(shift+tabtocycle)ctrl+gtoeditinVim",
		"▶▶ ▶▶",
	}
	for _, s := range cases {
		if !IsPTYNoiseLineForChat(s) {
			t.Fatalf("expected noise: %q", s)
		}
	}
	if IsPTYNoiseLineForChat("当前时间是 20:43。") {
		t.Fatal("normal prose should not be noise")
	}
}
