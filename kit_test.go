package tuikit

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/Gandalfalex/vim-kit/cmdline"
)

func TestText(t *testing.T) {
	for n, want := range map[int64]string{0: "0", 999: "999", 1000: "1,000", 1234567: "1,234,567", -1234: "-1,234"} {
		if got := FmtInt(n); got != want {
			t.Errorf("FmtInt(%d) = %q, want %q", n, got, want)
		}
	}
	if got := Pad("abcdef", 4); got != "abc…" {
		t.Errorf("Pad = %q", got)
	}
	if got := PadLeft("ab", 4); got != "  ab" {
		t.Errorf("PadLeft = %q", got)
	}
	if got := TruncLeft("/a/b/c/file", 6); got != "…/file" {
		t.Errorf("TruncLeft = %q", got)
	}
	if got := CleanLine("a\nb\x1b[31m\tc", 0); got != "a⏎b·[31m c" {
		t.Errorf("CleanLine = %q", got)
	}
	if got := CleanLine("héllo", 2); got != "h" {
		t.Errorf("CleanLine cut = %q", got)
	}
	if !SmartMatch("Users", "user") || SmartMatch("users", "User") || !SmartMatch("Users", "User") {
		t.Error("SmartMatch")
	}
}

type item struct{ name, text string }

func TestList(t *testing.T) {
	l := NewList(func(i item) string { return i.name })
	l.Hidden = func(i item) bool { return strings.HasPrefix(i.name, "sys") }
	l.Match = func(f string) func(item) bool {
		return func(i item) bool { return SmartMatch(i.name, f) || SmartMatch(i.text, f) }
	}
	l.Set([]item{{"app", "orders"}, {"sys", ""}, {"hr", "people"}})
	if len(l.Vis) != 2 || l.HiddenCount() != 1 {
		t.Fatalf("vis = %v, hidden = %d", l.Vis, l.HiddenCount())
	}
	l.SetFilter("peop")
	if e, ok := l.Current(); !ok || e.name != "hr" {
		t.Errorf("filter by text: %v", e)
	}
	l.SetFilter("sy") // a filter finds hidden entries
	if e, _ := l.Current(); e.name != "sys" {
		t.Errorf("filter finds hidden: %v", e)
	}
	l.SetFilter("")
	l.Select("hr")
	l.ToggleHidden()
	if len(l.Vis) != 3 || l.Cursor != 2 {
		t.Errorf("show hidden: vis = %v, cursor = %d", l.Vis, l.Cursor)
	}
	if !l.Has("sys") || l.Has("nope") {
		t.Error("Has")
	}
	out := ansi.Strip(l.View(TokyoNight, 20, 4, ListView[item]{Header: "NAME", Row: func(i item, _ int) string { return i.name }}))
	if want := "  NAME\n  app\n  sys\n▌ hr                \n~\n"; out != want {
		t.Errorf("view = %q, want %q", out, want)
	}
}

func TestStack(t *testing.T) {
	s := NewShell(TokyoNight, 0)
	s.Stack.Push(1)
	s.Stack.Push(2)
	s.Stack.Push(1)
	if !s.Stack.PopTo(2) || s.Stack.Top() != 2 || s.Stack.Len() != 3 {
		t.Fatalf("PopTo: %v", s.Stack.Items())
	}
	s.Stack.PopWhile(func(x int) bool { return x > 0 })
	s.Stack.Pop() // never the root
	if s.Stack.Len() != 1 || s.Stack.Top() != 0 {
		t.Errorf("after pops: %v", s.Stack.Items())
	}
	if s.Quit(false) == nil {
		t.Error("q at the top level does not quit")
	}
	s.Stack.Push(1)
	if s.Quit(false) != nil || s.Quit(false) == nil {
		t.Error("q below the top level must be pressed twice")
	}
	s.Disarm()
	if s.Quit(true) == nil {
		t.Error("q at home does not quit")
	}
}

type result struct{ Op }

func TestOps(t *testing.T) {
	o := NewOps(TokyoNight)
	_, first := o.Start("one")
	_, second := o.Start("two") // cancels the first
	if !o.Stale(result{first}) || o.Stale(result{second}) || o.Stale("not a result") {
		t.Error("staleness")
	}
	if o.Tick() == nil || o.Tick() != nil {
		t.Error("one spinner tick in flight")
	}
	if !o.Cancel() || o.Loading() || !o.Stale(result{second}) || o.Cancel() {
		t.Error("cancel")
	}
}

func TestFail(t *testing.T) {
	s := NewShell(TokyoNight, 0)
	if s.Fail("load", fmt.Errorf("load: %w", context.Canceled)) {
		t.Error("cancellation recorded")
	}
	s.Fail("load", errors.New("boom\nmore"))
	if s.Errors.Len() != 1 || !s.StatusErr || s.Status != "load: boom  (E: error log)" {
		t.Errorf("status = %q", s.Status)
	}
	s.Errors.Max = 2
	s.Errors.Add("a", errors.New("a"))
	s.Errors.Add("b", errors.New("b"))
	if s.Errors.Len() != 2 || s.Errors.Entries[0].What != "a" {
		t.Errorf("cap: %+v", s.Errors.Entries)
	}
}

type ctx struct {
	ran  []string
	line int
}

func TestCommandLine(t *testing.T) {
	c := CommandLine[*ctx]{
		Commands: cmdline.Set[*ctx, tea.Cmd]{
			{Name: "connect", Run: func(x *ctx, arg string) tea.Cmd { x.ran = append(x.ran, "connect "+arg); return nil }},
			{Name: "copy", Run: func(x *ctx, arg string) tea.Cmd { x.ran = append(x.ran, "copy"); return nil }},
		},
		Goto: func(x *ctx, n int) { x.line = n },
	}
	x := &ctx{}
	p := NewPrompt[int]()
	key := func(k string) (bool, error) {
		msg := tea.KeyPressMsg{Code: []rune(k)[0], Text: k}
		switch k {
		case "enter":
			msg = tea.KeyPressMsg{Code: tea.KeyEnter}
		case "tab":
			msg = tea.KeyPressMsg{Code: tea.KeyTab}
		case "backspace":
			msg = tea.KeyPressMsg{Code: tea.KeyBackspace}
		}
		_, err := HandleCommandKey(&c, x, &p, msg)
		return !p.Active(), err
	}
	p.Open(1, ModeCommand, ":", Placeholder, "")
	c.Open(x)
	key("c")
	key("o")
	key("n")
	if len(c.Line.Items) != 1 {
		t.Fatalf("completions for con: %+v", c.Line.Items)
	}
	key("tab")
	if done, err := key("enter"); !done || err != nil || len(x.ran) != 1 || x.ran[0] != "connect " {
		t.Fatalf("enter: %v %v %v", done, err, x.ran)
	}
	p.Input.SetValue("co")
	if _, err := key("enter"); err == nil {
		t.Error("ambiguous prefix accepted")
	}
	p.Input.SetValue("42")
	key("enter")
	if x.line != 42 {
		t.Errorf("goto = %d", x.line)
	}
	p.Input.SetValue("")
	if done, _ := key("backspace"); !done {
		t.Error("backspace on an empty line does not close")
	}
}

func TestPopupOverlay(t *testing.T) {
	var line cmdline.Line
	var items []cmdline.Item
	for _, s := range []string{"a", "b", "c", "d", "e"} {
		items = append(items, cmdline.Item{Value: s, Label: s})
	}
	line.Open(items)
	rows := Popup(TokyoNight, &line, 20, 3)
	if len(rows) != 3 || !strings.Contains(ansi.Strip(rows[2]), "3 more") {
		t.Fatalf("popup = %q", rows)
	}
	body := Overlay("1\n2\n3\n4\n", []string{"x", "y"}, false)
	if body != "1\n2\nx\ny\n" {
		t.Errorf("overlay = %q", body)
	}
	if body := Overlay("1\n2\n3\n", []string{"x"}, true); body != "x\n2\n3\n" {
		t.Errorf("overlay at top = %q", body)
	}
}

func TestEditorHelpers(t *testing.T) {
	if KindOf("/usr/bin/nvim") != Neovim || KindOf("vi") != Vim || KindOf("nano") != PlusLine || KindOf("code") != PlainEditor {
		t.Error("KindOf")
	}
	if RuneIndex("héllo", 4) != 2 || RuneIndex("ab", 99) != 2 || RuneIndex("ab", 0) != 0 {
		t.Error("RuneIndex")
	}
	if VimString("it's") != "'it''s'" || LuaString("a]]b") != "[=[a]]b]=]" || VimEscape("a b%") != `a\ b\%` {
		t.Error("quoting")
	}
}

func TestSortedList(t *testing.T) {
	l := NewList(func(s string) string { return s })
	l.Compare = strings.Compare
	l.Set([]string{"d", "b"})
	l.Select("d")
	l.Append([]string{"c", "a", "e"})
	if strings.Join(l.All, "") != "abcde" {
		t.Errorf("merged = %v", l.All)
	}
	if e, _ := l.Current(); e != "d" {
		t.Errorf("cursor moved to %q", e)
	}
	l.Compare = func(a, b string) int { return strings.Compare(b, a) }
	l.Resort()
	if strings.Join(l.All, "") != "edcba" || l.Cursor != 1 {
		t.Errorf("resort = %v, cursor %d", l.All, l.Cursor)
	}
	calls := 0
	l.Match = func(f string) func(string) bool {
		calls++
		return func(s string) bool { return s >= f }
	}
	l.SetFilter("c")
	if calls != 1 || len(l.Vis) != 3 {
		t.Errorf("filter compiled %d times, vis = %v", calls, l.Vis)
	}
}

func TestTitleAndText(t *testing.T) {
	bar := ansi.Strip(TitleBarPath(TokyoNight, 30, "app", []string{"prof"}, "s3://bucket/very/deep/path/file.csv", nil))
	if !strings.Contains(bar, "…") || !strings.HasSuffix(strings.TrimSpace(bar), "file.csv") {
		t.Errorf("title = %q", bar)
	}
	for _, c := range []struct {
		fn   func(string, int) string
		in   string
		w    int
		want string
	}{
		{TruncMid, "short", 10, "short"},
		{TruncMid, "export-2024-05-17.parquet", 13, "export…arquet"},
		{TruncMid, "abcdef", 1, "…"},
		{TruncLeft, "s3://bucket/a/b/c/", 10, "…et/a/b/c/"},
		{TruncLeft, "short", 10, "short"},
	} {
		if got := c.fn(c.in, c.w); got != c.want {
			t.Errorf("trunc(%q, %d) = %q, want %q", c.in, c.w, got, c.want)
		}
	}
	for n, want := range map[int64]string{512: "512 B", 1536: "1.5 KiB", 3 << 30: "3.0 GiB"} {
		if got := FmtSize(n); got != want {
			t.Errorf("FmtSize(%d) = %q", n, got)
		}
	}
}

func TestCommandLineTabOnlyAndArgument(t *testing.T) {
	lookups := 0
	c := CommandLine[*ctx]{CompleteOnTab: true, Commands: cmdline.Set[*ctx, tea.Cmd]{
		{Name: "ssh", Args: "<host>", Run: func(*ctx, string) tea.Cmd { return nil },
			Complete: func(*ctx, string) []cmdline.Item { lookups++; return nil }},
		{Name: "start", Run: func(*ctx, string) tea.Cmd { return nil }},
	}}
	x, p := &ctx{}, NewPrompt[int]()
	press := func(k string) {
		msg := tea.KeyPressMsg{Code: []rune(k)[0], Text: k}
		if k == "tab" {
			msg = tea.KeyPressMsg{Code: tea.KeyTab}
		}
		HandleCommandKey(&c, x, &p, msg)
	}
	p.Open(1, ModeCommand, ":", "", "")
	c.Open(x)
	press("s")
	press("s")
	if len(c.Line.Items) != 0 {
		t.Fatalf("looked up while typing: %+v", c.Line.Items)
	}
	press("tab")
	if p.Value() != "ssh" {
		t.Fatalf("tab: %q", p.Value())
	}
	press("tab") // the name is complete: on to the argument
	if p.Value() != "ssh " {
		t.Fatalf("second tab: %q", p.Value())
	}
	press("tab")
	if lookups != 1 {
		t.Errorf("argument completer ran %d times", lookups)
	}
}

func TestStatusExpires(t *testing.T) {
	s := NewShell(TokyoNight, 0)
	if s.Expire("") != nil {
		t.Error("timer without a TTL")
	}
	s.StatusTTL = time.Second
	s.SetStatus("saved", false)
	if s.Expire("") == nil || s.Expire("saved") != nil {
		t.Fatal("timer only for a new message")
	}
	s.Update(statusExpiredMsg{s.statusSeq - 1}) // an older message's timer
	if s.Status != "saved" {
		t.Fatal("stale timer cleared the status")
	}
	s.Update(statusExpiredMsg{s.statusSeq})
	if s.Status != "" {
		t.Fatalf("status = %q", s.Status)
	}
}
