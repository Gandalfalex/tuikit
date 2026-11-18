// Package tuikit is the shell of a vim-style terminal app on bubbletea, so
// the app itself holds only its logic: what to load, what a row shows,
// what a command does.
//
//   - Shell: screen stack, status line, error log, window size, the count
//     and g/z prefix of key sequences, and q-twice-to-quit.
//   - Ops: one async operation at a time, with cancel, a spinner and stale
//     results dropped by generation.
//   - List: a filtered, scrolling list with smart-case matching and hidden
//     entries, and its standard rendering.
//   - CommandLine: the ":" line on vim-kit's cmdline, with history,
//     completion and the popup menu.
//   - Pager, Frame, StatusBar, Overlay, CommandHelp: the screen layout.
//   - Theme, Pad, Clean and friends: drawing text safely.
//   - Yank, Editor: the clipboard and $EDITOR.
//
// Key bindings, motions and selections come from vim-kit.
package tuikit
