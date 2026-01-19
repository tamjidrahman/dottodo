# dottodo - Vim-Native TUI Todo Manager

## Design Philosophy

A todo manager that feels like a natural extension of your vim/tmux workflow:
- **Modal** - Different modes for different actions (like vim)
- **Keyboard-first** - Never need a mouse
- **Fast** - Instant startup, no latency
- **Minimal** - Clean interface that works in small tmux panes
- **Composable** - Integrates with your existing tools

---

## Modes

### Normal Mode (default)
Navigate and manipulate todos without entering text.

### Insert Mode
Add new todos or edit existing ones.

### Command Mode
Execute commands (filter, search, bulk operations).

### Visual Mode
Select multiple todos for bulk operations.

---

## Keybindings

### Normal Mode Navigation
| Key | Action |
|-----|--------|
| `j` / `k` | Move down / up |
| `g` / `G` | Jump to first / last todo |
| `Ctrl-d` / `Ctrl-u` | Half-page down / up |
| `{` / `}` | Jump to prev / next project |
| `H` / `M` / `L` | Jump to top / middle / bottom of visible list |

### Normal Mode Actions
| Key | Action |
|-----|--------|
| `x` | Toggle todo complete |
| `dd` | Delete todo |
| `yy` | Yank (copy) todo |
| `p` | Paste todo below |
| `P` | Paste todo above |
| `u` | Undo |
| `Ctrl-r` | Redo |
| `>>` / `<<` | Increase / decrease priority |
| `J` | Move todo down |
| `K` | Move todo up |
| `r` | Rename/edit todo inline |
| `Enter` | Open todo detail view |

### Insert Mode
| Key | Action |
|-----|--------|
| `o` | Add new todo below current |
| `O` | Add new todo above current |
| `a` | Append to current todo |
| `i` | Insert at beginning of current todo |
| `Esc` | Return to Normal mode |
| `Ctrl-c` | Cancel and return to Normal mode |

### Command Mode
| Key | Action |
|-----|--------|
| `:` | Enter command mode |
| `/` | Search todos |
| `?` | Search backwards |
| `n` / `N` | Next / previous search result |

### Visual Mode
| Key | Action |
|-----|--------|
| `v` | Start visual selection |
| `V` | Select entire todo (line) |
| `d` | Delete selected |
| `x` | Toggle selected complete |
| `>>` | Increase priority of selected |

### Quick Actions
| Key | Action |
|-----|--------|
| `Space` | Leader key for quick actions |
| `Space a` | Quick add todo |
| `Space f` | Fuzzy find todo |
| `Space p` | Filter by project |
| `Space t` | Filter by tag |
| `Space d` | Show due today |
| `Space w` | Show due this week |
| `Space c` | Show completed |
| `Space q` | Quit |

---

## Visual Layout

### Main View (List)
```
┌─ dottodo ──────────────────────────── @work ─┐
│                                              │
│  [ ] Fix login timeout bug          +2 today │
│  [x] Review PR #423                    -1 3d │
│ >[ ] Write API documentation        +1 week  │
│  [ ] Refactor auth module                    │
│                                              │
│  ──── personal ────                          │
│  [ ] Buy groceries                     today │
│  [ ] Call mom                          sun   │
│                                              │
├──────────────────────────────────────────────┤
│ NORMAL │ 3/8 │ :help for commands            │
└──────────────────────────────────────────────┘
```

### Layout Elements
- **Header**: App name + current filter/project
- **Todo List**: Main content area
  - `[ ]` / `[x]` - Checkbox state
  - `>` - Current cursor position
  - `+1/+2` - Priority indicator
  - Due date on right (relative: today, tomorrow, 3d, week, mon)
- **Project Headers**: Visual grouping with `──── name ────`
- **Status Bar**: Current mode, position, hints

### Detail View (Enter on a todo)
```
┌─ Todo Detail ────────────────────────────────┐
│                                              │
│  Fix login timeout bug                       │
│                                              │
│  Project:  @work                             │
│  Priority: +2 (high)                         │
│  Due:      2024-01-20 (today)               │
│  Tags:     #bug #auth #urgent                │
│  Created:  2024-01-15                        │
│                                              │
│  Notes:                                      │
│  ─────                                       │
│  Users report being logged out after 5min.   │
│  Check session timeout config in auth.ts     │
│                                              │
│  Subtasks:                                   │
│  ─────────                                   │
│  [x] Reproduce the issue                     │
│  [ ] Check session timeout config            │
│  [ ] Test fix in staging                     │
│                                              │
├──────────────────────────────────────────────┤
│ q: back │ e: edit │ x: complete │ dd: delete │
└──────────────────────────────────────────────┘
```

### Responsive Design
- Adapts to terminal width
- In narrow panes (< 40 cols): Hide due dates, truncate text
- In very narrow (< 30 cols): Minimal mode, just checkboxes + text

---

## Commands

### Filtering & Viewing
```
:all              - Show all todos
:done             - Show completed
:pending          - Show pending only
:today            - Due today
:week             - Due this week
:overdue          - Past due
:project @name    - Filter by project
:tag #name        - Filter by tag
:priority +1      - Filter by priority
:search <term>    - Search todos
```

### Todo Management
```
:add <text>       - Add new todo
:edit             - Edit current todo
:delete           - Delete current todo
:done             - Mark complete
:undone           - Mark incomplete
:pri +1           - Set priority (use +1, +2, +3 or -1, -2)
:due today        - Set due date
:due tomorrow     - Set due tomorrow
:due mon          - Set due next Monday
:due 2024-01-25   - Set specific date
:project @name    - Move to project
:tag #name        - Add tag
:untag #name      - Remove tag
:note             - Add/edit notes
```

### Bulk Operations
```
:all done         - Mark all visible as done
:all delete       - Delete all visible
:all tag #name    - Tag all visible
:clear done       - Delete all completed
```

### Misc
```
:sort priority    - Sort by priority
:sort due         - Sort by due date
:sort alpha       - Sort alphabetically
:sort created     - Sort by creation date
:sync             - Sync with remote (if configured)
:export           - Export todos
:help             - Show help
:q                - Quit
```

---

## Data Model

### Todo Structure
```
{
  id: string           // unique identifier
  text: string         // todo content
  done: bool           // completion state
  priority: int        // -2 to +3 (0 = normal)
  project: string      // @project name (optional)
  tags: string[]       // #tag1 #tag2
  due: date            // due date (optional)
  created: timestamp   // when created
  completed: timestamp // when completed (if done)
  notes: string        // extended notes
  subtasks: Todo[]     // nested subtasks
  order: int           // manual sort order
}
```

### Storage
- **Default**: `~/.dottodo/todos.json` (simple JSON file)
- **Per-project**: `.dottodo/todos.json` in project root
- Auto-detect: If in a git repo with `.dottodo/`, use that

### File Format (todos.json)
```json
{
  "version": 1,
  "todos": [...],
  "config": {
    "default_project": "@inbox",
    "auto_archive_after": 7
  }
}
```

---

## tmux Integration

### Popup Mode
```bash
# Add to .tmux.conf
bind-key t display-popup -E -w 80% -h 80% "dottodo"
```

### Sidebar Mode
```bash
# Quick sidebar for todos
bind-key T split-window -h -l 40 "dottodo"
```

### Status Line Integration
```bash
# Show pending count in tmux status
set -g status-right "#(dottodo --count) todos | %H:%M"
```

---

## nvim Integration

### Quick Commands
```lua
-- Add to init.lua
vim.keymap.set('n', '<leader>tt', ':!dottodo<CR>', { desc = 'Open todos' })
vim.keymap.set('n', '<leader>ta', ':!dottodo add ', { desc = 'Add todo' })
vim.keymap.set('n', '<leader>tc', ':r !dottodo --list<CR>', { desc = 'Insert todo list' })
```

### Floating Window
```lua
-- Open in floating terminal
vim.keymap.set('n', '<leader>td', function()
  vim.cmd('terminal dottodo')
end, { desc = 'Todos in terminal' })
```

---

## CLI Interface

### Quick Actions (non-TUI)
```bash
dottodo                    # Open TUI
dottodo add "Buy milk"     # Quick add
dottodo add "Fix bug @work #urgent due:today"
dottodo list               # Print todos to stdout
dottodo list --json        # JSON output
dottodo done 3             # Mark #3 as done
dottodo --count            # Print pending count
dottodo --today            # Print today's todos
```

### Piping Support
```bash
# Add from pipe
echo "New todo" | dottodo add -

# Export to grep/fzf
dottodo list | fzf | xargs dottodo done
```

---

## Color Scheme

### Default (adapts to terminal theme)
- **Normal text**: Default terminal foreground
- **Completed**: Dim/gray
- **High priority (+2, +3)**: Bold or accent color
- **Low priority (-1, -2)**: Dim
- **Overdue**: Red/warning color
- **Due today**: Yellow/attention
- **Projects (@)**: Cyan/blue
- **Tags (#)**: Magenta/purple
- **Cursor line**: Reverse/highlight

### Accessibility
- Works with 16-color terminals
- No color-only indicators (always has text alternatives)
- High contrast mode available

---

## Startup Behavior

1. **Fast**: Target < 50ms startup time
2. **Context-aware**:
   - If in a git repo with `.dottodo/`, show project todos
   - Otherwise, show global todos from `~/.dottodo/`
3. **Remember state**: Last filter, cursor position
4. **Auto-refresh**: Watch for file changes

---

## Future Enhancements (v2+)

- [ ] Sync with external services (GitHub Issues, Todoist, etc.)
- [ ] Recurring todos
- [ ] Time tracking
- [ ] Multiple todo files
- [ ] Collaborative todos (shared file)
- [ ] Calendar view
- [ ] Pomodoro integration

---

## Tech Stack Considerations

### Option A: Rust + ratatui
- Pros: Fast, single binary, great TUI ecosystem
- Cons: Steeper learning curve

### Option B: Go + bubbletea
- Pros: Simple, fast compilation, good TUI library
- Cons: Slightly slower than Rust

### Option C: Python + textual
- Pros: Rapid development, easy to extend
- Cons: Slower startup, requires Python

### Recommendation: **Go + bubbletea**
- Balance of development speed and runtime performance
- Elm-architecture fits well with modal interface
- Single binary distribution
- Good enough performance for a todo app

---

## Summary

dottodo is designed to feel like a natural part of your vim/tmux workflow:

1. **Vim-native keybindings** - j/k, dd, yy, etc.
2. **Modal interface** - Normal, Insert, Command, Visual modes
3. **Minimal & fast** - Works in small panes, instant startup
4. **Composable** - CLI flags, pipes, integrations
5. **Context-aware** - Per-project or global todos

The goal is a tool you reach for instinctively, without leaving your terminal flow.
