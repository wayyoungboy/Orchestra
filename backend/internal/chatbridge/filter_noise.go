package chatbridge

import "strings"

// IsPTYNoiseLineForChat reports lines that are almost certainly TUI chrome (status bar,
// key hints) rather than assistant prose — common with Claude Code and similar full-screen CLIs.
func IsPTYNoiseLineForChat(line string) bool {
	if line == "" {
		return false
	}
	lower := strings.ToLower(line)
	// Claude Code footer / permission strip
	if strings.Contains(lower, "bypass") && strings.Contains(lower, "permission") {
		return true
	}
	// Footer hints (often concatenated when CSI/columns are stripped)
	if strings.Contains(lower, "shift+tab") || (strings.Contains(lower, "shift") && strings.Contains(lower, "tab") && strings.Contains(lower, "cycle")) {
		return true
	}
	if strings.Contains(lower, "ctrl+g") || (strings.Contains(lower, "ctrl") && strings.Contains(lower, "edit") && strings.Contains(lower, "vim")) {
		return true
	}
	// Spinner / mode glyphs with no real words
	if strings.Count(line, "▶") >= 2 && len([]rune(line)) < 120 {
		return true
	}
	return false
}

// IsPTYNoiseStreamChunk drops flushed stream fragments that are clearly status-bar residue.
func IsPTYNoiseStreamChunk(chunk string) bool {
	if chunk == "" {
		return false
	}
	// Short chunks: only strip if entire thing looks like chrome
	runes := []rune(strings.TrimSpace(chunk))
	if len(runes) <= 200 {
		return IsPTYNoiseLineForChat(chunk)
	}
	lower := strings.ToLower(chunk)
	if strings.Contains(lower, "bypass") && strings.Contains(lower, "permission") {
		return true
	}
	return false
}

// NormalizeTTYLineForEchoCompare collapses prompt/whitespace so PTY echo can match injected chat text.
func NormalizeTTYLineForEchoCompare(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, ">")
	s = strings.TrimSpace(s)
	return strings.Join(strings.Fields(s), " ")
}
