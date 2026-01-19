package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tamjidrahman/dottodo/internal/model"
	"github.com/tamjidrahman/dottodo/internal/storage"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case clearMessageMsg:
		m.message = ""
		return m, nil
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	// Global keys that work in any mode
	switch key {
	case "ctrl+c":
		return m, tea.Quit
	}

	// Mode-specific handling
	switch m.mode {
	case ModeNormal:
		return m.handleNormalMode(key)
	case ModeInsert:
		return m.handleInsertMode(msg)
	case ModeCommand:
		return m.handleCommandMode(msg)
	case ModeSearch:
		return m.handleSearchMode(msg)
	case ModeVisual:
		return m.handleVisualMode(key)
	}

	return m, nil
}

func (m Model) handleNormalMode(key string) (tea.Model, tea.Cmd) {
	// Handle leader key sequences
	if m.leaderActive {
		m.leaderActive = false
		return m.handleLeaderKey(key)
	}

	// Handle pending multi-key commands
	if m.pendingKey != "" {
		return m.handlePendingKey(key)
	}

	switch key {
	// Quit
	case "q":
		return m, tea.Quit

	// Navigation
	case "j", "down":
		if m.cursor < len(m.todos)-1 {
			m.cursor++
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
		}
	case "g":
		m.pendingKey = "g"
	case "G":
		m.cursor = len(m.todos) - 1
	case "ctrl+d":
		m.cursor += m.height / 2
		if m.cursor >= len(m.todos) {
			m.cursor = len(m.todos) - 1
		}
	case "ctrl+u":
		m.cursor -= m.height / 2
		if m.cursor < 0 {
			m.cursor = 0
		}
	case "H":
		m.cursor = 0
	case "M":
		m.cursor = len(m.todos) / 2
	case "L":
		m.cursor = len(m.todos) - 1

	// Actions
	case "x":
		if todo := m.currentTodo(); todo != nil {
			todo.Toggle()
			m.storage.Update(todo)
			m.refreshTodos()
			if todo.Done {
				m.message = "Marked done"
			} else {
				m.message = "Marked pending"
			}
			return m, clearMessageAfter()
		}
	case "d":
		m.pendingKey = "d"
	case "y":
		m.pendingKey = "y"
	case "p":
		if m.yanked != nil {
			newTodo := model.NewTodo(m.yanked.Text)
			newTodo.Priority = m.yanked.Priority
			newTodo.Project = m.yanked.Project
			newTodo.Tags = m.yanked.Tags
			newTodo.Due = m.yanked.Due
			m.storage.AddAt(newTodo, m.cursor+1)
			m.refreshTodos()
			m.cursor++
			m.message = "Pasted todo"
			return m, clearMessageAfter()
		}
	case "P":
		if m.yanked != nil {
			newTodo := model.NewTodo(m.yanked.Text)
			newTodo.Priority = m.yanked.Priority
			newTodo.Project = m.yanked.Project
			newTodo.Tags = m.yanked.Tags
			newTodo.Due = m.yanked.Due
			m.storage.AddAt(newTodo, m.cursor)
			m.refreshTodos()
			m.message = "Pasted todo above"
			return m, clearMessageAfter()
		}
	case ">":
		m.pendingKey = ">"
	case "<":
		m.pendingKey = "<"
	case "J":
		if todo := m.currentTodo(); todo != nil && m.cursor < len(m.todos)-1 {
			m.storage.Move(todo.ID, 1)
			m.refreshTodos()
			m.cursor++
		}
	case "K":
		if todo := m.currentTodo(); todo != nil && m.cursor > 0 {
			m.storage.Move(todo.ID, -1)
			m.refreshTodos()
			m.cursor--
		}

	// Undo/Redo
	case "u":
		if m.storage.Undo() {
			m.refreshTodos()
			m.message = "Undone"
			return m, clearMessageAfter()
		}
	case "ctrl+r":
		if m.storage.Redo() {
			m.refreshTodos()
			m.message = "Redone"
			return m, clearMessageAfter()
		}

	// Enter insert mode
	case "o":
		m.mode = ModeInsert
		m.input.SetValue("")
		m.input.Focus()
		m.message = "-- INSERT -- (new todo below)"
	case "O":
		m.mode = ModeInsert
		m.input.SetValue("")
		m.input.Focus()
		m.message = "-- INSERT -- (new todo above)"
	case "a", "A":
		if todo := m.currentTodo(); todo != nil {
			m.mode = ModeInsert
			m.input.SetValue(todo.Text + " ")
			m.input.Focus()
			m.input.CursorEnd()
			m.message = "-- INSERT -- (append)"
		}
	case "i", "I":
		if todo := m.currentTodo(); todo != nil {
			m.mode = ModeInsert
			m.input.SetValue(todo.Text)
			m.input.Focus()
			m.input.CursorStart()
			m.message = "-- INSERT -- (edit)"
		}
	case "r":
		if todo := m.currentTodo(); todo != nil {
			m.mode = ModeInsert
			m.input.SetValue(todo.Text)
			m.input.Focus()
			m.message = "-- INSERT -- (rename)"
		}

	// Command mode
	case ":":
		m.mode = ModeCommand
		m.input.SetValue("")
		m.input.Focus()
		m.input.Prompt = ":"

	// Search mode
	case "/":
		m.mode = ModeSearch
		m.input.SetValue("")
		m.input.Focus()
		m.input.Prompt = "/"
	case "n":
		m.nextSearchResult()
	case "N":
		m.prevSearchResult()

	// Visual mode
	case "v", "V":
		m.mode = ModeVisual
		m.visualStart = m.cursor
		m.selected = map[int]bool{m.cursor: true}
		m.message = "-- VISUAL --"

	// Leader key
	case " ":
		m.leaderActive = true
		m.message = "LEADER-"

	// Toggle detail pane
	case "tab":
		m.showDetail = !m.showDetail
		if m.showDetail {
			m.message = "Detail pane on"
		} else {
			m.message = "Detail pane off"
		}
		return m, clearMessageAfter()

	// Enter to open detail pane
	case "enter":
		m.showDetail = true
		m.message = "Detail view (Tab to close)"
		return m, clearMessageAfter()

	// Urgency controls
	case "!":
		if todo := m.currentTodo(); todo != nil {
			todo.IncreaseUrgency()
			m.storage.Update(todo)
			m.refreshTodos()
			m.message = "Urgency: " + todo.Urgency.String()
			return m, clearMessageAfter()
		}
	case "~":
		if todo := m.currentTodo(); todo != nil {
			todo.DecreaseUrgency()
			m.storage.Update(todo)
			m.refreshTodos()
			m.message = "Urgency: " + todo.Urgency.String()
			return m, clearMessageAfter()
		}

	// Quick link (l key followed by type)
	case "l":
		m.pendingKey = "l"
	}

	return m, nil
}

func (m Model) handlePendingKey(key string) (tea.Model, tea.Cmd) {
	pending := m.pendingKey
	m.pendingKey = ""

	switch pending {
	case "d":
		if key == "d" {
			// dd - delete current todo
			if todo := m.currentTodo(); todo != nil {
				m.yanked = todo
				m.storage.Delete(todo.ID)
				m.refreshTodos()
				m.message = "Deleted (yanked)"
				return m, clearMessageAfter()
			}
		}
	case "y":
		if key == "y" {
			// yy - yank current todo
			if todo := m.currentTodo(); todo != nil {
				m.yanked = todo
				m.message = "Yanked"
				return m, clearMessageAfter()
			}
		}
	case "g":
		if key == "g" {
			// gg - go to first
			m.cursor = 0
		}
	case ">":
		if key == ">" {
			// >> - increase priority
			if todo := m.currentTodo(); todo != nil {
				todo.IncreasePriority()
				m.storage.Update(todo)
				m.refreshTodos()
				m.message = "Priority increased"
				return m, clearMessageAfter()
			}
		}
	case "<":
		if key == "<" {
			// << - decrease priority
			if todo := m.currentTodo(); todo != nil {
				todo.DecreasePriority()
				m.storage.Update(todo)
				m.refreshTodos()
				m.message = "Priority decreased"
				return m, clearMessageAfter()
			}
		}
	case "l":
		// Link commands: ll = linear, lg = github, lu = url
		switch key {
		case "l":
			// ll - add linear link
			m.mode = ModeCommand
			m.input.SetValue("link linear ")
			m.input.Focus()
			m.input.Prompt = ":"
			m.input.CursorEnd()
		case "g":
			// lg - add github link
			m.mode = ModeCommand
			m.input.SetValue("link github ")
			m.input.Focus()
			m.input.Prompt = ":"
			m.input.CursorEnd()
		case "u":
			// lu - add url link
			m.mode = ModeCommand
			m.input.SetValue("link url ")
			m.input.Focus()
			m.input.Prompt = ":"
			m.input.CursorEnd()
		}
	}

	return m, nil
}

func (m Model) handleLeaderKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "a":
		// Quick add
		m.mode = ModeInsert
		m.input.SetValue("")
		m.input.Focus()
		m.message = "-- INSERT -- (quick add)"
	case "f":
		// Fuzzy find (just search for now)
		m.mode = ModeSearch
		m.input.SetValue("")
		m.input.Focus()
		m.input.Prompt = "/"
	case "d":
		// Due today
		m.filter = storage.Filter{DueToday: true}
		m.filterName = "due today"
		m.refreshTodos()
		m.message = "Showing: due today"
		return m, clearMessageAfter()
	case "w":
		// Due this week
		m.filter = storage.Filter{DueWeek: true}
		m.filterName = "due this week"
		m.refreshTodos()
		m.message = "Showing: due this week"
		return m, clearMessageAfter()
	case "c":
		// Completed
		done := true
		m.filter = storage.Filter{ShowDone: &done}
		m.filterName = "completed"
		m.refreshTodos()
		m.message = "Showing: completed"
		return m, clearMessageAfter()
	case "p":
		// Pending
		done := false
		m.filter = storage.Filter{ShowDone: &done}
		m.filterName = "pending"
		m.refreshTodos()
		m.message = "Showing: pending"
		return m, clearMessageAfter()
	case " ":
		// Clear filter
		m.filter = storage.Filter{}
		m.filterName = ""
		m.refreshTodos()
		m.message = "Showing: all"
		return m, clearMessageAfter()
	case "q":
		return m, tea.Quit
	}
	m.message = ""
	return m, nil
}

func (m Model) handleInsertMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "esc", "ctrl+c":
		m.mode = ModeNormal
		m.input.Blur()
		m.message = ""
		return m, nil

	case "enter":
		text := strings.TrimSpace(m.input.Value())
		if text == "" {
			m.mode = ModeNormal
			m.input.Blur()
			m.message = ""
			return m, nil
		}

		// Check if we're editing an existing todo
		if todo := m.currentTodo(); todo != nil && (strings.HasPrefix(m.message, "-- INSERT -- (edit)") ||
			strings.HasPrefix(m.message, "-- INSERT -- (rename)") ||
			strings.HasPrefix(m.message, "-- INSERT -- (append)")) {
			todo.Text = text
			todo.ParseInlineMetadata()
			m.storage.Update(todo)
			m.refreshTodos()
			m.mode = ModeNormal
			m.input.Blur()
			m.message = "Updated"
			return m, clearMessageAfter()
		}

		// Adding new todo
		newTodo := model.NewTodo(text)

		if strings.Contains(m.message, "above") {
			m.storage.AddAt(newTodo, m.cursor)
		} else {
			m.storage.AddAt(newTodo, m.cursor+1)
			m.cursor++
		}

		m.refreshTodos()
		m.mode = ModeNormal
		m.input.Blur()
		m.message = "Added"
		return m, clearMessageAfter()
	}

	// Pass to text input
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) handleCommandMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "esc", "ctrl+c":
		m.mode = ModeNormal
		m.input.Blur()
		m.input.Prompt = ""
		m.message = ""
		return m, nil

	case "enter":
		cmd := strings.TrimSpace(m.input.Value())
		m.mode = ModeNormal
		m.input.Blur()
		m.input.Prompt = ""
		return m.executeCommand(cmd)
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) handleSearchMode(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	switch key {
	case "esc", "ctrl+c":
		m.mode = ModeNormal
		m.input.Blur()
		m.input.Prompt = ""
		m.message = ""
		m.searchResults = nil
		return m, nil

	case "enter":
		term := strings.TrimSpace(m.input.Value())
		m.mode = ModeNormal
		m.input.Blur()
		m.input.Prompt = ""

		if term == "" {
			m.message = ""
			return m, nil
		}

		// Search and jump to first result
		m.search(term)
		if len(m.searchResults) > 0 {
			m.cursor = m.searchResults[0]
			m.searchIndex = 0
			m.message = "Found " + string(rune('0'+len(m.searchResults))) + " matches (n/N to navigate)"
		} else {
			m.message = "No matches found"
		}
		return m, clearMessageAfter()
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *Model) search(term string) {
	m.searchResults = nil
	term = strings.ToLower(term)
	for i, todo := range m.todos {
		if strings.Contains(strings.ToLower(todo.Text), term) ||
			strings.Contains(strings.ToLower(todo.Project), term) ||
			strings.Contains(strings.ToLower(todo.Notes), term) {
			m.searchResults = append(m.searchResults, i)
		}
	}
}

func (m *Model) nextSearchResult() {
	if len(m.searchResults) == 0 {
		return
	}
	m.searchIndex = (m.searchIndex + 1) % len(m.searchResults)
	m.cursor = m.searchResults[m.searchIndex]
}

func (m *Model) prevSearchResult() {
	if len(m.searchResults) == 0 {
		return
	}
	m.searchIndex--
	if m.searchIndex < 0 {
		m.searchIndex = len(m.searchResults) - 1
	}
	m.cursor = m.searchResults[m.searchIndex]
}

func (m Model) handleVisualMode(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "esc", "ctrl+c":
		m.mode = ModeNormal
		m.selected = make(map[int]bool)
		m.message = ""
		return m, nil

	case "j", "down":
		if m.cursor < len(m.todos)-1 {
			m.cursor++
			m.updateVisualSelection()
		}
	case "k", "up":
		if m.cursor > 0 {
			m.cursor--
			m.updateVisualSelection()
		}

	case "x", " ":
		// Toggle all selected
		for i := range m.selected {
			if i < len(m.todos) {
				m.todos[i].Toggle()
				m.storage.Update(m.todos[i])
			}
		}
		m.refreshTodos()
		m.mode = ModeNormal
		m.selected = make(map[int]bool)
		m.message = "Toggled selected"
		return m, clearMessageAfter()

	case "d":
		// Delete all selected
		for i := len(m.todos) - 1; i >= 0; i-- {
			if m.selected[i] {
				m.storage.Delete(m.todos[i].ID)
			}
		}
		m.refreshTodos()
		m.mode = ModeNormal
		m.selected = make(map[int]bool)
		m.message = "Deleted selected"
		return m, clearMessageAfter()
	}

	return m, nil
}

func (m *Model) updateVisualSelection() {
	m.selected = make(map[int]bool)
	start, end := m.visualStart, m.cursor
	if start > end {
		start, end = end, start
	}
	for i := start; i <= end; i++ {
		m.selected[i] = true
	}
}

func (m Model) executeCommand(cmd string) (tea.Model, tea.Cmd) {
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return m, nil
	}

	command := parts[0]
	args := parts[1:]

	switch command {
	case "q", "quit":
		return m, tea.Quit

	case "w", "write":
		m.storage.Save()
		m.message = "Saved"
		return m, clearMessageAfter()

	case "wq":
		m.storage.Save()
		return m, tea.Quit

	case "add":
		if len(args) > 0 {
			text := strings.Join(args, " ")
			todo := model.NewTodo(text)
			m.storage.Add(todo)
			m.refreshTodos()
			m.cursor = len(m.todos) - 1
			m.message = "Added: " + text
			return m, clearMessageAfter()
		}

	case "all":
		m.filter = storage.Filter{}
		m.filterName = ""
		m.refreshTodos()
		m.message = "Showing: all"
		return m, clearMessageAfter()

	case "done":
		if len(args) == 0 {
			// Mark current as done
			if todo := m.currentTodo(); todo != nil {
				todo.Done = true
				m.storage.Update(todo)
				m.refreshTodos()
				m.message = "Marked done"
				return m, clearMessageAfter()
			}
		} else {
			// Filter to show done
			done := true
			m.filter = storage.Filter{ShowDone: &done}
			m.filterName = "completed"
			m.refreshTodos()
			m.message = "Showing: completed"
			return m, clearMessageAfter()
		}

	case "pending":
		done := false
		m.filter = storage.Filter{ShowDone: &done}
		m.filterName = "pending"
		m.refreshTodos()
		m.message = "Showing: pending"
		return m, clearMessageAfter()

	case "today":
		m.filter = storage.Filter{DueToday: true}
		m.filterName = "due today"
		m.refreshTodos()
		m.message = "Showing: due today"
		return m, clearMessageAfter()

	case "week":
		m.filter = storage.Filter{DueWeek: true}
		m.filterName = "due this week"
		m.refreshTodos()
		m.message = "Showing: due this week"
		return m, clearMessageAfter()

	case "overdue":
		m.filter = storage.Filter{Overdue: true}
		m.filterName = "overdue"
		m.refreshTodos()
		m.message = "Showing: overdue"
		return m, clearMessageAfter()

	case "project":
		if len(args) > 0 {
			project := strings.TrimPrefix(args[0], "@")
			m.filter = storage.Filter{Project: project}
			m.filterName = "@" + project
			m.refreshTodos()
			m.message = "Showing: @" + project
			return m, clearMessageAfter()
		}

	case "tag":
		if len(args) > 0 {
			tag := strings.TrimPrefix(args[0], "#")
			m.filter = storage.Filter{Tag: tag}
			m.filterName = "#" + tag
			m.refreshTodos()
			m.message = "Showing: #" + tag
			return m, clearMessageAfter()
		}

	case "sort":
		if len(args) > 0 {
			switch args[0] {
			case "priority", "pri":
				m.sortBy = storage.SortByPriority
				m.message = "Sorted by priority"
			case "due", "date":
				m.sortBy = storage.SortByDue
				m.message = "Sorted by due date"
			case "alpha", "name":
				m.sortBy = storage.SortByAlpha
				m.message = "Sorted alphabetically"
			case "created":
				m.sortBy = storage.SortByCreated
				m.message = "Sorted by creation date"
			case "order":
				m.sortBy = storage.SortByOrder
				m.message = "Sorted by order"
			}
			m.refreshTodos()
			return m, clearMessageAfter()
		}

	case "due":
		if len(args) > 0 {
			if todo := m.currentTodo(); todo != nil {
				date := model.ParseDueDate(args[0])
				if date != nil {
					todo.Due = date
					m.storage.Update(todo)
					m.refreshTodos()
					m.message = "Due date set: " + date.RelativeString()
					return m, clearMessageAfter()
				}
			}
		}

	case "pri", "priority":
		if len(args) > 0 {
			if todo := m.currentTodo(); todo != nil {
				var p int
				if _, err := parseIntFromString(args[0], &p); err == nil {
					if p >= -2 && p <= 3 {
						todo.Priority = p
						m.storage.Update(todo)
						m.refreshTodos()
						m.message = "Priority set"
						return m, clearMessageAfter()
					}
				}
			}
		}

	case "urgency", "urg":
		if len(args) > 0 {
			if todo := m.currentTodo(); todo != nil {
				todo.Urgency = model.ParseUrgency(args[0])
				m.storage.Update(todo)
				m.refreshTodos()
				m.message = "Urgency set: " + todo.Urgency.String()
				return m, clearMessageAfter()
			}
		}

	case "link":
		if len(args) >= 2 {
			if todo := m.currentTodo(); todo != nil {
				linkType := args[0]
				linkID := strings.Join(args[1:], " ")
				todo.AddLink(linkType, linkID)
				m.storage.Update(todo)
				m.refreshTodos()
				m.message = "Link added: " + linkType
				return m, clearMessageAfter()
			}
		} else {
			m.message = "Usage: link <type> <id> (e.g., link linear ABC-123)"
			return m, clearMessageAfter()
		}

	case "note", "notes":
		if len(args) > 0 {
			if todo := m.currentTodo(); todo != nil {
				todo.Notes = strings.Join(args, " ")
				m.storage.Update(todo)
				m.refreshTodos()
				m.message = "Notes updated"
				return m, clearMessageAfter()
			}
		}

	case "help":
		m.message = "j/k:nav x:toggle dd:del o:add Tab:detail !:urgency :q:quit"
		return m, clearMessageAfter()

	case "clear":
		if len(args) > 0 && args[0] == "done" {
			// Delete all completed todos
			for i := len(m.todos) - 1; i >= 0; i-- {
				if m.todos[i].Done {
					m.storage.Delete(m.todos[i].ID)
				}
			}
			m.refreshTodos()
			m.message = "Cleared completed todos"
			return m, clearMessageAfter()
		}

	default:
		m.message = "Unknown command: " + command
		return m, clearMessageAfter()
	}

	return m, nil
}

func parseIntFromString(s string, result *int) (bool, error) {
	n := 0
	negative := false
	for i, c := range s {
		if i == 0 && c == '-' {
			negative = true
			continue
		}
		if i == 0 && c == '+' {
			continue
		}
		if c < '0' || c > '9' {
			return false, nil
		}
		n = n*10 + int(c-'0')
	}
	if negative {
		n = -n
	}
	*result = n
	return true, nil
}
