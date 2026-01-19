package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tamjidrahman/dottodo/internal/model"
)

type TodoFile struct {
	Version int            `json:"version"`
	Todos   []*model.Todo  `json:"todos"`
	Config  StorageConfig  `json:"config,omitempty"`
}

type StorageConfig struct {
	DefaultProject   string `json:"default_project,omitempty"`
	AutoArchiveAfter int    `json:"auto_archive_after,omitempty"`
}

type Storage struct {
	path     string
	file     *TodoFile
	undoStack []*TodoFile
	redoStack []*TodoFile
}

func New() (*Storage, error) {
	path := findTodoFile()
	s := &Storage{
		path: path,
		file: &TodoFile{Version: 1},
	}

	if err := s.Load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return s, nil
}

func NewWithPath(path string) (*Storage, error) {
	s := &Storage{
		path: path,
		file: &TodoFile{Version: 1},
	}

	if err := s.Load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return s, nil
}

func findTodoFile() string {
	// Check for local .dottodo directory (in a git repo)
	if _, err := os.Stat(".dottodo"); err == nil {
		return ".dottodo/todos.json"
	}

	// Walk up to find .dottodo in parent directories
	dir, _ := os.Getwd()
	for {
		dotDir := filepath.Join(dir, ".dottodo")
		if _, err := os.Stat(dotDir); err == nil {
			return filepath.Join(dotDir, "todos.json")
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// Fall back to home directory
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return filepath.Join(home, ".dottodo", "todos.json")
}

func (s *Storage) Load() error {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, &s.file)
}

func (s *Storage) Save() error {
	// Ensure directory exists
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s.file, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.path, data, 0644)
}

func (s *Storage) saveState() {
	// Deep copy current state for undo
	data, _ := json.Marshal(s.file)
	var copy TodoFile
	json.Unmarshal(data, &copy)
	s.undoStack = append(s.undoStack, &copy)
	s.redoStack = nil // Clear redo stack on new action

	// Limit undo stack size
	if len(s.undoStack) > 50 {
		s.undoStack = s.undoStack[1:]
	}
}

func (s *Storage) Undo() bool {
	if len(s.undoStack) == 0 {
		return false
	}

	// Save current state to redo
	data, _ := json.Marshal(s.file)
	var current TodoFile
	json.Unmarshal(data, &current)
	s.redoStack = append(s.redoStack, &current)

	// Restore previous state
	s.file = s.undoStack[len(s.undoStack)-1]
	s.undoStack = s.undoStack[:len(s.undoStack)-1]
	s.Save()
	return true
}

func (s *Storage) Redo() bool {
	if len(s.redoStack) == 0 {
		return false
	}

	// Save current state to undo
	data, _ := json.Marshal(s.file)
	var current TodoFile
	json.Unmarshal(data, &current)
	s.undoStack = append(s.undoStack, &current)

	// Restore redo state
	s.file = s.redoStack[len(s.redoStack)-1]
	s.redoStack = s.redoStack[:len(s.redoStack)-1]
	s.Save()
	return true
}

func (s *Storage) Todos() []*model.Todo {
	return s.file.Todos
}

func (s *Storage) Add(todo *model.Todo) {
	s.saveState()

	// Set order to end
	maxOrder := 0
	for _, t := range s.file.Todos {
		if t.Order > maxOrder {
			maxOrder = t.Order
		}
	}
	todo.Order = maxOrder + 1

	s.file.Todos = append(s.file.Todos, todo)
	s.Save()
}

func (s *Storage) AddAt(todo *model.Todo, index int) {
	s.saveState()

	if index < 0 {
		index = 0
	}
	if index > len(s.file.Todos) {
		index = len(s.file.Todos)
	}

	// Insert at position
	s.file.Todos = append(s.file.Todos[:index], append([]*model.Todo{todo}, s.file.Todos[index:]...)...)

	// Reorder
	for i, t := range s.file.Todos {
		t.Order = i
	}

	s.Save()
}

func (s *Storage) Update(todo *model.Todo) {
	s.saveState()

	for i, t := range s.file.Todos {
		if t.ID == todo.ID {
			s.file.Todos[i] = todo
			break
		}
	}
	s.Save()
}

func (s *Storage) Delete(id string) *model.Todo {
	s.saveState()

	for i, t := range s.file.Todos {
		if t.ID == id {
			deleted := t
			s.file.Todos = append(s.file.Todos[:i], s.file.Todos[i+1:]...)
			s.Save()
			return deleted
		}
	}
	return nil
}

func (s *Storage) Move(id string, delta int) {
	s.saveState()

	for i, t := range s.file.Todos {
		if t.ID == id {
			newIndex := i + delta
			if newIndex < 0 || newIndex >= len(s.file.Todos) {
				return
			}

			// Swap
			s.file.Todos[i], s.file.Todos[newIndex] = s.file.Todos[newIndex], s.file.Todos[i]

			// Update order
			for j, todo := range s.file.Todos {
				todo.Order = j
			}

			s.Save()
			return
		}
	}
}

func (s *Storage) GetByID(id string) *model.Todo {
	for _, t := range s.file.Todos {
		if t.ID == id {
			return t
		}
	}
	return nil
}

func (s *Storage) Count() int {
	return len(s.file.Todos)
}

func (s *Storage) PendingCount() int {
	count := 0
	for _, t := range s.file.Todos {
		if !t.Done {
			count++
		}
	}
	return count
}

func (s *Storage) Path() string {
	return s.path
}

// Filtering

type Filter struct {
	ShowDone    *bool
	Project     string
	Tag         string
	DueToday    bool
	DueWeek     bool
	Overdue     bool
	SearchTerm  string
	MinPriority *int
}

func (s *Storage) Filter(f Filter) []*model.Todo {
	var result []*model.Todo

	for _, t := range s.file.Todos {
		if f.ShowDone != nil {
			if *f.ShowDone != t.Done {
				continue
			}
		}

		if f.Project != "" && t.Project != f.Project {
			continue
		}

		if f.Tag != "" {
			hasTag := false
			for _, tag := range t.Tags {
				if tag == f.Tag {
					hasTag = true
					break
				}
			}
			if !hasTag {
				continue
			}
		}

		if f.DueToday && (t.Due == nil || !t.Due.IsToday()) {
			continue
		}

		if f.DueWeek && (t.Due == nil || t.Due.DaysUntil() > 7 || t.Due.DaysUntil() < 0) {
			continue
		}

		if f.Overdue && (t.Due == nil || !t.Due.IsOverdue()) {
			continue
		}

		if f.MinPriority != nil && t.Priority < *f.MinPriority {
			continue
		}

		if f.SearchTerm != "" {
			// Case-insensitive search in text, notes, tags, project
			term := f.SearchTerm
			found := containsIgnoreCase(t.Text, term) ||
				containsIgnoreCase(t.Notes, term) ||
				containsIgnoreCase(t.Project, term)
			if !found {
				for _, tag := range t.Tags {
					if containsIgnoreCase(tag, term) {
						found = true
						break
					}
				}
			}
			if !found {
				continue
			}
		}

		result = append(result, t)
	}

	return result
}

func containsIgnoreCase(s, substr string) bool {
	s = strings.ToLower(s)
	substr = strings.ToLower(substr)
	return strings.Contains(s, substr)
}

// Sorting

type SortBy int

const (
	SortByOrder SortBy = iota
	SortByPriority
	SortByDue
	SortByAlpha
	SortByCreated
)

func (s *Storage) Sort(todos []*model.Todo, by SortBy) []*model.Todo {
	result := make([]*model.Todo, len(todos))
	copy(result, todos)

	sort.Slice(result, func(i, j int) bool {
		switch by {
		case SortByPriority:
			return result[i].Priority > result[j].Priority
		case SortByDue:
			if result[i].Due == nil && result[j].Due == nil {
				return result[i].Order < result[j].Order
			}
			if result[i].Due == nil {
				return false
			}
			if result[j].Due == nil {
				return true
			}
			return result[i].Due.Time().Before(result[j].Due.Time())
		case SortByAlpha:
			return result[i].Text < result[j].Text
		case SortByCreated:
			return result[i].Created.Before(result[j].Created)
		default:
			return result[i].Order < result[j].Order
		}
	})

	return result
}
