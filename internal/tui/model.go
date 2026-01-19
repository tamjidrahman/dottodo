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
	default:
		return "UNKNOWN"
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
