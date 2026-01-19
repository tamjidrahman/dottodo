package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/tamjidrahman/dottodo/internal/model"
	"github.com/tamjidrahman/dottodo/internal/storage"
)

type CLI struct {
	storage *storage.Storage
}

func New(s *storage.Storage) *CLI {
	return &CLI{storage: s}
}

func (c *CLI) Run(args []string) error {
	if len(args) == 0 {
		return nil // Will run TUI
	}

	command := args[0]
	cmdArgs := args[1:]

	switch command {
	case "add":
		return c.add(cmdArgs)
	case "list", "ls":
		return c.list(cmdArgs)
	case "done":
		return c.done(cmdArgs)
	case "undone":
		return c.undone(cmdArgs)
	case "delete", "del", "rm":
		return c.delete(cmdArgs)
	case "--count", "-c":
		return c.count()
	case "--today":
		return c.today()
	case "--pending":
		return c.pending()
	case "--help", "-h", "help":
		return c.help()
	case "--version", "-v":
		fmt.Println("dottodo v0.1.0")
		return nil
	default:
		// Check if it looks like a todo text (quick add)
		if !strings.HasPrefix(command, "-") {
			text := strings.Join(args, " ")
			return c.quickAdd(text)
		}
		return fmt.Errorf("unknown command: %s", command)
	}
}

func (c *CLI) add(args []string) error {
	if len(args) == 0 {
		// Read from stdin
		var lines []string
		buf := make([]byte, 1024)
		n, _ := os.Stdin.Read(buf)
		if n > 0 {
			lines = strings.Split(strings.TrimSpace(string(buf[:n])), "\n")
		}
		for _, line := range lines {
			if line = strings.TrimSpace(line); line != "" {
				todo := model.NewTodo(line)
				c.storage.Add(todo)
				fmt.Printf("Added: %s\n", todo.Text)
			}
		}
		return nil
	}

	text := strings.Join(args, " ")
	todo := model.NewTodo(text)
	c.storage.Add(todo)
	fmt.Printf("Added: %s\n", todo.Text)
	return nil
}

func (c *CLI) quickAdd(text string) error {
	todo := model.NewTodo(text)
	c.storage.Add(todo)
	fmt.Printf("Added: %s\n", todo.Text)
	return nil
}

func (c *CLI) list(args []string) error {
	todos := c.storage.Todos()

	// Check for JSON output
	jsonOutput := false
	for _, arg := range args {
		if arg == "--json" || arg == "-j" {
			jsonOutput = true
		}
	}

	if jsonOutput {
		data, err := json.MarshalIndent(todos, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}

	// Plain text output
	for i, todo := range todos {
		checkbox := "[ ]"
		if todo.Done {
			checkbox = "[x]"
		}

		// Urgency indicator
		urgency := "   "
		if todo.Urgency > 0 {
			urgency = todo.Urgency.Symbol()
			// Pad to 3 chars
			for len(urgency) < 3 {
				urgency += " "
			}
		}

		line := fmt.Sprintf("%d. %s %s %s", i+1, checkbox, urgency, todo.Text)

		if todo.Project != "" {
			line += " @" + todo.Project
		}
		if len(todo.Tags) > 0 {
			for _, tag := range todo.Tags {
				line += " #" + tag
			}
		}
		if len(todo.Links) > 0 {
			line += fmt.Sprintf(" [%d links]", len(todo.Links))
		}
		if todo.Due != nil {
			line += " (due: " + todo.Due.RelativeString() + ")"
		}
		if todo.Priority != 0 {
			line += fmt.Sprintf(" [%+d]", todo.Priority)
		}

		fmt.Println(line)
	}

	if len(todos) == 0 {
		fmt.Println("No todos. Use 'dottodo add \"your todo\"' to create one.")
	}

	return nil
}

func (c *CLI) done(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: dottodo done <number>")
	}

	index := parseIndex(args[0])
	if index < 0 {
		return fmt.Errorf("invalid index: %s", args[0])
	}

	todos := c.storage.Todos()
	if index >= len(todos) {
		return fmt.Errorf("todo #%d not found", index+1)
	}

	todo := todos[index]
	todo.Done = true
	c.storage.Update(todo)
	fmt.Printf("Completed: %s\n", todo.Text)
	return nil
}

func (c *CLI) undone(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: dottodo undone <number>")
	}

	index := parseIndex(args[0])
	if index < 0 {
		return fmt.Errorf("invalid index: %s", args[0])
	}

	todos := c.storage.Todos()
	if index >= len(todos) {
		return fmt.Errorf("todo #%d not found", index+1)
	}

	todo := todos[index]
	todo.Done = false
	c.storage.Update(todo)
	fmt.Printf("Marked pending: %s\n", todo.Text)
	return nil
}

func (c *CLI) delete(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: dottodo delete <number>")
	}

	index := parseIndex(args[0])
	if index < 0 {
		return fmt.Errorf("invalid index: %s", args[0])
	}

	todos := c.storage.Todos()
	if index >= len(todos) {
		return fmt.Errorf("todo #%d not found", index+1)
	}

	todo := todos[index]
	c.storage.Delete(todo.ID)
	fmt.Printf("Deleted: %s\n", todo.Text)
	return nil
}

func (c *CLI) count() error {
	fmt.Println(c.storage.PendingCount())
	return nil
}

func (c *CLI) today() error {
	filter := storage.Filter{DueToday: true}
	todos := c.storage.Filter(filter)

	for _, todo := range todos {
		checkbox := "[ ]"
		if todo.Done {
			checkbox = "[x]"
		}
		fmt.Printf("%s %s\n", checkbox, todo.Text)
	}

	if len(todos) == 0 {
		fmt.Println("No todos due today.")
	}

	return nil
}

func (c *CLI) pending() error {
	done := false
	filter := storage.Filter{ShowDone: &done}
	todos := c.storage.Filter(filter)

	for _, todo := range todos {
		fmt.Printf("[ ] %s\n", todo.Text)
	}

	if len(todos) == 0 {
		fmt.Println("No pending todos.")
	}

	return nil
}

func (c *CLI) help() error {
	help := `dottodo - Vim-native TUI todo manager

USAGE:
  dottodo                    Open TUI
  dottodo add "Buy milk"     Add a new todo
  dottodo list               List all todos
  dottodo list --json        List as JSON
  dottodo done 1             Mark todo #1 as done
  dottodo undone 1           Mark todo #1 as pending
  dottodo delete 1           Delete todo #1
  dottodo --count            Print pending count
  dottodo --today            Print today's todos
  dottodo --pending          Print pending todos

INLINE METADATA:
  @project                   Assign to project
  #tag                       Add tag
  +1, +2, -1                 Set priority
  due:today                  Set due date
  due:tomorrow, due:mon      Relative dates
  due:2024-01-25             Specific date

EXAMPLES:
  dottodo add "Fix bug @work #urgent +2 due:today"
  dottodo "Quick add without 'add' command"

TUI KEYBINDINGS:
  j/k         Move down/up
  x           Toggle complete
  dd          Delete todo
  o/O         Add todo below/above
  >>          Increase priority
  <<          Decrease priority
  u           Undo
  Ctrl+r      Redo
  /           Search
  :           Command mode
  q           Quit

LEADER KEY (Space):
  Space+a     Quick add
  Space+d     Filter: due today
  Space+w     Filter: due this week
  Space+c     Filter: completed
  Space+p     Filter: pending
  Space+Space Clear filter

For more info: https://github.com/tamjidrahman/dottodo`

	fmt.Println(help)
	return nil
}

func parseIndex(s string) int {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return -1
		}
		n = n*10 + int(c-'0')
	}
	return n - 1 // Convert to 0-based
}
