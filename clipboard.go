package tuikit

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// StatusMsg sets the status line when it arrives; Shell.Update handles it.
type StatusMsg string

// Yank copies text to the clipboard: through the terminal (OSC 52), and on
// macOS also through pbcopy, which works where the terminal refuses. The
// status line then says what was yanked.
func Yank(text, what string) tea.Cmd {
	cmds := []tea.Cmd{tea.SetClipboard(text)}
	if runtime.GOOS == "darwin" {
		cmds = append(cmds, func() tea.Msg {
			c := exec.Command("pbcopy")
			c.Stdin = strings.NewReader(text)
			_ = c.Run()
			return nil
		})
	}
	cmds = append(cmds, func() tea.Msg {
		return StatusMsg(fmt.Sprintf("yanked %s (%d chars)", what, len([]rune(text))))
	})
	return tea.Batch(cmds...)
}

// Update handles what the shell owns: the spinner and status messages. It
// also takes the window size, but leaves the resize to the app too
// (handled is false), which sizes its own parts.
func (s *Shell[S]) Update(msg tea.Msg) (cmd tea.Cmd, handled bool) {
	if s.Resize(msg) {
		return nil, false
	}
	if m, ok := msg.(StatusMsg); ok {
		s.SetStatus(string(m), false)
		return nil, true
	}
	return s.Ops.Update(msg)
}
