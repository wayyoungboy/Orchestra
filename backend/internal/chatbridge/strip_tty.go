package chatbridge

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

// orphanTTY matches CSI-like "[...]" chunks that lost a leading ESC (common when
// terminal UI sync tokens leak into plain text). Requires a digit after "[" or "?digits".
var orphanTTY = regexp.MustCompile(`\[(?:\?[0-9]+[A-Za-z]+|[0-9][\d;]*[A-Za-z])`)

// SanitizeTTYForChat strips ANSI/OSC escape sequences and normalizes whitespace
// so PTY output is safe to show as plain chat text.
func SanitizeTTYForChat(s string) string {
	if s == "" {
		return ""
	}
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); {
		c := s[i]
		if c == '\r' {
			i++
			continue
		}
		if c == '\x1b' || c == '\x9b' { // ESC or C1 CSI
			n := consumeEscape(s, i)
			if n > i {
				i = n
				continue
			}
		}
		r, w := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && w == 1 {
			i++
			continue
		}
		b.WriteRune(r)
		i += w
	}
	out := strings.TrimSpace(b.String())
	// Claude / modern CLIs sometimes emit sync tokens without a leading ESC in logs
	out = stripLooseSyncTokens(out)
	return out
}

func consumeEscape(s string, start int) int {
	if start >= len(s) {
		return start
	}
	if s[start] == '\x1b' {
		if start+1 >= len(s) {
			return start + 1
		}
		switch s[start+1] {
		case '[':
			return consumeCSI(s, start+2)
		case ']':
			return consumeOSC(s, start+2)
		case 'P':
			return consumeUntilString(s, start+2, "\x1b\\")
		case 'X', '^', '_':
			return consumeUntilString(s, start+2, "\x1b\\")
		case '(', ')', '#':
			if start+2 < len(s) {
				return start + 3
			}
			return start + 2
		case 'N', 'O', '\\', 'c', 'E', 'H', 'M', '7', '8', '9', '=', '>':
			return start + 2
		default:
			return start + 2
		}
	}
	// C1 CSI (UTF-8 0x9b) — treat like ESC [
	if s[start] == '\x9b' {
		return consumeCSI(s, start+1)
	}
	return start
}

func consumeCSI(s string, i int) int {
	for i < len(s) {
		c := s[i]
		if c >= 0x40 && c <= 0x7E {
			return i + 1
		}
		i++
	}
	return i
}

func consumeOSC(s string, i int) int {
	for i < len(s) {
		if s[i] == '\x07' {
			return i + 1
		}
		if s[i] == '\x1b' && i+1 < len(s) && s[i+1] == '\\' {
			return i + 2
		}
		i++
	}
	return i
}

func consumeUntilString(s string, i int, end string) int {
	idx := strings.Index(s[i:], end)
	if idx < 0 {
		return len(s)
	}
	return i + idx + len(end)
}

func stripLooseSyncTokens(s string) string {
	if s == "" {
		return s
	}
	// Known bracketed-paste / prompt-sync fragments seen when PTY bytes are split or logged
	replacers := []struct{ old, new string }{
		{"[?2026h", ""},
		{"[?2026l", ""},
	}
	for _, r := range replacers {
		s = strings.ReplaceAll(s, r.old, r.new)
	}
	s = orphanTTY.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}
