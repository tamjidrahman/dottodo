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

	// Urgency colors (more vibrant and distinct)
	urgencyLowColor    = lipgloss.Color("#5599FF") // Blue
	urgencyMediumColor = lipgloss.Color("#FFAA00") // Orange/Yellow
	urgencyHighColor   = lipgloss.Color("#FF3366") // Red/Pink - very attention-grabbing

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
			Foreground(danger).
			Bold(true)

	dueTodayStyle = lipgloss.NewStyle().
			Foreground(warning).
			Bold(true)

	projectStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00AAFF"))

	tagStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AA55FF"))

	linkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00CCAA")).
			Underline(true)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(subtle).
			Padding(0, 1)

	modeStyle = lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1)

	inputStyle = lipgloss.NewStyle().
			Padding(0, 1)

	// Detail pane styles
	detailPaneStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderLeft(true).
			BorderForeground(subtle).
			Padding(1, 2)

	detailTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(highlight).
				MarginBottom(1)

	detailLabelStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#888888")).
				Width(12)

	detailValueStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#FFFFFF"))

	detailSectionStyle = lipgloss.NewStyle().
				MarginTop(1).
				MarginBottom(1)

	detailNotesStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#AAAAAA")).
				Italic(true)
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

	// Calculate available height for content
	headerHeight := 2
	statusHeight := 1
	inputHeight := 0
	if m.mode == ModeInsert || m.mode == ModeCommand || m.mode == ModeSearch {
		inputHeight = 1
	}
	contentHeight := m.height - headerHeight - statusHeight - inputHeight - 1

	// Render content (list + optional detail pane)
	content := m.renderContent(contentHeight)
	b.WriteString(content)

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

	// Detail pane indicator
	detailIndicator := ""
	if m.showDetail {
		detailIndicator = " │ detail:on"
	}

	// Right side: storage path hint
	right := ""
	if m.storage.Path() != "" {
		path := m.storage.Path()
		if len(path) > 30 {
			path = "..." + path[len(path)-27:]
		}
		right = lipgloss.NewStyle().Faint(true).Render(path)
	}

	// Build header line
	leftWidth := lipgloss.Width(title + filterText + detailIndicator)
	rightWidth := lipgloss.Width(right)
	spacer := ""
	if m.width > leftWidth+rightWidth+2 {
		spacer = strings.Repeat(" ", m.width-leftWidth-rightWidth-2)
	}

	header := title + filterText + detailIndicator + spacer + right + "\n"
	header += strings.Repeat("─", m.width-1)

	return header
}

func (m Model) renderContent(height int) string {
	if !m.showDetail {
		// Full-width list
		return m.renderList(height, m.width-2)
	}

	// Split view: list on left, detail on right (70% for detail)
	detailWidth := m.detailWidth
	if detailWidth == 0 {
		// Auto: 70% of width for detail pane
		detailWidth = m.width * 70 / 100
		if detailWidth < 40 {
			detailWidth = 40
		}
	}

	listWidth := m.width - detailWidth - 3 // 3 for border

	// Render both panes
	list := m.renderList(height, listWidth)
	detail := m.renderDetailPane(height, detailWidth)

	// Join horizontally
	listLines := strings.Split(list, "\n")
	detailLines := strings.Split(detail, "\n")

	// Pad to same height
	for len(listLines) < height {
		listLines = append(listLines, strings.Repeat(" ", listWidth))
	}
	for len(detailLines) < height {
		detailLines = append(detailLines, strings.Repeat(" ", detailWidth))
	}

	var result strings.Builder
	for i := 0; i < height; i++ {
		listLine := listLines[i]
		detailLine := ""
		if i < len(detailLines) {
			detailLine = detailLines[i]
		}

		// Pad list line to width
		listLineWidth := lipgloss.Width(listLine)
		if listLineWidth < listWidth {
			listLine += strings.Repeat(" ", listWidth-listLineWidth)
		}

		result.WriteString(listLine)
		if i == 0 {
			result.WriteString(" │ ")
		} else {
			result.WriteString(" │ ")
		}
		result.WriteString(detailLine)
		if i < height-1 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

func (m Model) renderList(height int, width int) string {
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
		line := m.renderTodo(todo, i, width)
		lines = append(lines, line)
	}

	// Pad with empty lines if needed
	for len(lines) < height {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

func (m Model) renderTodo(todo *model.Todo, index int, maxWidth int) string {
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

	// Urgency indicator (colored !)
	urgency := "  "
	if todo.Urgency > 0 {
		urgencyText := todo.Urgency.Symbol()
		var urgencyStyle lipgloss.Style
		switch todo.Urgency {
		case model.UrgencyLow:
			urgencyStyle = lipgloss.NewStyle().Foreground(urgencyLowColor)
		case model.UrgencyMedium:
			urgencyStyle = lipgloss.NewStyle().Foreground(urgencyMediumColor).Bold(true)
		case model.UrgencyHigh:
			urgencyStyle = lipgloss.NewStyle().Foreground(urgencyHighColor).Bold(true).Blink(true)
		}
		urgency = urgencyStyle.Render(urgencyText)
		// Pad to 3 chars
		if todo.Urgency == model.UrgencyLow {
			urgency += "  "
		} else if todo.Urgency == model.UrgencyMedium {
			urgency += " "
		}
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

	// Link indicator (just show count if has links)
	linkIndicator := ""
	if len(todo.Links) > 0 {
		linkIndicator = linkStyle.Render(fmt.Sprintf("[%d]", len(todo.Links)))
	}

	// Text
	text := todo.Text
	if todo.Done {
		text = doneStyle.Render(text)
	}

	// Build line: > [x] !!! Text @project [2]  +1  due
	leftPart := cursor + checkbox + " " + urgency + " " + text
	if project != "" {
		leftPart += " " + project
	}
	if linkIndicator != "" {
		leftPart += " " + linkIndicator
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
	maxLeft := maxWidth - rightWidth - 4
	if maxLeft < 20 {
		maxLeft = 20
	}
	if leftWidth > maxLeft {
		leftPart = truncate(leftPart, maxLeft-3) + "..."
		leftWidth = maxLeft
	}

	spacer := ""
	totalWidth := leftWidth + rightWidth
	if maxWidth > totalWidth+2 {
		spacer = strings.Repeat(" ", maxWidth-totalWidth-2)
	}

	line := leftPart + spacer + rightPart

	// Apply selection styling
	if isSelected && m.mode == ModeVisual {
		line = selectedStyle.Render(line)
	}

	return line
}

func (m Model) renderDetailPane(height int, width int) string {
	todo := m.currentTodo()
	if todo == nil {
		return lipgloss.NewStyle().
			Faint(true).
			Width(width).
			Render("No todo selected")
	}

	var b strings.Builder
	inDetailMode := m.mode == ModeDetail

	// Helper to render a field row with selection highlight
	renderField := func(field DetailField, label, value string) {
		isSelected := inDetailMode && m.detailField == field

		// If editing this field, show input
		if isSelected && m.editingField {
			b.WriteString(detailLabelStyle.Render(label + ":") + " ")
			b.WriteString(m.input.View())
			b.WriteString("\n")
			return
		}

		row := detailLabelStyle.Render(label+":") + " " + detailValueStyle.Render(value)
		if isSelected {
			row = lipgloss.NewStyle().
				Background(lipgloss.Color("#333333")).
				Foreground(lipgloss.Color("#FFFFFF")).
				Bold(true).
				Render("> " + label + ": " + value)
		}
		b.WriteString(row + "\n")
	}

	// Title (FieldText)
	title := todo.Text
	if len(title) > width-4 {
		title = title[:width-7] + "..."
	}
	if todo.Done && !(inDetailMode && m.detailField == FieldText) {
		title = doneStyle.Render(title)
	}

	if inDetailMode && m.detailField == FieldText {
		if m.editingField {
			b.WriteString(lipgloss.NewStyle().Bold(true).Render("Title: "))
			b.WriteString(m.input.View())
			b.WriteString("\n")
		} else {
			b.WriteString(lipgloss.NewStyle().
				Background(lipgloss.Color("#333333")).
				Bold(true).
				Render("> Title: " + title))
			b.WriteString("\n")
		}
	} else {
		b.WriteString(detailTitleStyle.Render(title))
		b.WriteString("\n")
	}
	b.WriteString("\n")

	// Description/Notes (FieldDescription)
	notesValue := todo.Notes
	if notesValue == "" {
		notesValue = "(no description)"
	}
	renderField(FieldDescription, "Description", notesValue)

	// Status row (not editable, just display)
	status := "Pending"
	if todo.Done {
		status = "Completed"
		if todo.Completed != nil {
			status += " " + todo.Completed.Format("Jan 2")
		}
	}
	b.WriteString(detailLabelStyle.Render("Status:") + " " + detailValueStyle.Render(status) + "\n")

	// Project (FieldProject)
	projectValue := todo.Project
	if projectValue == "" {
		projectValue = "(none)"
	} else {
		projectValue = "@" + projectValue
	}
	renderField(FieldProject, "Project", projectValue)

	// Urgency (FieldUrgency)
	urgencyValue := todo.Urgency.String()
	if todo.Urgency > 0 {
		urgencyValue += " " + todo.Urgency.Symbol()
	}
	if inDetailMode && m.detailField == FieldUrgency && !m.editingField {
		// Show as selected with urgency color
		var style lipgloss.Style
		switch todo.Urgency {
		case model.UrgencyLow:
			style = lipgloss.NewStyle().Foreground(urgencyLowColor)
		case model.UrgencyMedium:
			style = lipgloss.NewStyle().Foreground(urgencyMediumColor).Bold(true)
		case model.UrgencyHigh:
			style = lipgloss.NewStyle().Foreground(urgencyHighColor).Bold(true)
		default:
			style = lipgloss.NewStyle()
		}
		b.WriteString(lipgloss.NewStyle().
			Background(lipgloss.Color("#333333")).
			Bold(true).
			Render("> Urgency: " + style.Render(urgencyValue) + " [!/~ to cycle]"))
		b.WriteString("\n")
	} else {
		renderField(FieldUrgency, "Urgency", urgencyValue)
	}

	// Priority (FieldPriority)
	priValue := fmt.Sprintf("%+d", todo.Priority)
	if inDetailMode && m.detailField == FieldPriority && !m.editingField {
		b.WriteString(lipgloss.NewStyle().
			Background(lipgloss.Color("#333333")).
			Bold(true).
			Render("> Priority: " + priValue + " [</> to change]"))
		b.WriteString("\n")
	} else {
		renderField(FieldPriority, "Priority", priValue)
	}

	// Due date (FieldDue)
	dueValue := "(not set)"
	if todo.Due != nil {
		dueValue = todo.Due.String() + " (" + todo.Due.RelativeString() + ")"
	}
	renderField(FieldDue, "Due", dueValue)

	// Tags (FieldTags)
	tagsValue := "(none)"
	if len(todo.Tags) > 0 {
		tagStrs := make([]string, len(todo.Tags))
		for i, t := range todo.Tags {
			tagStrs[i] = "#" + t
		}
		tagsValue = strings.Join(tagStrs, " ")
	}
	renderField(FieldTags, "Tags", tagsValue)

	// Links (FieldLinks)
	linksValue := "(none)"
	if len(todo.Links) > 0 {
		linkStrs := make([]string, len(todo.Links))
		for i, l := range todo.Links {
			linkStrs[i] = l.Type + ":" + l.ID
		}
		linksValue = strings.Join(linkStrs, ", ")
	}
	renderField(FieldLinks, "Links", linksValue)

	// Notes section (FieldNotes) - extended notes area
	b.WriteString("\n")
	if inDetailMode && m.detailField == FieldNotes {
		if m.editingField {
			b.WriteString(lipgloss.NewStyle().Bold(true).Render("> Notes:") + "\n")
			b.WriteString(m.input.View())
		} else {
			b.WriteString(lipgloss.NewStyle().
				Background(lipgloss.Color("#333333")).
				Bold(true).
				Render("> Notes (Enter to edit)"))
			b.WriteString("\n")
			if todo.Notes != "" {
				notes := wrapText(todo.Notes, width-4)
				b.WriteString(detailNotesStyle.Render(notes))
			}
		}
	} else {
		b.WriteString(lipgloss.NewStyle().Bold(true).Faint(true).Render("Notes"))
		b.WriteString("\n")
		if todo.Notes != "" {
			notes := wrapText(todo.Notes, width-4)
			b.WriteString(detailNotesStyle.Render(notes))
		} else {
			b.WriteString(lipgloss.NewStyle().Faint(true).Render("(no notes)"))
		}
	}
	b.WriteString("\n")

	// Created date
	b.WriteString("\n")
	b.WriteString(lipgloss.NewStyle().Faint(true).Render(
		"Created " + todo.Created.Format("Jan 2, 2006 15:04")))

	// Keybindings hint
	b.WriteString("\n\n")
	if inDetailMode {
		b.WriteString(lipgloss.NewStyle().Faint(true).Render(
			"j/k:nav Enter:edit !/~:urg </>:pri h:list"))
	} else {
		b.WriteString(lipgloss.NewStyle().Faint(true).Render(
			"l:detail │ Tab:toggle │ Enter:edit"))
	}

	return b.String()
}

func (m Model) renderDetailRow(label, value string, width int) string {
	return detailLabelStyle.Render(label+":") + " " + detailValueStyle.Render(value) + "\n"
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
	case ModeDetail:
		modeColor = lipgloss.Color("#FF8800")
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
		if m.showDetail {
			message = "l:detail │ h:list │ :help"
		} else {
			message = "Tab:detail │ :help"
		}
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

func wrapText(text string, width int) string {
	if width <= 0 {
		return text
	}

	var result strings.Builder
	words := strings.Fields(text)
	lineLen := 0

	for i, word := range words {
		wordLen := len(word)

		if lineLen+wordLen+1 > width && lineLen > 0 {
			result.WriteString("\n")
			lineLen = 0
		}

		if i > 0 && lineLen > 0 {
			result.WriteString(" ")
			lineLen++
		}

		result.WriteString(word)
		lineLen += wordLen
	}

	return result.String()
}
