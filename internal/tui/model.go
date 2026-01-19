package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tamjidrahman/dottodo/internal/model"
	"github.com/tamjidrahman/dottodo/internal/storage"
)

// Mode represents the current editing mode (vim-style)
type Mode int

const (
	ModeNormal Mode = iota
	ModeInsert
	ModeCommand
	ModeSearch
	ModeVisual
	ModeDetail // Interactive detail pane editing
)

// DetailField represents editable fields in the detail pane
type DetailField int

const (
	FieldText DetailField = iota
	FieldDescription
	FieldProject
	FieldUrgency
	FieldPriority
	FieldDue
	FieldTags
	FieldLinks
	FieldNotes
	FieldCount // Used to know total number of fields
)

func (m Mode) String() string {
	switch m {
	case ModeNormal:
		return "NORMAL"
	case ModeInsert:
		return "INSERT"
	case ModeCommand:
		return "COMMAND"
	case ModeSearch:
		return "SEARCH"
	case ModeVisual:
		return "VISUAL"
	case ModeDetail:
		return "DETAIL"
	default:
		return "UNKNOWN"
	}
}

func (f DetailField) String() string {
	switch f {
	case FieldText:
		return "Title"
	case FieldDescription:
		return "Description"
	case FieldProject:
		return "Project"
	case FieldUrgency:
		return "Urgency"
	case FieldPriority:
		return "Priority"
	case FieldDue:
		return "Due Date"
	case FieldTags:
		return "Tags"
	case FieldLinks:
		return "Links"
	case FieldNotes:
		return "Notes"
	default:
		return "Unknown"
	}
}

// Model is the main TUI model
type Model struct {
	storage    *storage.Storage
	todos      []*model.Todo
	cursor     int
	mode       Mode
	width      int
	height     int
	input      textinput.Model
	message    string
	yanked     *model.Todo
	filter     storage.Filter
	filterName string
	sortBy     storage.SortBy

	// Visual mode selection
	visualStart int
	selected    map[int]bool

	// Search
	searchResults []int
	searchIndex   int

	// Pending keys for multi-key commands (like dd, gg)
	pendingKey string

	// Leader key state
	leaderActive bool

	// Detail pane (split view like Gmail)
	showDetail    bool
	detailWidth   int // width of detail pane (0 = auto)

	// Detail pane field editing
	detailField   DetailField // currently selected field in detail pane
	editingField  bool        // true when editing a field value
}

func New(s *storage.Storage) Model {
	ti := textinput.New()
	ti.Prompt = ""
	ti.CharLimit = 256

	m := Model{
		storage:  s,
		input:    ti,
		selected: make(map[int]bool),
	}
	m.refreshTodos()
	return m
}

func (m *Model) refreshTodos() {
	m.todos = m.storage.Filter(m.filter)
	m.todos = m.storage.Sort(m.todos, m.sortBy)

	// Ensure cursor is in bounds
	if m.cursor >= len(m.todos) {
		m.cursor = len(m.todos) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) currentTodo() *model.Todo {
	if m.cursor >= 0 && m.cursor < len(m.todos) {
		return m.todos[m.cursor]
	}
	return nil
}

// Messages

type clearMessageMsg struct{}

func clearMessageAfter() tea.Cmd {
	return tea.Tick(time.Second*2, func(_ time.Time) tea.Msg {
		return clearMessageMsg{}
	})
}
