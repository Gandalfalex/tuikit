package tuikit

import (
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// Pad fits s into w cells: cut with "…" when longer, padded right when
// shorter. s may hold styling.
func Pad(s string, w int) string {
	if w <= 0 {
		return ""
	}
	sw := lipgloss.Width(s)
	if sw > w {
		return ansi.Truncate(s, w, "…")
	}
	return s + strings.Repeat(" ", w-sw)
}

// PadLeft is Pad aligned right.
func PadLeft(s string, w int) string {
	sw := lipgloss.Width(s)
	if sw > w {
		return ansi.Truncate(s, w, "…")
	}
	return strings.Repeat(" ", w-sw) + s
}

// TruncLeft keeps the last w cells of s, marking the cut with "…".
func TruncLeft(s string, w int) string {
	sw := ansi.StringWidth(s)
	if sw <= w {
		return s
	}
	return "…" + ansi.TruncateLeft(s, sw-(w-1), "")
}

// FmtInt writes n with thousands separators: 1,234,567.
func FmtInt(n int64) string {
	s := strconv.FormatInt(n, 10)
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")
	if len(s) <= 3 {
		if neg {
			return "-" + s
		}
		return s
	}
	var b strings.Builder
	if neg {
		b.WriteByte('-')
	}
	pre := len(s) % 3
	if pre > 0 {
		b.WriteString(s[:pre])
	}
	for i := pre; i < len(s); i += 3 {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(s[i : i+3])
	}
	return b.String()
}

// TruncMid shortens s to w cells by cutting the middle, so the start and
// the end (dates, extensions, a trailing /) stay visible.
func TruncMid(s string, w int) string {
	sw := ansi.StringWidth(s)
	if sw <= w {
		return s
	}
	head := (w - 1) / 2
	tail := w - 1 - head
	return ansi.Truncate(s, head, "") + "…" + ansi.TruncateLeft(s, sw-tail, "")
}

// FmtSize writes a byte count in binary units: 512 B, 1.5 KiB, 3.0 GiB.
func FmtSize(n int64) string {
	const unit = 1024
	if n < unit {
		return strconv.FormatInt(n, 10) + " B"
	}
	div, exp := int64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return strconv.FormatFloat(float64(n)/float64(div), 'f', 1, 64) + " " + string("KMGTPE"[exp]) + "iB"
}

// FmtDur rounds a duration to what is worth reading in a status line.
func FmtDur(d time.Duration) string {
	switch {
	case d < time.Millisecond:
		return "<1ms"
	case d < time.Second:
		return d.Round(time.Millisecond).String()
	}
	return d.Round(10 * time.Millisecond).String()
}

// Clean makes text from outside safe to draw: control characters, which
// could carry terminal escape sequences, are replaced and tabs become
// spaces. Line breaks are kept.
func Clean(s string) string { return clean(s, false) }

// CleanLine is Clean for a single line: line breaks become "⏎", and of a
// long value only the first maxBytes are looked at (0: all).
func CleanLine(s string, maxBytes int) string {
	if maxBytes > 0 && len(s) > maxBytes {
		cut := maxBytes
		for cut > 0 && !utf8.RuneStart(s[cut]) {
			cut--
		}
		s = s[:cut]
	}
	return clean(s, true)
}

func clean(s string, oneLine bool) string {
	if strings.IndexFunc(s, unicode.IsControl) < 0 {
		return s
	}
	var b strings.Builder
	for _, r := range s {
		switch {
		case r == '\n' && oneLine:
			b.WriteRune('⏎')
		case r == '\n':
			b.WriteRune(r)
		case r == '\t':
			b.WriteByte(' ')
		case r == '\r':
		case unicode.IsControl(r):
			b.WriteRune('·')
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

// SmartMatch reports whether s contains needle, ignoring case unless
// needle has an upper-case letter (vim's smartcase).
func SmartMatch(s, needle string) bool {
	if needle == "" {
		return true
	}
	if strings.ToLower(needle) != needle {
		return strings.Contains(s, needle)
	}
	return strings.Contains(strings.ToLower(s), needle)
}

// ShortErr is the first line of an error, safe for the status line.
func ShortErr(err error) string {
	s, _, _ := strings.Cut(err.Error(), "\n")
	return CleanLine(s, 0)
}
