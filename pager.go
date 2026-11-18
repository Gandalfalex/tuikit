package tuikit

import (
	"fmt"

	"charm.land/bubbles/v2/viewport"

	"github.com/Gandalfalex/vim-kit/keys"
	"github.com/Gandalfalex/vim-kit/motion"
)

// Pager shows a page of text: help, the error log, a record.
type Pager struct{ VP viewport.Model }

func NewPager() Pager { return Pager{VP: viewport.New()} }

// Resize fits the pager into a frame of w×h.
func (p *Pager) Resize(w, h int) {
	p.VP.SetWidth(w)
	p.VP.SetHeight(max(1, h-Chrome))
}

// Show replaces the content, soft-wrapped or not, at the top or bottom.
func (p *Pager) Show(content string, wrap, atBottom bool) {
	p.VP.SoftWrap = wrap
	p.VP.SetContent(content)
	if atBottom {
		p.VP.GotoBottom()
	} else {
		p.VP.GotoTop()
	}
}

// Refresh replaces the content and keeps the scroll position.
func (p *Pager) Refresh(content string) { p.VP.SetContent(content) }

// Apply moves the page by a motion.
func (p *Pager) Apply(ev keys.Event) bool { return motion.ApplyPager(pagerAdapter{&p.VP}, ev) }

// Goto scrolls to line n (1-based), for :n.
func (p *Pager) Goto(n int) { p.VP.SetYOffset(n - 1) }

// View is the page; it ends in a newline, as a frame body.
func (p *Pager) View() string { return p.VP.View() + "\n" }

// Percent is how far down the page is, for the status line.
func (p *Pager) Percent() string { return fmt.Sprintf("%3.0f%%", p.VP.ScrollPercent()*100) }

// PagerMotions are the vim motions plus less-style enter and space.
func PagerMotions() keys.Bindings {
	b := motion.VerticalBindings()
	b[motion.Down] = append(b[motion.Down], "enter")
	b[motion.PageDown] = append(b[motion.PageDown], "space")
	return b
}

type pagerAdapter struct{ *viewport.Model }

func (p pagerAdapter) Top()    { p.GotoTop() }
func (p pagerAdapter) Bottom() { p.GotoBottom() }
