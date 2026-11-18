package tuikit

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"unicode/utf8"

	tea "charm.land/bubbletea/v2"
)

// EditorCommand is $VISUAL, else $EDITOR, else vi, split into words.
func EditorCommand() []string {
	for _, v := range []string{"VISUAL", "EDITOR"} {
		if f := strings.Fields(os.Getenv(v)); len(f) > 0 {
			return f
		}
	}
	return []string{"vi"}
}

// HasEditor reports whether the user named an editor of their own.
func HasEditor() bool { return os.Getenv("VISUAL") != "" || os.Getenv("EDITOR") != "" }

// EditorKind is what can be done with an editor beyond opening a file.
type EditorKind int

const (
	PlainEditor EditorKind = iota
	PlusLine               // takes +N to start on line N
	Vim                    // vi or vim: also reports the cursor on leaving
	Neovim                 // like Vim
)

// KindOf tells the kind of an editor from its command.
func KindOf(cmd string) EditorKind {
	switch filepath.Base(cmd) {
	case "vi", "vim", "view":
		return Vim
	case "nvim", "nview":
		return Neovim
	case "nano", "emacs", "micro", "kak":
		return PlusLine
	}
	return PlainEditor
}

// EditedMsg arrives when the editor has exited. Line and Col (1-based, Col
// in runes) are where vim or neovim left the cursor in the file, 0 when
// unknown.
type EditedMsg struct {
	Path      string
	Line, Col int
	Err       error
}

// Edit suspends the app and opens path in the user's editor on line (1-
// based). With vim and neovim, extra may add arguments; dir is a scratch
// directory for files they need, removed after the editor exits. The
// result is an EditedMsg.
func Edit(path string, line int, extra func(kind EditorKind, dir string) []string) (tea.Cmd, error) {
	args := EditorCommand()
	var dir, spot string
	switch kind := KindOf(args[0]); kind {
	case Vim, Neovim:
		var err error
		if dir, err = os.MkdirTemp("", "tuikit-edit-*"); err != nil {
			return nil, err
		}
		spot = filepath.Join(dir, "cursor")
		args = append(args, "+"+strconv.Itoa(line), "-c",
			"autocmd VimLeavePre * call writefile([line('.'), col('.'), expand('%:p')], "+VimString(spot)+")")
		if extra != nil {
			args = append(args, extra(kind, dir)...)
		}
	case PlusLine:
		args = append(args, "+"+strconv.Itoa(line))
	}
	args = append(args, path)
	return tea.ExecProcess(exec.Command(args[0], args[1:]...), func(err error) tea.Msg {
		msg := EditedMsg{Path: path, Err: err}
		if dir == "" {
			return msg
		}
		defer os.RemoveAll(dir)
		data, _ := os.ReadFile(spot)
		f := strings.Split(strings.TrimSpace(string(data)), "\n")
		if len(f) == 3 && samePath(f[2], path) {
			msg.Line, _ = strconv.Atoi(f[0])
			byteCol, _ := strconv.Atoi(f[1])
			if text, err := os.ReadFile(path); err == nil && msg.Line > 0 {
				if lines := strings.Split(string(text), "\n"); msg.Line <= len(lines) {
					msg.Col = RuneIndex(lines[msg.Line-1], byteCol) + 1
				}
			}
		}
		return msg
	}), nil
}

func samePath(a, b string) bool {
	ra, err1 := filepath.EvalSymlinks(a)
	rb, err2 := filepath.EvalSymlinks(b)
	return err1 == nil && err2 == nil && ra == rb
}

// RuneIndex turns vim's 1-based byte column of line into a rune index.
func RuneIndex(line string, byteCol int) int {
	return utf8.RuneCountInString(line[:max(0, min(byteCol-1, len(line)))])
}

// VimString quotes s as a vim script string literal.
func VimString(s string) string { return "'" + strings.ReplaceAll(s, "'", "''") + "'" }

// LuaString quotes s as a Lua long string.
func LuaString(s string) string {
	eq := ""
	for strings.Contains(s, "]"+eq+"]") {
		eq += "="
	}
	return "[" + eq + "[" + s + "]" + eq + "]"
}

// VimEscape escapes a file name for an ex command.
func VimEscape(s string) string {
	var b strings.Builder
	for _, r := range s {
		if strings.ContainsRune(" \\\t%#|\"", r) {
			b.WriteByte('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}
