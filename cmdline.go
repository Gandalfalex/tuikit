package tuikit

import (
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"

	"github.com/Gandalfalex/vim-kit/cmdline"
)

// CommandLine is the ":" line: the app's commands, their completion and
// history.
type CommandLine[C any] struct {
	Commands cmdline.Set[C, tea.Cmd]
	Line     cmdline.Line
	// Goto handles ":42"; nil reports it as an unknown command.
	Goto func(ctx C, line int)
}

// Placeholder is the hint in an empty command line.
const Placeholder = "command  (tab: complete, ↑↓: history, :help)"

// Open starts completion; open the prompt with ":" alongside.
func (c *CommandLine[C]) Open(ctx C) { c.Line.Open(c.Commands.Complete(ctx, "")) }

// Key handles a key of the open command line. closed is set when the
// line closed; submit is then what was entered, if anything, to Run once
// the app has closed its prompt (the command may open another).
func (c *CommandLine[C]) Key(ctx C, in *textinput.Model, msg tea.KeyPressMsg) (closed bool, submit string, cmd tea.Cmd) {
	switch k := msg.String(); k {
	case "esc", "ctrl+[":
		c.Line.Close()
		return true, "", nil
	case "enter":
		line := strings.TrimSpace(in.Value())
		c.Line.Close()
		if line != "" {
			c.Line.Commit(line)
		}
		return true, line, nil
	case "tab", "ctrl+n", "shift+tab", "ctrl+p":
		dir := 1
		if k == "shift+tab" || k == "ctrl+p" {
			dir = -1
		}
		if val, ok := c.Line.Cycle(dir); ok {
			in.SetValue(val)
			in.CursorEnd()
		}
		return false, "", nil
	case "up", "down":
		dir := 1
		if k == "up" {
			dir = -1
		}
		if val, ok := c.Line.Recall(dir); ok {
			in.SetValue(val)
			in.CursorEnd()
			c.Line.SetItems(c.Commands.Complete(ctx, val))
		}
		return false, "", nil
	case "backspace":
		if in.Value() == "" { // like vim: backspace on an empty line leaves
			c.Line.Close()
			return true, "", nil
		}
	}
	*in, cmd = in.Update(msg)
	c.Line.SetItems(c.Commands.Complete(ctx, in.Value()))
	return false, "", cmd
}

// HandleCommandKey is the usual handling of a key in the command line
// shown in prompt p: close the prompt when the line closes, then run what
// was entered.
func HandleCommandKey[C any, K comparable](c *CommandLine[C], ctx C, p *Prompt[K], msg tea.KeyPressMsg) (tea.Cmd, error) {
	closed, submit, cmd := c.Key(ctx, &p.Input, msg)
	if !closed {
		return cmd, nil
	}
	p.Close()
	if submit == "" {
		return nil, nil
	}
	return c.Run(ctx, submit)
}

// Run executes a command line: ":42" goes to line 42, anything else to the
// command its name resolves to.
func (c *CommandLine[C]) Run(ctx C, line string) (tea.Cmd, error) {
	name, arg := cmdline.Split(line)
	if n, ok := cmdline.LineNumber(name); ok && c.Goto != nil {
		c.Goto(ctx, n)
		return nil, nil
	}
	cmd, err := c.Commands.Resolve(name)
	if err != nil {
		return nil, err
	}
	return cmd.Run(ctx, arg), nil
}
