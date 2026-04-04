package chatbridge

import (
	"strings"
	"sync"

	"github.com/danielgatis/go-vte"
)

// ScreenBuffer implements vte.Performer to maintain a terminal screen buffer
type ScreenBuffer struct {
	mu     sync.RWMutex
	cols   int
	rows   int
	cursor struct {
		row, col int
	}
	screen   [][]rune
	scrollY  int // scrollback offset
	altScreen bool // alternate screen mode (used by TUI apps)
}

// NewScreenBuffer creates a new terminal screen buffer
func NewScreenBuffer(cols, rows int) *ScreenBuffer {
	sb := &ScreenBuffer{
		cols: cols,
		rows: rows,
	}
	sb.clearScreen()
	return sb
}

func (sb *ScreenBuffer) clearScreen() {
	sb.screen = make([][]rune, sb.rows)
	for i := range sb.screen {
		sb.screen[i] = make([]rune, sb.cols)
		for j := range sb.screen[i] {
			sb.screen[i][j] = ' '
		}
	}
	sb.cursor.row = 0
	sb.cursor.col = 0
}

// Print implements vte.Performer.Print
func (sb *ScreenBuffer) Print(r rune) {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	if sb.cursor.row < 0 || sb.cursor.row >= sb.rows {
		return
	}
	if sb.cursor.col < 0 || sb.cursor.col >= sb.cols {
		return
	}

	sb.screen[sb.cursor.row][sb.cursor.col] = r
	sb.cursor.col++
	if sb.cursor.col >= sb.cols {
		sb.cursor.col = sb.cols - 1
	}
}

// Execute implements vte.Performer.Execute
func (sb *ScreenBuffer) Execute(b byte) {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	switch b {
	case '\n', '\v', '\f':
		sb.cursor.row++
		sb.cursor.col = 0
		if sb.cursor.row >= sb.rows {
			sb.scrollUp()
		}
	case '\r':
		sb.cursor.col = 0
	case '\t':
		sb.cursor.col = ((sb.cursor.col / 8) + 1) * 8
		if sb.cursor.col >= sb.cols {
			sb.cursor.col = sb.cols - 1
		}
	case '\b':
		if sb.cursor.col > 0 {
			sb.cursor.col--
		}
	}
}

// Put implements vte.Performer.Put
func (sb *ScreenBuffer) Put(b byte) {
	// DCS hook data - ignore
}

// Unhook implements vte.Performer.Unhook
func (sb *ScreenBuffer) Unhook() {
	// DCS unhook - ignore
}

// Hook implements vte.Performer.Hook
func (sb *ScreenBuffer) Hook(params [][]uint16, intermediates []byte, ignore bool, r rune) {
	// DCS hook - ignore
}

// OscDispatch implements vte.Performer.OscDispatch
func (sb *ScreenBuffer) OscDispatch(params [][]byte, bellTerminated bool) {
	// OSC sequences (title, etc) - ignore for chat
}

// CsiDispatch implements vte.Performer.CsiDispatch
func (sb *ScreenBuffer) CsiDispatch(params [][]uint16, intermediates []byte, ignore bool, r rune) {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	// Parse common CSI sequences
	switch r {
	case 'A': // CUU - cursor up
		n := paramValue(params, 0, 1)
		sb.cursor.row -= n
		if sb.cursor.row < 0 {
			sb.cursor.row = 0
		}
	case 'B': // CUD - cursor down
		n := paramValue(params, 0, 1)
		sb.cursor.row += n
		if sb.cursor.row >= sb.rows {
			sb.cursor.row = sb.rows - 1
		}
	case 'C': // CUF - cursor forward
		n := paramValue(params, 0, 1)
		sb.cursor.col += n
		if sb.cursor.col >= sb.cols {
			sb.cursor.col = sb.cols - 1
		}
	case 'D': // CUB - cursor back
		n := paramValue(params, 0, 1)
		sb.cursor.col -= n
		if sb.cursor.col < 0 {
			sb.cursor.col = 0
		}
	case 'E': // CNL - cursor next line
		n := paramValue(params, 0, 1)
		sb.cursor.row += n
		sb.cursor.col = 0
		if sb.cursor.row >= sb.rows {
			sb.cursor.row = sb.rows - 1
		}
	case 'F': // CPL - cursor previous line
		n := paramValue(params, 0, 1)
		sb.cursor.row -= n
		sb.cursor.col = 0
		if sb.cursor.row < 0 {
			sb.cursor.row = 0
		}
	case 'G': // CHA - cursor horizontal absolute
		n := paramValue(params, 0, 1) - 1
		if n < 0 {
			n = 0
		}
		if n >= sb.cols {
			n = sb.cols - 1
		}
		sb.cursor.col = n
	case 'H', 'f': // CUP - cursor position
		row := paramValue(params, 0, 1) - 1
		col := paramValue(params, 1, 1) - 1
		if row < 0 {
			row = 0
		}
		if col < 0 {
			col = 0
		}
		if row >= sb.rows {
			row = sb.rows - 1
		}
		if col >= sb.cols {
			col = sb.cols - 1
		}
		sb.cursor.row = row
		sb.cursor.col = col
	case 'J': // ED - erase in display
		n := paramValue(params, 0, 0)
		switch n {
		case 0: // erase from cursor to end
			for j := sb.cursor.col; j < sb.cols; j++ {
				sb.screen[sb.cursor.row][j] = ' '
			}
			for i := sb.cursor.row + 1; i < sb.rows; i++ {
				for j := 0; j < sb.cols; j++ {
					sb.screen[i][j] = ' '
				}
			}
		case 1: // erase from start to cursor
			for i := 0; i < sb.cursor.row; i++ {
				for j := 0; j < sb.cols; j++ {
					sb.screen[i][j] = ' '
				}
			}
			for j := 0; j <= sb.cursor.col; j++ {
				sb.screen[sb.cursor.row][j] = ' '
			}
		case 2, 3: // erase entire screen
			sb.clearScreen()
		}
	case 'K': // EL - erase in line
		n := paramValue(params, 0, 0)
		switch n {
		case 0: // erase from cursor to end
			for j := sb.cursor.col; j < sb.cols; j++ {
				sb.screen[sb.cursor.row][j] = ' '
			}
		case 1: // erase from start to cursor
			for j := 0; j <= sb.cursor.col; j++ {
				sb.screen[sb.cursor.row][j] = ' '
			}
		case 2: // erase entire line
			for j := 0; j < sb.cols; j++ {
				sb.screen[sb.cursor.row][j] = ' '
			}
		}
	case 'L': // IL - insert lines
		n := paramValue(params, 0, 1)
		for i := sb.rows - 1; i >= sb.cursor.row+n; i-- {
			sb.screen[i] = sb.screen[i-n]
		}
		for i := sb.cursor.row; i < sb.cursor.row+n && i < sb.rows; i++ {
			sb.screen[i] = make([]rune, sb.cols)
			for j := range sb.screen[i] {
				sb.screen[i][j] = ' '
			}
		}
	case 'M': // DL - delete lines
		n := paramValue(params, 0, 1)
		for i := sb.cursor.row; i < sb.rows-n; i++ {
			sb.screen[i] = sb.screen[i+n]
		}
		for i := sb.rows - n; i < sb.rows; i++ {
			sb.screen[i] = make([]rune, sb.cols)
			for j := range sb.screen[i] {
				sb.screen[i][j] = ' '
			}
		}
	case 'P': // DCH - delete characters
		n := paramValue(params, 0, 1)
		for j := sb.cursor.col; j < sb.cols-n; j++ {
			sb.screen[sb.cursor.row][j] = sb.screen[sb.cursor.row][j+n]
		}
		for j := sb.cols - n; j < sb.cols; j++ {
			sb.screen[sb.cursor.row][j] = ' '
		}
	case '@': // ICH - insert characters
		n := paramValue(params, 0, 1)
		for j := sb.cols - 1; j >= sb.cursor.col+n; j-- {
			sb.screen[sb.cursor.row][j] = sb.screen[sb.cursor.row][j-n]
		}
		for j := sb.cursor.col; j < sb.cursor.col+n && j < sb.cols; j++ {
			sb.screen[sb.cursor.row][j] = ' '
		}
	case 'S': // SU - scroll up
		n := paramValue(params, 0, 1)
		for i := 0; i < n; i++ {
			sb.scrollUp()
		}
	case 'T': // SD - scroll down
		n := paramValue(params, 0, 1)
		for i := 0; i < n; i++ {
			sb.scrollDown()
		}
	case 'd': // VPA - vertical position absolute
		n := paramValue(params, 0, 1) - 1
		if n < 0 {
			n = 0
		}
		if n >= sb.rows {
			n = sb.rows - 1
		}
		sb.cursor.row = n
	case 'm': // SGR - select graphic rendition
		// Ignore styling for chat output
	case 'h', 'l': // SM/RM - set/reset mode
		// Handle alternate screen buffer
		if len(params) > 0 && len(params[0]) > 0 {
			if params[0][0] == 1049 || params[0][0] == 47 || params[0][0] == 1047 {
				sb.altScreen = (r == 'h')
				if sb.altScreen {
					sb.clearScreen()
				}
			}
		}
	case 's': // SCP - save cursor position
		// TODO: implement if needed
	case 'u': // RCP - restore cursor position
		// TODO: implement if needed
	}
}

// EscDispatch implements vte.Performer.EscDispatch
func (sb *ScreenBuffer) EscDispatch(intermediates []byte, ignore bool, b byte) {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	switch b {
	case 'c': // RIS - reset to initial state
		sb.clearScreen()
	case 'D': // IND - index (move down)
		sb.cursor.row++
		if sb.cursor.row >= sb.rows {
			sb.scrollUp()
		}
	case 'E': // NEL - next line
		sb.cursor.row++
		sb.cursor.col = 0
		if sb.cursor.row >= sb.rows {
			sb.scrollUp()
		}
	case 'M': // RI - reverse index (move up)
		sb.cursor.row--
		if sb.cursor.row < 0 {
			sb.cursor.row = 0
			sb.scrollDown()
		}
	}
}

// SosPmApcDispatch implements vte.Performer.SosPmApcDispatch
func (sb *ScreenBuffer) SosPmApcDispatch(kind vte.SosPmApcKind, data []byte, bellTerminated bool) {
	// Ignore SOS/PM/APC sequences
}

func (sb *ScreenBuffer) scrollUp() {
	for i := 0; i < sb.rows-1; i++ {
		sb.screen[i] = sb.screen[i+1]
	}
	sb.screen[sb.rows-1] = make([]rune, sb.cols)
	for j := range sb.screen[sb.rows-1] {
		sb.screen[sb.rows-1][j] = ' '
	}
}

func (sb *ScreenBuffer) scrollDown() {
	for i := sb.rows - 1; i > 0; i-- {
		sb.screen[i] = sb.screen[i-1]
	}
	sb.screen[0] = make([]rune, sb.cols)
	for j := range sb.screen[0] {
		sb.screen[0][j] = ' '
	}
}

func paramValue(params [][]uint16, index int, def int) int {
	if index >= len(params) || len(params[index]) == 0 {
		return def
	}
	return int(params[index][0])
}

// GetVisibleContent returns the current screen content as a string
func (sb *ScreenBuffer) GetVisibleContent() string {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	var lines []string
	for i := 0; i < sb.rows; i++ {
		line := strings.TrimRight(string(sb.screen[i]), " ")
		if line != "" {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, "\n")
}

// GetLine returns a specific line (1-indexed)
func (sb *ScreenBuffer) GetLine(row int) string {
	sb.mu.RLock()
	defer sb.mu.RUnlock()

	if row < 1 || row > sb.rows {
		return ""
	}
	return strings.TrimRight(string(sb.screen[row-1]), " ")
}

// Resize changes the screen dimensions
func (sb *ScreenBuffer) Resize(cols, rows int) {
	sb.mu.Lock()
	defer sb.mu.Unlock()

	newScreen := make([][]rune, rows)
	for i := range newScreen {
		newScreen[i] = make([]rune, cols)
		for j := range newScreen[i] {
			if i < sb.rows && j < sb.cols {
				newScreen[i][j] = sb.screen[i][j]
			} else {
				newScreen[i][j] = ' '
			}
		}
	}
	sb.screen = newScreen
	sb.cols = cols
	sb.rows = rows
	if sb.cursor.row >= rows {
		sb.cursor.row = rows - 1
	}
	if sb.cursor.col >= cols {
		sb.cursor.col = cols - 1
	}
}