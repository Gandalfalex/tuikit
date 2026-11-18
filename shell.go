package tuikit

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/Gandalfalex/vim-kit/keys"
)

// Shell is what every screen of an app shares. Embed it in the app model.
type Shell[S comparable] struct {
	Theme  Theme
	W, H   int
	Stack  Stack[S]
	Ops    Ops
	Errors ErrorLog
	Keys   keys.Input // count prefix and pending g/z sequence

	Status    string
	StatusErr bool
	// Brief turns an error into its status line form; nil is ShortErr.
	Brief func(error) string

	quitArmed bool
}

func NewShell[S comparable](th Theme, root S) Shell[S] {
	return Shell[S]{Theme: th, Stack: Stack[S]{items: []S{root}}, Ops: NewOps(th), Errors: ErrorLog{Max: 200}}
}

// Resize takes the window size; it reports whether msg was a resize.
func (s *Shell[S]) Resize(msg tea.Msg) bool {
	if m, ok := msg.(tea.WindowSizeMsg); ok {
		s.W, s.H = m.Width, m.Height
		return true
	}
	return false
}

func (s *Shell[S]) SetStatus(text string, isErr bool) { s.Status, s.StatusErr = text, isErr }

// Fail records an error: status line, error log and log file, with attrs
// added to the log record. Cancellations are not errors; Fail reports
// whether err was recorded.
func (s *Shell[S]) Fail(what string, err error, attrs ...any) bool {
	if err == nil || errors.Is(err, context.Canceled) {
		return false
	}
	slog.Error(what, append([]any{"err", err}, attrs...)...)
	s.Errors.Add(what, err)
	brief := ShortErr
	if s.Brief != nil {
		brief = s.Brief
	}
	s.SetStatus(what+": "+brief(err)+"  (E: error log)", true)
	return true
}

// Quit is q: at the top level, or where the app says it is at home, it
// quits; deeper it must be pressed twice so a stray q does not drop what
// was opened.
func (s *Shell[S]) Quit(home bool) tea.Cmd {
	if home || s.Stack.Len() == 1 || s.quitArmed {
		return tea.Quit
	}
	s.quitArmed = true
	s.SetStatus("press q again to quit (h goes back)", false)
	return nil
}

// Disarm forgets a first q; call it for every other key.
func (s *Shell[S]) Disarm() { s.quitArmed = false }

// Stack is the screens opened, the current one last. It is never empty.
type Stack[S comparable] struct{ items []S }

func (st *Stack[S]) Top() S   { return st.items[len(st.items)-1] }
func (st *Stack[S]) Len() int { return len(st.items) }
func (st *Stack[S]) Push(s S) { st.items = append(st.items, s) }
func (st *Stack[S]) Reset()   { st.items = st.items[:1] }

// Keep leaves at most the n bottom screens (at least the root).
func (st *Stack[S]) Keep(n int) { st.items = st.items[:max(1, min(n, len(st.items)))] }
func (st *Stack[S]) Items() []S { return st.items }

// Pop leaves the current screen, never the root.
func (st *Stack[S]) Pop() {
	if len(st.items) > 1 {
		st.items = st.items[:len(st.items)-1]
	}
}

// Has reports whether s is anywhere on the stack.
func (st *Stack[S]) Has(s S) bool {
	for _, x := range st.items {
		if x == s {
			return true
		}
	}
	return false
}

// PopTo unwinds to the topmost s and reports whether it was on the stack.
func (st *Stack[S]) PopTo(s S) bool {
	for i := len(st.items) - 1; i >= 0; i-- {
		if st.items[i] == s {
			st.items = st.items[:i+1]
			return true
		}
	}
	return false
}

// PopWhile pops while pred holds for the current screen.
func (st *Stack[S]) PopWhile(pred func(S) bool) {
	for len(st.items) > 1 && pred(st.Top()) {
		st.Pop()
	}
}

// ErrorLog keeps the errors of the session, newest last.
type ErrorLog struct {
	Entries []ErrorEntry
	Max     int // 0: unlimited
}

type ErrorEntry struct {
	At   time.Time
	What string
	Err  error
}

func (l *ErrorLog) Add(what string, err error) {
	l.Entries = append(l.Entries, ErrorEntry{At: time.Now(), What: what, Err: err})
	if l.Max > 0 && len(l.Entries) > l.Max {
		l.Entries = l.Entries[len(l.Entries)-l.Max:]
	}
}

func (l *ErrorLog) Len() int { return len(l.Entries) }

// Render is the error log page.
func (l *ErrorLog) Render(th Theme) string {
	if len(l.Entries) == 0 {
		return "\n" + th.Muted.Render("  no errors this session") + "\n"
	}
	var b strings.Builder
	b.WriteString("\n")
	for _, e := range l.Entries {
		fmt.Fprintf(&b, "  %s  %s\n    %s\n\n", th.Muted.Render(e.At.Format("15:04:05")), th.Warn.Render(e.What), th.ErrText.Render(Clean(e.Err.Error())))
	}
	return b.String()
}
