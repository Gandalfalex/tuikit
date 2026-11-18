package tuikit

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

// Prompt is the one-line input at the bottom: filter, search, ":" and
// whatever else the app asks for. Kind says which; the app routes keys by
// it while the prompt is open.
type Prompt[K comparable] struct {
	Input textinput.Model
	Kind  K
	Mode  Mode // the status line badge while open
	open  bool
}

func NewPrompt[K comparable]() Prompt[K] {
	in := textinput.New()
	in.CharLimit = 8192
	return Prompt[K]{Input: in}
}

// Open shows the prompt with text before the cursor at the end.
func (p *Prompt[K]) Open(kind K, mode Mode, prompt, placeholder, text string) tea.Cmd {
	p.Kind, p.Mode, p.open = kind, mode, true
	p.Input.Prompt, p.Input.Placeholder = prompt, placeholder
	p.Input.SetValue(text)
	p.Input.CursorEnd()
	return p.Input.Focus()
}

// Close hides the prompt; Kind goes back to its zero value.
func (p *Prompt[K]) Close() {
	var zero K
	p.Kind, p.open = zero, false
	p.Input.Blur()
}

// Active reports whether the prompt is open.
func (p *Prompt[K]) Active() bool { return p.open }

// Is reports whether the prompt is open as kind.
func (p *Prompt[K]) Is(kind K) bool { return p.open && p.Kind == kind }

func (p *Prompt[K]) Value() string { return p.Input.Value() }

// Update passes a message to the input.
func (p *Prompt[K]) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	p.Input, cmd = p.Input.Update(msg)
	return cmd
}
