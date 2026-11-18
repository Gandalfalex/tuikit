package tuikit

import (
	"sort"
	"strings"

	"github.com/Gandalfalex/vim-kit/cmdline"
)

// MotionHelp describes the motions every list and page has.
const MotionHelp = `  MOTIONS (all accept a count, e.g. 5j, 10G)
    j / k  ↓ / ↑        move down / up
    gg / G              first / last  (3gg, 42G jump to line)
    ctrl+d / ctrl+u     half page down / up
    ctrl+f / ctrl+b     full page down / up
    ctrl+e / ctrl+y     scroll view without moving
    H / M / L           top / middle / bottom of screen
    zz / zt / zb        center / top / bottom cursor line
`

// CommandHelp lists commands sorted by name, with arguments and aliases.
func CommandHelp[C, R any](set cmdline.Set[C, R]) string {
	cs := append(cmdline.Set[C, R](nil), set...)
	sort.Slice(cs, func(i, j int) bool { return cs[i].Name < cs[j].Name })
	var b strings.Builder
	for _, c := range cs {
		label := ":" + c.Name
		if c.Args != "" {
			label += " " + c.Args
		}
		alias := ""
		if len(c.Aliases) > 0 {
			alias = " (:" + strings.Join(c.Aliases, ", :") + ")"
		}
		b.WriteString("    " + Pad(label, 26) + c.Desc + alias + "\n")
	}
	return b.String()
}
