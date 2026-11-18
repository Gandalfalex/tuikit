package tuikit

import (
	"strings"

	"github.com/charmbracelet/x/ansi"

	"github.com/Gandalfalex/vim-kit/keys"
	"github.com/Gandalfalex/vim-kit/motion"
)

// List is a filtered, scrolling list. Vis indexes the entries of All that
// pass the filter; the cursor (motion.Window) moves over Vis.
type List[T any] struct {
	All []T
	Vis []int
	motion.Window
	Filter     string
	ShowHidden bool

	// Key identifies an entry, for keeping the cursor on it.
	Key func(T) string
	// Match reports whether an entry passes a non-empty filter; nil
	// matches Key smart-case.
	Match func(item T, filter string) bool
	// Hidden entries are left out unless ShowHidden is set or a filter is
	// typed, which finds them too; nil hides nothing.
	Hidden func(T) bool
}

// NewList is a list whose entries are identified by key.
func NewList[T any](key func(T) string) List[T] { return List[T]{Key: key} }

// Set replaces the entries, keeping the filter.
func (l *List[T]) Set(items []T) {
	l.All = items
	l.Refilter()
}

// SetFilter replaces the filter and moves the cursor to the first match.
func (l *List[T]) SetFilter(f string) {
	l.Filter = f
	l.Cursor, l.Top = 0, 0
	l.Refilter()
}

// Refilter recomputes Vis after entries or the filter changed.
func (l *List[T]) Refilter() {
	l.Vis = l.Vis[:0]
	for i, e := range l.All {
		if l.Hidden != nil && !l.ShowHidden && l.Filter == "" && l.Hidden(e) {
			continue
		}
		if l.Filter == "" || l.match(e) {
			l.Vis = append(l.Vis, i)
		}
	}
	l.Clamp(len(l.Vis))
}

func (l *List[T]) match(e T) bool {
	if l.Match != nil {
		return l.Match(e, l.Filter)
	}
	return SmartMatch(l.Key(e), l.Filter)
}

// Current is the entry under the cursor.
func (l *List[T]) Current() (T, bool) {
	if len(l.Vis) == 0 {
		var zero T
		return zero, false
	}
	return l.All[l.Vis[l.Cursor]], true
}

// CurrentPtr is the entry under the cursor, to change in place.
func (l *List[T]) CurrentPtr() *T {
	if len(l.Vis) == 0 {
		return nil
	}
	return &l.All[l.Vis[l.Cursor]]
}

// Select moves the cursor to the visible entry with key.
func (l *List[T]) Select(key string) bool {
	return l.Window.Select(len(l.Vis), func(i int) bool { return l.Key(l.All[l.Vis[i]]) == key })
}

// Has reports whether an entry with key exists, visible or not.
func (l *List[T]) Has(key string) bool {
	for _, e := range l.All {
		if l.Key(e) == key {
			return true
		}
	}
	return false
}

// HiddenCount counts the entries Hidden leaves out.
func (l *List[T]) HiddenCount() int {
	n := 0
	for _, e := range l.All {
		if l.Hidden != nil && l.Hidden(e) {
			n++
		}
	}
	return n
}

// ToggleHidden shows or hides the hidden entries, keeping the cursor on
// the entry under it.
func (l *List[T]) ToggleHidden() {
	cur, ok := l.Current()
	l.ShowHidden = !l.ShowHidden
	l.Refilter()
	if ok {
		l.Select(l.Key(cur))
	}
}

// Apply moves the cursor by a motion within a list of height h.
func (l *List[T]) Apply(ev keys.Event, h int) bool {
	defer l.Scroll(h)
	return l.Window.Apply(ev, len(l.Vis), h)
}

// Goto puts the cursor on line n (1-based) and centres it, for :n.
func (l *List[T]) Goto(n, h int) {
	l.Cursor = n - 1
	l.Clamp(len(l.Vis))
	l.Center(len(l.Vis), h)
}

// Scroll keeps the cursor visible in a list of height h.
func (l *List[T]) Scroll(h int) { l.Window.Scroll(len(l.Vis), h) }

// Position is "3/40", plus "(of 52)" when the filter hides some.
func (l *List[T]) Position() []string {
	pos := 0
	if len(l.Vis) > 0 {
		pos = l.Cursor + 1
	}
	out := []string{itoa(pos) + "/" + itoa(len(l.Vis))}
	if len(l.Vis) != len(l.All) {
		out = append(out, "(of "+itoa(len(l.All))+")")
	}
	return out
}

// ListView is how List.View draws.
type ListView[T any] struct {
	Header string                  // column headers, after the cursor gutter
	Row    func(e T, i int) string // one row, without the gutter; i indexes All
	Empty  string                  // shown without entries; "" for "(empty)"
}

// View draws the header, h rows with the cursor marked, and "~" below.
func (l *List[T]) View(th Theme, w, h int, v ListView[T]) string {
	l.Scroll(h)
	var b strings.Builder
	b.WriteString(th.Header.Render(ansi.Truncate("  "+v.Header, w, "")))
	b.WriteByte('\n')
	lines := 0
	if len(l.Vis) == 0 {
		msg := v.Empty
		switch {
		case l.Filter != "":
			msg = "no matches for filter: " + l.Filter
		case msg == "":
			msg = "(empty)"
		}
		b.WriteString(th.Muted.Render("  " + msg))
		b.WriteByte('\n')
		lines++
	}
	for i := l.Top; i < len(l.Vis) && i < l.Top+h; i++ {
		row := v.Row(l.All[l.Vis[i]], l.Vis[i])
		if i == l.Cursor {
			row = th.Sel.Render(Pad("▌ "+ansi.Strip(row), w))
		} else {
			row = "  " + row
		}
		b.WriteString(ansi.Truncate(row, w, ""))
		b.WriteByte('\n')
		lines++
	}
	b.WriteString(Filler(th, h-lines))
	return b.String()
}

// Filler is n "~" lines, vim's mark for lines past the end.
func Filler(th Theme, n int) string {
	if n <= 0 {
		return ""
	}
	return strings.Repeat(th.Muted.Render("~")+"\n", n)
}
