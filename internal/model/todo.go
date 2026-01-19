package model

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Todo struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	Done      bool      `json:"done"`
	Priority  int       `json:"priority"` // -2 to +3, 0 = normal
	Project   string    `json:"project,omitempty"`
	Tags      []string  `json:"tags,omitempty"`
	Due       *Date     `json:"due,omitempty"`
	Created   time.Time `json:"created"`
	Completed *time.Time `json:"completed,omitempty"`
	Notes     string    `json:"notes,omitempty"`
	Order     int       `json:"order"`
}

// Date is a date-only type (no time component)
type Date struct {
	Year  int
	Month time.Month
	Day   int
}

func (d Date) Time() time.Time {
	return time.Date(d.Year, d.Month, d.Day, 0, 0, 0, 0, time.Local)
}

func (d Date) String() string {
	return d.Time().Format("2006-01-02")
}

func (d Date) MarshalJSON() ([]byte, error) {
	return []byte(`"` + d.String() + `"`), nil
}

func (d *Date) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	d.Year = t.Year()
	d.Month = t.Month()
	d.Day = t.Day()
	return nil
}

func Today() Date {
	now := time.Now()
	return Date{Year: now.Year(), Month: now.Month(), Day: now.Day()}
}

func DateFromTime(t time.Time) Date {
	return Date{Year: t.Year(), Month: t.Month(), Day: t.Day()}
}

func (d Date) IsToday() bool {
	today := Today()
	return d.Year == today.Year && d.Month == today.Month && d.Day == today.Day
}

func (d Date) IsTomorrow() bool {
	tomorrow := DateFromTime(time.Now().AddDate(0, 0, 1))
	return d.Year == tomorrow.Year && d.Month == tomorrow.Month && d.Day == tomorrow.Day
}

func (d Date) IsOverdue() bool {
	return d.Time().Before(time.Now().Truncate(24 * time.Hour))
}

func (d Date) DaysUntil() int {
	today := time.Now().Truncate(24 * time.Hour)
	return int(d.Time().Sub(today).Hours() / 24)
}

func (d Date) RelativeString() string {
	days := d.DaysUntil()
	switch {
	case days < 0:
		if days == -1 {
			return "yesterday"
		}
		return d.Time().Format("Jan 2")
	case days == 0:
		return "today"
	case days == 1:
		return "tomorrow"
	case days < 7:
		return d.Time().Format("Mon")
	case days < 14:
		return "next " + d.Time().Format("Mon")
	default:
		return d.Time().Format("Jan 2")
	}
}

func NewTodo(text string) *Todo {
	todo := &Todo{
		ID:      uuid.New().String()[:8],
		Text:    text,
		Created: time.Now(),
	}
	todo.ParseInlineMetadata()
	return todo
}

// ParseInlineMetadata extracts @project, #tags, due:date, and +priority from text
func (t *Todo) ParseInlineMetadata() {
	text := t.Text

	// Extract project (@project)
	projectRe := regexp.MustCompile(`@(\w+)`)
	if matches := projectRe.FindStringSubmatch(text); len(matches) > 1 {
		t.Project = matches[1]
		text = projectRe.ReplaceAllString(text, "")
	}

	// Extract tags (#tag)
	tagRe := regexp.MustCompile(`#(\w+)`)
	tagMatches := tagRe.FindAllStringSubmatch(text, -1)
	for _, match := range tagMatches {
		if len(match) > 1 {
			t.Tags = append(t.Tags, match[1])
		}
	}
	text = tagRe.ReplaceAllString(text, "")

	// Extract priority (+1, +2, -1, etc)
	priorityRe := regexp.MustCompile(`([+-]\d)`)
	if matches := priorityRe.FindStringSubmatch(text); len(matches) > 1 {
		var p int
		if _, err := parseIntFromString(matches[1], &p); err == nil {
			if p >= -2 && p <= 3 {
				t.Priority = p
			}
		}
		text = priorityRe.ReplaceAllString(text, "")
	}

	// Extract due date (due:today, due:tomorrow, due:mon, due:2024-01-25)
	dueRe := regexp.MustCompile(`due:(\S+)`)
	if matches := dueRe.FindStringSubmatch(text); len(matches) > 1 {
		if date := ParseDueDate(matches[1]); date != nil {
			t.Due = date
		}
		text = dueRe.ReplaceAllString(text, "")
	}

	// Clean up extra spaces
	t.Text = strings.TrimSpace(regexp.MustCompile(`\s+`).ReplaceAllString(text, " "))
}

func parseIntFromString(s string, result *int) (bool, error) {
	var n int
	for i, c := range s {
		if i == 0 && (c == '+' || c == '-') {
			continue
		}
		if c < '0' || c > '9' {
			return false, nil
		}
		n = n*10 + int(c-'0')
	}
	if s[0] == '-' {
		n = -n
	}
	*result = n
	return true, nil
}

func ParseDueDate(s string) *Date {
	s = strings.ToLower(s)
	now := time.Now()

	switch s {
	case "today":
		d := Today()
		return &d
	case "tomorrow", "tom":
		d := DateFromTime(now.AddDate(0, 0, 1))
		return &d
	case "sun", "sunday":
		return nextWeekday(time.Sunday)
	case "mon", "monday":
		return nextWeekday(time.Monday)
	case "tue", "tuesday":
		return nextWeekday(time.Tuesday)
	case "wed", "wednesday":
		return nextWeekday(time.Wednesday)
	case "thu", "thursday":
		return nextWeekday(time.Thursday)
	case "fri", "friday":
		return nextWeekday(time.Friday)
	case "sat", "saturday":
		return nextWeekday(time.Saturday)
	case "week":
		d := DateFromTime(now.AddDate(0, 0, 7))
		return &d
	}

	// Try parsing as date
	formats := []string{"2006-01-02", "01-02", "1/2", "Jan 2", "Jan2"}
	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			// If no year, use current year (or next year if date has passed)
			if t.Year() == 0 {
				t = time.Date(now.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
				if t.Before(now) {
					t = t.AddDate(1, 0, 0)
				}
			}
			d := DateFromTime(t)
			return &d
		}
	}

	return nil
}

func nextWeekday(day time.Weekday) *Date {
	now := time.Now()
	daysUntil := int(day) - int(now.Weekday())
	if daysUntil <= 0 {
		daysUntil += 7
	}
	d := DateFromTime(now.AddDate(0, 0, daysUntil))
	return &d
}

func (t *Todo) Toggle() {
	t.Done = !t.Done
	if t.Done {
		now := time.Now()
		t.Completed = &now
	} else {
		t.Completed = nil
	}
}

func (t *Todo) IncreasePriority() {
	if t.Priority < 3 {
		t.Priority++
	}
}

func (t *Todo) DecreasePriority() {
	if t.Priority > -2 {
		t.Priority--
	}
}
