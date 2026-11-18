package tuikit

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Palette holds the colours of a theme.
type Palette struct {
	Accent, Muted, Dir, Err, OK, Warn, SelBg, BarBg, Dark color.Color
}

// Theme holds the colours and the styles built from them.
type Theme struct {
	C Palette

	Title   lipgloss.Style // app name in the title bar
	Crumb   lipgloss.Style // path parts in the title bar, names
	Muted   lipgloss.Style
	Dir     lipgloss.Style // containers: directories, tables
	Sel     lipgloss.Style // the line under the cursor
	Header  lipgloss.Style // column headers
	ErrText lipgloss.Style
	OKText  lipgloss.Style
	Warn    lipgloss.Style
	Bar     lipgloss.Style // popup and bar background
	Mode    lipgloss.Style // the mode badge; its background is the mode colour
	Section lipgloss.Style // section headings in pages
	Found   lipgloss.Style // search matches
	Danger  lipgloss.Style // badges for things that cannot be undone
	Caution lipgloss.Style
}

// NewTheme builds the styles from the colours.
func NewTheme(c Palette) Theme {
	s := lipgloss.NewStyle
	accent, muted, dir, errC, ok, warn, selBg, barBg, dark := c.Accent, c.Muted, c.Dir, c.Err, c.OK, c.Warn, c.SelBg, c.BarBg, c.Dark
	return Theme{
		C:       c,
		Title:   s().Bold(true).Foreground(dark).Background(accent).Padding(0, 1),
		Crumb:   s().Foreground(accent),
		Muted:   s().Foreground(muted),
		Dir:     s().Foreground(dir).Bold(true),
		Sel:     s().Background(selBg).Bold(true),
		Header:  s().Foreground(muted).Bold(true),
		ErrText: s().Foreground(errC),
		OKText:  s().Foreground(ok),
		Warn:    s().Foreground(warn),
		Bar:     s().Background(barBg),
		Mode:    s().Bold(true).Foreground(dark).Padding(0, 1),
		Section: s().Foreground(warn).Bold(true),
		Found:   s().Background(warn).Foreground(dark),
		Danger:  s().Bold(true).Foreground(dark).Background(errC).Padding(0, 1),
		Caution: s().Bold(true).Foreground(dark).Background(warn).Padding(0, 1),
	}
}

// TokyoNight is the default theme.
var TokyoNight = NewTheme(Palette{
	Accent: lipgloss.Color("#7aa2f7"), Muted: lipgloss.Color("#565f89"), Dir: lipgloss.Color("#7dcfff"),
	Err: lipgloss.Color("#f7768e"), OK: lipgloss.Color("#9ece6a"), Warn: lipgloss.Color("#e0af68"),
	SelBg: lipgloss.Color("#283457"), BarBg: lipgloss.Color("#1f2335"), Dark: lipgloss.Color("#1a1b26"),
})

// Mode is what the badge at the left of the status line says.
type Mode struct {
	Name  string
	Color color.Color
}

// The modes every app has; apps add their own.
var (
	ModeNormal  = Mode{"NORMAL", TokyoNight.C.Accent}
	ModeFilter  = Mode{"FILTER", TokyoNight.C.OK}
	ModeCommand = Mode{"COMMAND", TokyoNight.C.Dir}
	ModeSearch  = Mode{"SEARCH", TokyoNight.C.Warn}
	ModeInsert  = Mode{"INSERT", TokyoNight.C.OK}
	ModeVisual  = Mode{"VISUAL", TokyoNight.C.Warn}
	ModeVLine   = Mode{"V-LINE", TokyoNight.C.Warn}
)
