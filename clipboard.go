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

// Yank copies text to the clipboard: through the terminal (OSC 52), and
// also with Copy, which works where the terminal refuses. The status line
// then says what was yanked.
func Yank(text, what string) tea.Cmd {
	return yank(text, fmt.Sprintf("yanked %s (%d chars)", what, len([]rune(text))))
}

// YankValue is Yank for a short value, such as a path, which the status
// line then shows.
func YankValue(text string) tea.Cmd { return yank(text, "yanked "+CleanLine(text, 0)) }

func yank(text, status string) tea.Cmd {
	cmds := []tea.Cmd{tea.SetClipboard(text), func() tea.Msg {
		_ = Copy(text) // OSC 52 may have done it
		return nil
	}}
	cmds = append(cmds, func() tea.Msg { return StatusMsg(status) })
	return tea.Batch(cmds...)
}

// Copy puts text on the system clipboard with the platform's tool: pbcopy,
// xclip or xsel, clip.exe.
func Copy(text string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "linux", "freebsd", "openbsd", "netbsd":
		switch {
		case lookPath("wl-copy"):
			cmd = exec.Command("wl-copy")
		case lookPath("xclip"):
			cmd = exec.Command("xclip", "-selection", "clipboard")
		case lookPath("xsel"):
			cmd = exec.Command("xsel", "--clipboard", "--input")
		default:
			return fmt.Errorf("no clipboard tool found (install wl-clipboard, xclip or xsel)")
		}
	case "windows":
		cmd = exec.Command("clip.exe")
	default:
		return fmt.Errorf("no clipboard on %s", runtime.GOOS)
	}
	cmd.Stdin = strings.NewReader(text)
	return cmd.Run()
}

func lookPath(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// Update handles what the shell owns: the spinner and status messages. It
// also takes the window size, but leaves the resize to the app too
// (handled is false), which sizes its own parts.
func (s *Shell[S]) Update(msg tea.Msg) (cmd tea.Cmd, handled bool) {
	if s.Resize(msg) {
		return nil, false
	}
	switch m := msg.(type) {
	case StatusMsg:
		s.SetStatus(string(m), false)
		return nil, true
	case statusExpiredMsg:
		if m.seq == s.statusSeq {
			s.Status, s.StatusErr = "", false
		}
		return nil, true
	}
	return s.Ops.Update(msg)
}
