package model

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Urgency levels
type Urgency int

const (
	UrgencyNone   Urgency = 0
	UrgencyLow    Urgency = 1
	UrgencyMedium Urgency = 2
	UrgencyHigh   Urgency = 3
)

func (u Urgency) String() string {
	switch u {
	case UrgencyLow:
		return "low"
	case UrgencyMedium:
		return "medium"
	case UrgencyHigh:
		return "high"
	default:
		return "none"
	}
}

func (u Urgency) Symbol() string {
	switch u {
	case UrgencyLow:
		return "!"
	case UrgencyMedium:
		return "!!"
	case UrgencyHigh:
		return "!!!"
	default:
		return ""
	}
}

func ParseUrgency(s string) Urgency {
	switch strings.ToLower(s) {
	case "!", "low", "1":
		return UrgencyLow
	case "!!", "medium", "med", "2":
		return UrgencyMedium
	case "!!!", "high", "hi", "3":
		return UrgencyHigh
	default:
		return UrgencyNone
	}
}

// Link represents an external reference
type Link struct {
	Type  string `json:"type"`  // linear, github, url, jira, etc.
	ID    string `json:"id"`    // ticket ID or URL
	Title string `json:"title,omitempty"` // optional display title
}

func (l Link) String() string {
	if l.Title != "" {
		return l.Title
	}
	switch l.Type {
	case "linear":
		return "Linear: " + l.ID
	case "github":
		return "GitHub: " + l.ID
	case "jira":
		return "Jira: " + l.ID
	case "url":
		// Shorten URL for display
		if len(l.ID) > 40 {
			return l.ID[:37] + "..."
		}
		return l.ID
	default:
		return l.Type + ": " + l.ID
	}
}

func (l Link) URL() string {
	switch l.Type {
	case "linear":
		// Linear URLs are like: https://linear.app/team/issue/ABC-123
		return "https://linear.app/issue/" + l.ID
	case "github":
		// GitHub: owner/repo#123
		return "https://github.com/" + strings.Replace(l.ID, "#", "/issues/", 1)
	case "jira":
		// Jira needs base URL, just return ID for now
		return l.ID
	case "url":
		return l.ID
	default:
		return l.ID
	}
}

type Todo struct {
	ID        string     `json:"id"`
	Text      string     `json:"text"`
	Done      bool       `json:"done"`
	Priority  int        `json:"priority"`          // -2 to +3, 0 = normal (importance)
	Urgency   Urgency    `json:"urgency,omitempty"` // 0-3: none, low, medium, high
	Project   string     `json:"project,omitempty"`
	Tags      []string   `json:"tags,omitempty"`
	Links     []Link     `json:"links,omitempty"`   // external references
	Due       *Date      `json:"due,omitempty"`
	Created   time.Time  `json:"created"`
	Completed *time.Time `json:"completed,omitempty"`
	Notes     string     `json:"notes,omitempty"`
	Order     int        `json:"order"`
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

// ParseInlineMetadata extracts @project, #tags, due:date, +priority, urgency, and links from text
func (t *Todo) ParseInlineMetadata() {
	text := t.Text

	// Extract project (@project)
	projectRe := regexp.MustCompile(`@(\w+)`)
	if matches := projectRe.FindStringSubmatch(text); len(matches) > 1 {
		t.Project = matches[1]
		text = projectRe.ReplaceAllString(text, "")
	}

	// Extract tags (#tag) - but not #123 style numbers (those might be issue refs)
	tagRe := regexp.MustCompile(`#([a-zA-Z]\w*)`)
	tagMatches := tagRe.FindAllStringSubmatch(text, -1)
	for _, match := range tagMatches {
		if len(match) > 1 {
			t.Tags = append(t.Tags, match[1])
		}
	}
	text = tagRe.ReplaceAllString(text, "")

	// Extract priority (+1, +2, -1, etc)
	priorityRe := regexp.MustCompile(`([+-]\d)\b`)
	if matches := priorityRe.FindStringSubmatch(text); len(matches) > 1 {
		var p int
		if _, err := parseIntFromString(matches[1], &p); err == nil {
			if p >= -2 && p <= 3 {
				t.Priority = p
			}
		}
		text = priorityRe.ReplaceAllString(text, "")
	}

	// Extract urgency (!, !!, !!!)
	urgencyRe := regexp.MustCompile(`\s(!!!|!!|!)\s|^(!!!|!!|!)[\s]|[\s](!!!|!!|!)$`)
	if matches := urgencyRe.FindStringSubmatch(text); len(matches) > 0 {
		for _, m := range matches[1:] {
			if m != "" {
				t.Urgency = ParseUrgency(m)
				break
			}
		}
		text = urgencyRe.ReplaceAllString(text, " ")
	}

	// Extract due date (due:today, due:tomorrow, due:mon, due:2024-01-25)
	dueRe := regexp.MustCompile(`due:(\S+)`)
	if matches := dueRe.FindStringSubmatch(text); len(matches) > 1 {
		if date := ParseDueDate(matches[1]); date != nil {
			t.Due = date
		}
		text = dueRe.ReplaceAllString(text, "")
	}

	// Extract Linear links (linear:ABC-123)
	linearRe := regexp.MustCompile(`linear:([A-Z]+-\d+)`)
	if matches := linearRe.FindAllStringSubmatch(text, -1); len(matches) > 0 {
		for _, match := range matches {
			t.Links = append(t.Links, Link{Type: "linear", ID: match[1]})
		}
		text = linearRe.ReplaceAllString(text, "")
	}

	// Extract GitHub links (github:owner/repo#123 or gh:owner/repo#123)
	githubRe := regexp.MustCompile(`(?:github|gh):([a-zA-Z0-9_-]+/[a-zA-Z0-9_-]+#\d+)`)
	if matches := githubRe.FindAllStringSubmatch(text, -1); len(matches) > 0 {
		for _, match := range matches {
			t.Links = append(t.Links, Link{Type: "github", ID: match[1]})
		}
		text = githubRe.ReplaceAllString(text, "")
	}

	// Extract Jira links (jira:PROJ-123)
	jiraRe := regexp.MustCompile(`jira:([A-Z]+-\d+)`)
	if matches := jiraRe.FindAllStringSubmatch(text, -1); len(matches) > 0 {
		for _, match := range matches {
			t.Links = append(t.Links, Link{Type: "jira", ID: match[1]})
		}
		text = jiraRe.ReplaceAllString(text, "")
	}

	// Extract URLs (url:https://... or just https://...)
	urlRe := regexp.MustCompile(`(?:url:)?(https?://\S+)`)
	if matches := urlRe.FindAllStringSubmatch(text, -1); len(matches) > 0 {
		for _, match := range matches {
			t.Links = append(t.Links, Link{Type: "url", ID: match[1]})
		}
		text = urlRe.ReplaceAllString(text, "")
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

	// Try parsing as date (various formats)
	// Note: Go uses reference date Mon Jan 2 15:04:05 MST 2006
	formats := []string{
		"01/02/2006", // MM/DD/YYYY
		"01/02/06",   // MM/DD/YY
		"1/2/2006",   // M/D/YYYY
		"1/2/06",     // M/D/YY
		"2006-01-02", // YYYY-MM-DD (ISO)
		"01-02-2006", // MM-DD-YYYY
		"01-02",      // MM-DD
		"1/2",        // M/D
		"Jan 2",      // Mon D
		"Jan2",       // MonD
		"Jan 2, 2006", // Mon D, YYYY
	}
	for _, format := range formats {
		if t, err := time.Parse(format, s); err == nil {
			// If no year (year == 0), use current year (or next year if date has passed)
			if t.Year() == 0 {
				t = time.Date(now.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
				if t.Before(now.Truncate(24 * time.Hour)) {
					t = t.AddDate(1, 0, 0)
				}
			}
			// Handle 2-digit years (06 -> 2006)
			if t.Year() < 100 {
				t = t.AddDate(2000, 0, 0)
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

func (t *Todo) IncreaseUrgency() {
	if t.Urgency < UrgencyHigh {
		t.Urgency++
	}
}

func (t *Todo) DecreaseUrgency() {
	if t.Urgency > UrgencyNone {
		t.Urgency--
	}
}

func (t *Todo) AddLink(linkType, id string) {
	t.Links = append(t.Links, Link{Type: linkType, ID: id})
}

func (t *Todo) RemoveLink(index int) {
	if index >= 0 && index < len(t.Links) {
		t.Links = append(t.Links[:index], t.Links[index+1:]...)
	}
}
