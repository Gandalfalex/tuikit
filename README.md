# tuikit

The shell of a vim-style terminal app on [bubbletea](https://github.com/charmbracelet/bubbletea) v2,
so the app holds only its own logic. Key bindings, motions and selections
come from [vim-kit](https://github.com/Gandalfalex/vim-kit), which stays
dependency-free; this kit is the bubbletea layer on top.

| Piece | What it does |
|---|---|
| `Shell[S]` | screen stack, status line, error log (`E`), window size, count / `g` / `z` key prefixes, q-twice-to-quit; embed it in the app model |
| `Ops` | one async operation at a time: cancel, spinner, stale results dropped by generation (`Op` in result messages) |
| `List[T]` | filtered, scrolling list: smart-case filter, hidden entries (`zh`), keep-cursor-on-key, `:42`, standard rendering |
| `Prompt[K]` | the one-line input for filter, search, `:` and app prompts, with the mode badge |
| `CommandLine[C]` | `:` commands on vim-kit's `cmdline`: completion, history, popup, `:42` |
| `Pager` | help, error log and record pages with vim and less keys |
| `Frame`, `TitleBar`, `StatusBar`, `PromptBar`, `HintBar`, `Popup`, `Overlay` | the screen layout |
| `CommandHelp`, `MotionHelp` | the help page from the command set |
| `Theme`, `Mode` | colours and styles (TokyoNight default), status line modes |
| `Pad`, `PadLeft`, `TruncLeft`, `Clean`, `CleanLine`, `FmtInt`, `FmtDur`, `SmartMatch`, `ShortErr` | drawing text safely |
| `Yank`, `StatusMsg` | clipboard (OSC 52, pbcopy) with a status message |
| `Edit`, `EditedMsg`, `EditorKind` | open a file in `$VISUAL`/`$EDITOR` on a line; vim and neovim report the cursor back and can be given extra setup |

The app keeps: what it loads, what a row and a page show, its commands and
key actions, and its own modes.

```go
type model struct {
	tk.Shell[screen]
	items  tk.List[item]
	prompt tk.Prompt[inputKind]
	cmd    tk.CommandLine[*model]
}
```

Used by dbreader; s3reader and gdocker are next.
