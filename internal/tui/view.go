package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/tamjidrahman/dottodo/internal/model"
)

var (
	// Colors
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}
	warning   = lipgloss.AdaptiveColor{Light: "#FF8800", Dark: "#FFAA00"}
	danger    = lipgloss.AdaptiveColor{Light: "#FF0000", Dark: "#FF5555"}

	// Styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(highlight)

	cursorStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(special)

	selectedStyle = lipgloss.NewStyle().
			Background(subtle)

	doneStyle = lipgloss.NewStyle().
			Faint(true).
			Strikethrough(true)

	priorityHighStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(warning)

	priorityLowStyle = lipgloss.NewStyle().
				Faint(true)

	overdueStyle = lipgloss.NewStyle().
			Foreground(danger)

	dueTodayStyle = lipgloss.NewStyle().
			Foreground(warning)

	projectStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00AAFF"))

	tagStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AA55FF"))

	statusBarStyle = lipgloss.NewStyle().
			Foreground(subtle).
			Padding(0, 1)

	modeStyle = lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1)

	inputStyle = lipgloss.NewStyle().
			Padding(0, 1)
)

func (m Model) View() string {
	if m.width == 0 {
		return "Loading..."
	}

	var b strings.Builder

	// Header
	header := m.renderHeader()
	b.WriteString(header)
	b.WriteString("\n")

	// Calculate available height for list
	// Header: 2 lines, Status bar: 1 line, Input (if visible): 1 line
	headerHeight := 2
	statusHeight := 1
	inputHeight := 0
	if m.mode == ModeInsert || m.mode == ModeCommand || m.mode == ModeSearch {
		inputHeight = 1
	}
	listHeight := m.height - headerHeight - statusHeight - inputHeight - 1

	// Todo list
	list := m.renderList(listHeight)
	b.WriteString(list)

	// Input line (for insert/command/search modes)
	if inputHeight > 0 {
		b.WriteString("\n")
		b.WriteString(inputStyle.Render(m.input.View()))
	}

	// Status bar
	b.WriteString("\n")
	statusBar := m.renderStatusBar()
	b.WriteString(statusBar)

	return b.String()
}

func (m Model) renderHeader() string {
	title := titleStyle.Render("dottodo")

	// Filter indicator
	filterText := ""
	if m.filterName != "" {
		filterText = " [" + m.filterName + "]"
	}

	// Right side: storage path hint
	right := ""
	if m.storage.Path() != "" {
		// Show abbreviated path
		path := m.storage.Path()
		if len(path) > 30 {
			path = "..." + path[len(path)-27:]
		}
		right = lipgloss.NewStyle().Faint(true).Render(path)
	}

	// Build header line
	leftWidth := lipgloss.Width(title + filterText)
	rightWidth := lipgloss.Width(right)
	spacer := ""
	if m.width > leftWidth+rightWidth+2 {
		spacer = strings.Repeat(" ", m.width-leftWidth-rightWidth-2)
	}

	header := title + filterText + spacer + right + "\n"
	header += strings.Repeat("─", m.width-1)

	return header
}

func (m Model) renderList(height int) string {
	if len(m.todos) == 0 {
		empty := "\n  No todos"
		if m.filterName != "" {
			empty += " matching filter"
		}
		empty += ". Press 'o' to add one."
		return lipgloss.NewStyle().Faint(true).Render(empty)
	}

	var lines []string

	// Calculate viewport
	start := 0
	if m.cursor >= height {
		start = m.cursor - height + 1
	}
	end := start + height
	if end > len(m.todos) {
		end = len(m.todos)
	}

	for i := start; i < end; i++ {
		todo := m.todos[i]
		line := m.renderTodo(todo, i)
		lines = append(lines, line)
	}

	// Pad with empty lines if needed
	for len(lines) < height {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderTodo(todo *model.Todo, index int) string {
	isCursor := index == m.cursor
	isSelected := m.selected[index]

	// Checkbox
	checkbox := "[ ]"
	if todo.Done {
		checkbox = "[x]"
	}

	// Cursor indicator
	cursor := "  "
	if isCursor {
		cursor = cursorStyle.Render("> ")
	}

	// Priority indicator
	priority := "  "
	if todo.Priority > 0 {
		priority = priorityHighStyle.Render(fmt.Sprintf("+%d", todo.Priority))
	} else if todo.Priority < 0 {
		priority = priorityLowStyle.Render(fmt.Sprintf("%d", todo.Priority))
	}

	// Due date
	due := ""
	if todo.Due != nil {
		dueText := todo.Due.RelativeString()
		if todo.Due.IsOverdue() {
			due = overdueStyle.Render(dueText)
		} else if todo.Due.IsToday() {
			due = dueTodayStyle.Render(dueText)
		} else {
			due = lipgloss.NewStyle().Faint(true).Render(dueText)
		}
	}

	// Project
	project := ""
	if todo.Project != "" {
		project = projectStyle.Render("@" + todo.Project)
	}

	// Tags
	tags := ""
	if len(todo.Tags) > 0 {
		tagStrs := make([]string, len(todo.Tags))
		for i, t := range todo.Tags {
			tagStrs[i] = "#" + t
		}
		tags = tagStyle.Render(strings.Join(tagStrs, " "))
	}

	// Text
	text := todo.Text
	if todo.Done {
		text = doneStyle.Render(text)
	}

	// Build line
	// Format: > [ ] Text          @project #tags  +1  due
	leftPart := cursor + checkbox + " " + text
	if project != "" {
		leftPart += " " + project
	}
	if tags != "" {
		leftPart += " " + tags
	}

	rightPart := ""
	if priority != "  " {
		rightPart += priority + " "
	}
	if due != "" {
		rightPart += due
	}

	// Calculate spacing
	leftWidth := lipgloss.Width(leftPart)
	rightWidth := lipgloss.Width(rightPart)

	// Truncate text if needed
	maxLeft := m.width - rightWidth - 4
	if maxLeft < 20 {
		maxLeft = 20
	}
	if leftWidth > maxLeft {
		// Truncate
		leftPart = truncate(leftPart, maxLeft-3) + "..."
		leftWidth = maxLeft
	}

	spacer := ""
	totalWidth := leftWidth + rightWidth
	if m.width > totalWidth+2 {
		spacer = strings.Repeat(" ", m.width-totalWidth-2)
	}

	line := leftPart + spacer + rightPart

	// Apply selection styling
	if isSelected && m.mode == ModeVisual {
		line = selectedStyle.Render(line)
	}

	return line
}

func (m Model) renderStatusBar() string {
	// Mode
	modeText := m.mode.String()
	modeColor := lipgloss.Color("#888888")
	switch m.mode {
	case ModeInsert:
		modeColor = lipgloss.Color("#00AA00")
	case ModeCommand:
		modeColor = lipgloss.Color("#AAAA00")
	case ModeSearch:
		modeColor = lipgloss.Color("#00AAAA")
	case ModeVisual:
		modeColor = lipgloss.Color("#AA00AA")
	}
	mode := modeStyle.Foreground(modeColor).Render(modeText)

	// Position
	pos := ""
	if len(m.todos) > 0 {
		pos = fmt.Sprintf("%d/%d", m.cursor+1, len(m.todos))
	}

	// Pending count
	pending := fmt.Sprintf("%d pending", m.storage.PendingCount())

	// Message or help hint
	message := m.message
	if message == "" && m.mode == ModeNormal {
		message = ":help for commands"
	}

	// Build status bar
	left := mode + " │ " + pos + " │ " + pending
	right := message

	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	spacer := ""
	if m.width > leftWidth+rightWidth+2 {
		spacer = strings.Repeat(" ", m.width-leftWidth-rightWidth-2)
	}

	return statusBarStyle.Render(left + spacer + right)
}

func truncate(s string, max int) string {
	if max <= 0 {
		return ""
	}

	// Count visible characters (ignoring ANSI escape codes)
	visible := 0
	result := strings.Builder{}
	inEscape := false

	for _, r := range s {
		if r == '\x1b' {
			inEscape = true
		}

		if inEscape {
			result.WriteRune(r)
			if r == 'm' {
				inEscape = false
			}
			continue
		}

		if visible >= max {
			break
		}

		result.WriteRune(r)
		visible++
	}

	return result.String()
}
