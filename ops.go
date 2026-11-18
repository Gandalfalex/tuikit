package tuikit

import (
	"context"
	"log/slog"
	"time"

	"charm.land/bubbles/v2/spinner"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Op tags the result of an async operation with the generation of the
// operation that produced it. Embed it in result messages.
type Op int

// OpGen is the generation the result belongs to.
func (o Op) OpGen() int { return int(o) }

// Result is a message produced by an operation started with Ops.Start.
type Result interface{ OpGen() int }

// Ops runs one async operation at a time. Starting one cancels the one in
// flight; results of a superseded or cancelled operation are stale.
type Ops struct {
	gen      int
	cancel   context.CancelFunc
	loading  bool
	what     string
	start    time.Time
	spin     spinner.Model
	spinning bool // a spinner tick is in flight
}

func NewOps(th Theme) Ops {
	return Ops{spin: spinner.New(spinner.WithSpinner(spinner.MiniDot), spinner.WithStyle(lipgloss.NewStyle().Foreground(th.C.Accent)))}
}

// Start begins an operation described by what, cancelling the previous one.
// Tag its result message with the returned Op.
func (o *Ops) Start(what string) (context.Context, Op) {
	if o.cancel != nil {
		o.cancel()
	}
	o.gen++
	o.loading, o.what, o.start = true, what, time.Now()
	slog.Debug("op start", "op", what, "gen", o.gen)
	ctx, cancel := context.WithCancel(context.Background())
	o.cancel = cancel
	return ctx, Op(o.gen)
}

// Busy shows work that must not be cancelled or superseded, such as a
// commit, without starting an operation: its result is never stale.
func (o *Ops) Busy(what string) {
	o.loading, o.what, o.start = true, what, time.Now()
}

// Done ends the operation whose result arrived.
func (o *Ops) Done() {
	if o.loading {
		slog.Debug("op done", "op", o.what, "took", time.Since(o.start).Round(time.Millisecond))
	}
	o.loading, o.what = false, ""
}

// Cancel aborts the operation in flight, if any, and reports whether there
// was one. Its result will be stale.
func (o *Ops) Cancel() bool {
	if !o.loading {
		return false
	}
	o.cancel()
	o.Done()
	o.gen++
	return true
}

// Stale reports whether msg is the result of an operation that was
// superseded or cancelled.
func (o *Ops) Stale(msg tea.Msg) bool {
	r, ok := msg.(Result)
	return ok && r.OpGen() != o.gen
}

func (o *Ops) Loading() bool { return o.loading }

// What describes the operation in flight.
func (o *Ops) What() string { return o.what }

// Spinner is the spinner and what is being done, "" when idle.
func (o *Ops) Spinner() string {
	if !o.loading {
		return ""
	}
	return o.spin.View() + " " + o.what
}

// Frame is the spinner alone, for showing progress elsewhere.
func (o *Ops) Frame() string { return o.spin.View() }

// Tick keeps the spinner turning while an operation runs; call it after
// every Update.
func (o *Ops) Tick() tea.Cmd {
	if !o.loading || o.spinning {
		return nil
	}
	o.spinning = true
	return o.spin.Tick
}

// Update advances the spinner; handled is false for other messages.
func (o *Ops) Update(msg tea.Msg) (cmd tea.Cmd, handled bool) {
	if _, ok := msg.(spinner.TickMsg); !ok {
		return nil, false
	}
	if !o.loading {
		o.spinning = false
		return nil, true
	}
	o.spin, cmd = o.spin.Update(msg)
	return cmd, true
}
