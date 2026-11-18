package tuikit

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/Gandalfalex/vim-kit/cmdline"
)

// Chrome is the lines the frame takes besides the body: title, status and
// hints.
const Chrome = 3

// Frame stacks the title bar, the body (which ends in a newline), the
// status line and the hints.
func Frame(title, body, status, hints string) string {
	return title + "\n" + body + status + "\n" + hints
}

// TitleBar is the app name followed by where the user is.
func TitleBar(th Theme, w int, app string, parts ...string) string {
	all := append([]string{th.Title.Render(app)}, parts...)
	return ansi.Truncate(strings.Join(all, th.Muted.Render(" › ")), w, "…")
}

// TitleBarPath is TitleBar with a path among the parts (before, path,
// after) that, when the bar is too wide, is cut from the left: its deepest
// part is the one that matters.
func TitleBarPath(th Theme, w int, app string, before []string, path string, after []string) string {
	parts := func(p string) []string {
		out := append([]string(nil), before...)
		if p != "" {
			out = append(out, th.Crumb.Render(p))
		}
		return append(out, after...)
	}
	s := TitleBar(th, 1<<30, app, parts(path)...)
	if over := lipgloss.Width(s) - w; over > 0 && path != "" {
		s = TitleBar(th, 1<<30, app, parts(TruncLeft(path, max(8, lipgloss.Width(path)-over)))...)
	}
	return ansi.Truncate(s, w, "…")
}

// StatusBar is the mode badge and parts separated by bars.
func StatusBar(th Theme, w int, mode Mode, parts []string) string {
	line := th.Mode.Background(mode.Color).Render(mode.Name)
	var kept []string
	for _, p := range parts {
		if p != "" {
			kept = append(kept, p)
		}
	}
	if len(kept) > 0 {
		line += " " + strings.Join(kept, th.Muted.Render(" │ "))
	}
	return th.Bar.Render(Pad(ansi.Truncate(line, w, "…"), w))
}

// PromptBar is the status line while a prompt is open.
func PromptBar(th Theme, w int, mode Mode, in textinput.Model) string {
	left := th.Mode.Background(mode.Color).Render(mode.Name) + " "
	return left + ansi.Truncate(in.View(), w-lipgloss.Width(left), "")
}

// HintBar is the line of key hints at the bottom.
func HintBar(th Theme, w int, hints string) string {
	return th.Muted.Render(ansi.Truncate(" "+hints, w, "…"))
}

// StatusParts adds what every screen shows to the app's parts: the
// operation in flight first, then the error count (unless errorsShown),
// a pending key sequence and the status message.
func (s *Shell[S]) StatusParts(parts []string, errorsShown bool) []string {
	out := []string{s.Ops.Spinner()}
	out = append(out, parts...)
	if n := s.Errors.Len(); n > 0 && !errorsShown {
		out = append(out, s.Theme.ErrText.Render(fmt.Sprintf("E:%d", n)))
	}
	if p := s.Keys.Partial(); p != "" {
		out = append(out, s.Theme.Warn.Render(p))
	}
	if s.Status != "" {
		st := s.Status
		if s.StatusErr {
			st = s.Theme.ErrText.Render(st)
		}
		out = append(out, st)
	}
	return out
}

// Popup draws a completion menu of at most rows rows, scrolled so the
// selection is visible.
func Popup(th Theme, line *cmdline.Line, w, rows int) []string {
	items, sel := line.Items, line.Index
	n := len(items)
	if n == 0 {
		return nil
	}
	win := min(n, rows)
	if n > rows {
		win-- // the last row says how many more
	}
	start := 0
	if sel >= win {
		start = sel - win + 1
	}
	labelW := 0
	for _, it := range items {
		labelW = max(labelW, ansi.StringWidth(it.Label))
	}
	labelW = min(labelW+2, max(10, w/2))
	var out []string
	for i := start; i < start+win; i++ {
		it := items[i]
		row := Pad(" "+Pad(it.Label, labelW)+th.Muted.Render(it.Desc), w)
		if i == sel {
			row = th.Sel.Render(ansi.Strip(row))
		} else {
			row = th.Bar.Render(row)
		}
		out = append(out, row)
	}
	if rest := n - start - win; n > rows {
		out = append(out, th.Bar.Render(Pad(th.Muted.Render(fmt.Sprintf(" %d/%d · %d more (tab to cycle)", max(sel+1, 0), n, rest)), w)))
	}
	return out
}

// Overlay draws popup over the bottom of body (a frame's body, ending in
// a newline), or over its top with atTop, like vim's popup menu.
func Overlay(body string, popup []string, atTop bool) string {
	if len(popup) == 0 {
		return body
	}
	lines := strings.Split(strings.TrimSuffix(body, "\n"), "\n")
	k := min(len(popup), len(lines))
	if atTop {
		copy(lines, popup[:k])
	} else {
		copy(lines[len(lines)-k:], popup[len(popup)-k:])
	}
	return strings.Join(lines, "\n") + "\n"
}
